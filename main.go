package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/AdarshGT33/ednr/events"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s\n", err)
	}

	go event_processor()

	r := gin.Default()

	r.POST("/events", func(c *gin.Context) {
		var event events.Events
		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"status": "received"})
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
			continue
		}

		event_json := result[1]
		var event events.Events
		json.Unmarshal([]byte(event_json), &event)

		//sending events on the basis of severity
	}
}
