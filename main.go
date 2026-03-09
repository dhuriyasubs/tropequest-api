package main

import (
	"log"
	"tropequest-api/handlers"
	"tropequest-api/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Init sheets service (pre-warms cache on startup)
	sheets := services.NewSheetsService()
	h := handlers.NewHandler(sheets)

	r := gin.Default()

	// CORS for Flutter web
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		api.GET("/books", h.GetBooks)
		api.GET("/tropes", h.GetTropes)
		api.GET("/search", h.Search)
	}

	log.Println("TropeQuest API running on :8080")
	r.Run(":8080")
}
