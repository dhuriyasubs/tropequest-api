package handlers

import (
	"net/http"
	"sort"
	"strings"
	"tropequest-api/models"
	"tropequest-api/services"

	"github.com/gin-gonic/gin"
)

// BookService is satisfied by both SheetsService and DBService.
type BookService interface {
	GetBooks() []models.Book
}

type Handler struct {
	svc BookService
}

func NewHandler(svc BookService) *Handler {
	return &Handler{svc: svc}
}

// GET /api/books
func (h *Handler) GetBooks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"books": h.svc.GetBooks(),
		"total": len(h.svc.GetBooks()),
	})
}

// GET /api/tropes
func (h *Handler) GetTropes(c *gin.Context) {
	books := h.svc.GetBooks()
	seen := map[string]bool{}
	var tropes []string
	for _, b := range books {
		for _, t := range b.Tropes {
			if !seen[t] {
				seen[t] = true
				tropes = append(tropes, t)
			}
		}
	}
	sort.Strings(tropes)
	c.JSON(http.StatusOK, gin.H{"tropes": tropes})
}

// GET /api/search?tropes=Enemies+to+Lovers|Slow+Burn&q=colleen+hoover&limit=20
func (h *Handler) Search(c *gin.Context) {
	tropesParam := c.Query("tropes")
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))

	if tropesParam == "" && q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide at least one of: tropes, q"})
		return
	}

	// Parse requested tropes
	var queryTropes []string
	for _, t := range strings.Split(tropesParam, "|") {
		if t = strings.TrimSpace(t); t != "" {
			queryTropes = append(queryTropes, strings.ToLower(t))
		}
	}

	limit := 20
	books := h.svc.GetBooks()
	var results []models.SearchResult

	for _, book := range books {
		// ── Trope matching ──────────────────────────────────────────
		bookTropesLower := make(map[string]string)
		for _, t := range book.Tropes {
			bookTropesLower[strings.ToLower(t)] = t
		}
		var matched []string
		for _, qt := range queryTropes {
			if orig, ok := bookTropesLower[qt]; ok {
				matched = append(matched, orig)
			}
		}

		// ── Text matching with fuzzy spell check ───────────────────
		textScore := 0.0
		if q != "" {
			titleScore  := services.FuzzyTextScore(q, book.Title)
			authorScore := services.FuzzyTextScore(q, book.Author)
			descScore   := services.FuzzyTextScore(q, book.Description)

			textScore = titleScore*3.0 + authorScore*2.0 + descScore*1.0
		}

		// Must match something
		if len(matched) == 0 && textScore == 0 {
			continue
		}

		tropeScore := 0.0
		if len(book.Tropes) > 0 {
			tropeScore = float64(len(matched)) / float64(len(book.Tropes))
		}

		results = append(results, models.SearchResult{
			Book:          book,
			MatchCount:    len(matched),
			MatchedTropes: matched,
			MatchScore:    tropeScore + textScore,
		})
	}

	// Sort by score desc, then match count desc
	sort.Slice(results, func(i, j int) bool {
		if results[i].MatchScore != results[j].MatchScore {
			return results[i].MatchScore > results[j].MatchScore
		}
		return results[i].MatchCount > results[j].MatchCount
	})

	if len(results) > limit {
		results = results[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   len(results),
		"tropes":  queryTropes,
	})
}
