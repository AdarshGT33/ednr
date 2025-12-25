package main

import (
	"log"
	"net/http"

	"github.com/AdarshGT33/ednr/events"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s\n", err)
	}
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
