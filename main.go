package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/AdarshGT33/ednr/adapters"
	"github.com/AdarshGT33/ednr/events"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client
var adapter map[string]adapters.NotificationAdapter

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s\n", err)
	}

	adapter = map[string]adapters.NotificationAdapter{
		"email": adapters.NewEmailAdapter(),
		"sms":   adapters.NewSMSAdapter(),
	}

	go event_processor()

	r := gin.Default()

	r.POST("/events", func(c *gin.Context) {
		var event events.Events
		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if _, exists := adapter[event.Channel]; !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("unknown channel: %s", event.Channel),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "queued",
			"channel": event.Channel,
		})
	})

	r.Run(":8080")
}

// background worker that processes events
func event_processor() {
	for {
		//infinite loop, worker never stops
		//BRPop: Blocking Right Pop
		//Blocking means wait here until something appears in the queue
		//Right means pop from the right end of the queue
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
		adap_er, exists := adapter[event.Channel]
		if !exists {
			fmt.Printf("Unknown channel: %s\n", event.Channel)
		}

		if err := adap_er.Send(event.Recipient, event.Message); err != nil {
			fmt.Printf("Failed to send: %v\n", err)
			//impliment DLQ here
		} else {
			fmt.Printf("Event successfully send: %s\n", event.Channel)
		}
	}
}
