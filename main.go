package main

import (
	"log"
	"os"
	"tropequest-api/handlers"
	"tropequest-api/services"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // register the postgres driver
)

func main() {
	// Choose backend: Supabase PostgreSQL when SUPABASE_DB_URL is set,
	// otherwise fall back to the Google Sheets CSV service.
	var svc handlers.BookService

	// Priority 1: Supabase REST API (works over HTTPS, no direct DB needed)
	if sbURL := os.Getenv("SUPABASE_URL"); sbURL != "" {
		sbKey := os.Getenv("SUPABASE_SERVICE_KEY")
		log.Println("SUPABASE_URL found — using SupabaseService (REST API)")
		sbSvc := services.NewSupabaseService(sbURL, sbKey)
		if sbSvc != nil {
			svc = sbSvc
		} else {
			log.Println("SupabaseService failed — trying direct DB")
		}
	}

	// Priority 2: Direct PostgreSQL connection
	if svc == nil {
		if dbURL := os.Getenv("SUPABASE_DB_URL"); dbURL != "" {
			log.Println("SUPABASE_DB_URL found — using DBService (PostgreSQL)")
			dbSvc := services.NewDBService(dbURL)
			if dbSvc != nil {
				svc = dbSvc
			} else {
				log.Println("DBService failed — falling back to SheetsService")
			}
		}
	}

	// Priority 3: Google Sheets fallback
	if svc == nil {
		log.Println("Using SheetsService (Google Sheets CSV)")
		svc = services.NewSheetsService()
	}

	h := handlers.NewHandler(svc)

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
