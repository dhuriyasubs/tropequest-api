package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"tropequest-api/models"
)

// SupabaseService fetches books from Supabase via the PostgREST HTTP API.
// No direct DB connection required — works anywhere HTTPS is available.
type SupabaseService struct {
	baseURL    string
	serviceKey string
	client     *http.Client
	mu         sync.RWMutex
	books      []models.Book
	lastFetched time.Time
	cacheTTL   time.Duration
}

func NewSupabaseService(supabaseURL, serviceKey string) *SupabaseService {
	if supabaseURL == "" || serviceKey == "" {
		log.Println("[SupabaseService] missing SUPABASE_URL or SUPABASE_SERVICE_KEY — disabled")
		return nil
	}
	svc := &SupabaseService{
		baseURL:    strings.TrimRight(supabaseURL, "/"),
		serviceKey: serviceKey,
		client:     &http.Client{Timeout: 15 * time.Second},
		cacheTTL:   10 * time.Minute,
	}
	svc.refresh()
	return svc
}

func (s *SupabaseService) GetBooks() []models.Book {
	s.mu.RLock()
	expired := time.Since(s.lastFetched) > s.cacheTTL
	s.mu.RUnlock()

	if expired {
		s.refresh()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.books
}

func (s *SupabaseService) refresh() {
	books, err := s.fetchAll()
	if err != nil {
		log.Printf("[SupabaseService] refresh error: %v", err)
		return
	}
	s.mu.Lock()
	s.books = books
	s.lastFetched = time.Now()
	s.mu.Unlock()
	log.Printf("[SupabaseService] cached %d books", len(books))
}

// supabaseBook mirrors the JSON returned by Supabase PostgREST.
type supabaseBook struct {
	Title          string   `json:"title"`
	Author         string   `json:"author"`
	Tropes         []string `json:"tropes"`
	Description    *string  `json:"description"`
	CoverURL       *string  `json:"cover_url"`
	BookID         *string  `json:"book_id"`
	BuyURLAmazonUS *string  `json:"buy_url_amazon_us"`
	BuyURLAmazonIN *string  `json:"buy_url_amazon_in"`
	BuyURLFlipkart *string  `json:"buy_url_flipkart"`
	Rating         *float64 `json:"rating"`
	RatingCount    *int     `json:"rating_count"`
	RatingSource   *string  `json:"rating_source"`
	Subjects       []string `json:"subjects"`
	Language       *string  `json:"language"`
	SeriesName     *string  `json:"series_name"`
	SeriesNumber   *string  `json:"series_number"`
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *SupabaseService) fetchAll() ([]models.Book, error) {
	// Fetch up to 1000 books; increase if needed
	url := fmt.Sprintf("%s/rest/v1/books?select=*&order=title.asc&limit=1000", s.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var rows []supabaseBook
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	books := make([]models.Book, 0, len(rows))
	for _, r := range rows {
		tropes := r.Tropes
		if tropes == nil {
			tropes = []string{}
		}
		subjects := r.Subjects
		if subjects == nil {
			subjects = []string{}
		}
		rating := 0.0
		if r.Rating != nil {
			rating = *r.Rating
		}
		ratingCount := 0
		if r.RatingCount != nil {
			ratingCount = *r.RatingCount
		}
		books = append(books, models.Book{
			Title:          r.Title,
			Author:         r.Author,
			Tropes:         tropes,
			Description:    derefStr(r.Description),
			CoverURL:       derefStr(r.CoverURL),
			BookID:         derefStr(r.BookID),
			BuyURLAmazonUS: derefStr(r.BuyURLAmazonUS),
			BuyURLAmazonIN: derefStr(r.BuyURLAmazonIN),
			BuyURLFlipkart: derefStr(r.BuyURLFlipkart),
			Rating:         rating,
			RatingCount:    ratingCount,
			RatingSource:   derefStr(r.RatingSource),
			Subjects:       subjects,
			Language:       derefStr(r.Language),
			SeriesName:     derefStr(r.SeriesName),
			SeriesNumber:   derefStr(r.SeriesNumber),
		})
	}
	return books, nil
}
