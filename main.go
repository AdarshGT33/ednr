package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AdarshGT33/ednr/adapters"
	"github.com/AdarshGT33/ednr/events"
	"github.com/AdarshGT33/ednr/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client
var adapter map[string]adapters.NotificationAdapter

const RETRY_QUEUE_KEY = "retry_queue"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s\n", err)
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	adapter = map[string]adapters.NotificationAdapter{
		"email": adapters.NewEmailAdapter(),
		"sms":   adapters.NewSMSAdapter(),
	}

	go event_processor()
	go retry_processor()

	r := gin.Default()

	r.POST("/events", func(c *gin.Context) {
		var event events.Events
		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		channel := events.DetermineChannel(event)
		if _, exists := adapter[channel]; !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("unknown channel: %s", channel),
			})
			return
		}

		if event.MaxAttempts == 0 {
			event.MaxAttempts = 3
		}
		event.AttemptCount = 0
		event.CreatedAt = time.Now()

		eventJSON, _ := json.Marshal(event)
		rdb.LPush(ctx, "event_queue", eventJSON)

		c.JSON(http.StatusOK, gin.H{
			"status":       "queued",
			"channel":      channel,
			"max_attempts": event.MaxAttempts,
		})
	})
	r.GET("/dlq/stats", getDLQStatsHandler)
	r.GET("/dlq/events", listDLQEventsHandler)

	r.Run(":8080")
}

// background worker that processes events
func event_processor() {
	fmt.Println("😋 Event processor started...")

	for {
		//infinite loop, worker never stops
		//BRPop: Blocking Right Pop
		//Blocking means wait here until something appears in the queue
		//Right means pop from the right end of the queue
		fmt.Println("have entered the for loop")
		result, err := rdb.BRPop(ctx, 0, "event_queue").Result()
		if err != nil {
			fmt.Printf("Error popping from queue: %v\n", err)
			continue
		}

		event_json := result[1]
		var event events.Events
		if err := json.Unmarshal([]byte(event_json), &event); err != nil {
			fmt.Printf("Error parsing event: %v\n", err)
			continue
		}

		//sending events on the basis of severity
		channel := events.DetermineChannel(event)
		adap_er, exists := adapter[channel]
		if !exists {
			fmt.Printf("Unknown channel: %s\n", channel)
		}

		event.AttemptCount++
		event.LastAttemptAt = time.Now()

		if err := adap_er.Send(event.Recipient, event.Message); err != nil {
			fmt.Printf("Failed to send: %v", err)
			//impliment DLQ here
			if event.ShouldRetry() {
				// Schedule for retry
				if retryErr := utils.ScheduleRetry(ctx, rdb, event, err); retryErr != nil {
					fmt.Printf("❌ Failed to schedule retry: %v\n", retryErr)
				}
			} else {
				// Max attempts reached, move to DLQ
				if dlqErr := utils.MoveToDLQ(ctx, rdb, event, err); dlqErr != nil {
					fmt.Printf("❌ Failed to move to DLQ: %v\n", dlqErr)
				}
			}
		} else {
			fmt.Printf("Event successfully send: %s\n", channel)
		}
	}
}

// background worker that retries events
func retry_processor() {
	fmt.Println("🔄 Retry processor started...")

	for {
		// Pop event from retry queue (non-blocking check)
		result, err := rdb.BRPop(ctx, 1*time.Second, RETRY_QUEUE_KEY).Result()
		if err != nil {
			// Queue is empty, that's fine
			continue
		}

		eventJSON := result[1]
		var event events.Events
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			fmt.Printf("❌ Error parsing retry event: %v\n", err)
			continue
		}

		// Calculate how long to wait
		backoff := event.GetBackOffDuration()
		timeSinceLastAttempt := time.Since(event.LastAttemptAt)

		if timeSinceLastAttempt < backoff {
			// Not time yet, put back in queue
			rdb.LPush(ctx, RETRY_QUEUE_KEY, eventJSON)
			time.Sleep(100 * time.Millisecond) // Small delay to avoid hot loop
			continue
		}

		// Time to retry! Push back to main queue
		rdb.LPush(ctx, "event_queue", eventJSON)
		fmt.Printf("⏰ [RETRY] Retry time reached for event %s (attempt #%d)\n",
			event.Event_Type, event.AttemptCount+1)
	}
}

// get DLQ stats
func getDLQStatsHandler(c *gin.Context) {
	count, err := utils.GetDLQStats(ctx, rdb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dlq_size": count,
		"status":   "ok",
	})
}

// List DLQ events
func listDLQEventsHandler(c *gin.Context) {
	events, err := utils.ListDLQEvents(ctx, rdb, 100) // Get up to 100
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}
