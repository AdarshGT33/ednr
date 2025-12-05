package main

import (
	"net/http"

	"github.com/AdarshGT33/ednr/handler"
	"github.com/gin-gonic/gin"
)

type Event struct {
	UserID    string `json:"used_id"`
	EventType string `json:"event_type"`
	Message   string `json:"message"`
}

func main() {
	r := gin.Default()

	r.POST("/events", func(c *gin.Context) {
		var event Event
		c.BindJSON(&event)
		err := handler.SendEmail("tomaradarsh18@gmail.com", "Checking smtp", event.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "sent"})
	})
	r.Run(":8080")
}
