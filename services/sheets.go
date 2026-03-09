package services

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"tropequest-api/models"
)

const sheetsURL = "https://docs.google.com/spreadsheets/d/e/2PACX-1vSRmtG-NI3-kaQfFVUWktL-ILpwK3iBuJVd2jkJ7kD8rWwJRrWcUHZRGcjtDECcqdZVhd2YBQXgCeAy/pub?gid=0&single=true&output=csv"

type SheetsService struct {
	mu          sync.RWMutex
	books       []models.Book
	lastFetched time.Time
	cacheTTL    time.Duration
}

func NewSheetsService() *SheetsService {
	s := &SheetsService{cacheTTL: 10 * time.Minute}
	// Pre-warm cache on startup
	s.refresh()
	return s
}

func (s *SheetsService) GetBooks() []models.Book {
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

func (s *SheetsService) refresh() {
	books, err := fetchFromSheets()
	if err != nil {
		return
	}
	s.mu.Lock()
	s.books = books
	s.lastFetched = time.Now()
	s.mu.Unlock()
}

func fetchFromSheets() ([]models.Book, error) {
	resp, err := http.Get(sheetsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var books []models.Book
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 2 || strings.TrimSpace(row[0]) == "" {
			continue
		}

		book := models.Book{
			Title:  strings.TrimSpace(row[0]),
			Author: strings.TrimSpace(row[1]),
		}
		if len(row) > 2 && row[2] != "" {
			for _, t := range strings.Split(row[2], "|") {
				if t = strings.TrimSpace(t); t != "" {
					book.Tropes = append(book.Tropes, t)
				}
			}
		}
		if len(row) > 3 {
			book.Description = strings.TrimSpace(row[3])
		}
		if len(row) > 4 {
			book.CoverURL = strings.TrimSpace(row[4])
		}
		if len(row) > 5 {
			book.BookID = strings.TrimSpace(row[5])
		}
		if len(row) > 6 {
			book.BuyURL = strings.TrimSpace(row[6])
		}
		if len(row) > 7 {
			if r, err := strconv.ParseFloat(strings.TrimSpace(row[7]), 64); err == nil {
				book.Rating = r
			}
		}
		if len(row) > 8 {
			if n, err := strconv.Atoi(strings.TrimSpace(row[8])); err == nil {
				book.RatingCount = n
			}
		}
		if len(row) > 9 {
			book.RatingSource = strings.TrimSpace(row[9])
		}
		if len(row) > 10 && row[10] != "" {
			for _, s := range strings.Split(row[10], "|") {
				if s = strings.TrimSpace(s); s != "" {
					book.Subjects = append(book.Subjects, s)
				}
			}
		}
		if len(row) > 11 {
			book.Language = strings.TrimSpace(row[11])
		}
		if len(row) > 12 {
			book.SeriesName = strings.TrimSpace(row[12])
		}
		if len(row) > 13 {
			book.SeriesNumber = strings.TrimSpace(row[13])
		}
		books = append(books, book)
	}
	return books, nil
}
