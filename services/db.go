package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"tropequest-api/models"

	"github.com/lib/pq"
)

// DBService queries books from Supabase PostgreSQL and caches results
// with the same 10-minute TTL used by SheetsService.
type DBService struct {
	db          *sql.DB
	mu          sync.RWMutex
	books       []models.Book
	lastFetched time.Time
	cacheTTL    time.Duration
}

// NewDBService opens a connection pool and pre-warms the in-memory cache.
// connStr is a PostgreSQL connection string / URL (e.g. SUPABASE_DB_URL).
// Returns nil and logs an error if the connection cannot be established.
func NewDBService(connStr string) *DBService {
	if connStr == "" {
		log.Println("[DBService] empty connection string — service disabled")
		return nil
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("[DBService] sql.Open error: %v", err)
		return nil
	}

	// Validate the connection immediately
	if err = db.Ping(); err != nil {
		log.Printf("[DBService] db.Ping error: %v — service disabled", err)
		db.Close()
		return nil
	}

	// Connection pool tuning
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	svc := &DBService{
		db:       db,
		cacheTTL: 10 * time.Minute,
	}

	// Pre-warm cache on startup (best-effort; errors are already logged inside)
	svc.refresh()

	log.Println("[DBService] connected and cache warmed")
	return svc
}

// GetBooks returns all books, refreshing the cache if it has expired.
func (s *DBService) GetBooks() []models.Book {
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

// SearchBooks performs a server-side search filtered by tropes and/or a free-
// text query against title and author using PostgreSQL full-text search.
// limit <= 0 defaults to 20.
func (s *DBService) SearchBooks(tropes []string, query string, limit int) ([]models.Book, error) {
	if limit <= 0 {
		limit = 20
	}

	var (
		conditions []string
		args       []interface{}
		argIdx     = 1
	)

	// Trope filter — book must contain ALL requested tropes
	if len(tropes) > 0 {
		conditions = append(conditions, fmt.Sprintf("tropes @> $%d::text[]", argIdx))
		args = append(args, pq.Array(tropes))
		argIdx++
	}

	// Full-text search on title + author
	if q := strings.TrimSpace(query); q != "" {
		conditions = append(conditions,
			fmt.Sprintf(
				"to_tsvector('english', coalesce(title,'') || ' ' || coalesce(author,'')) @@ plainto_tsquery('english', $%d)",
				argIdx,
			),
		)
		args = append(args, q)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	limitClause := fmt.Sprintf("$%d", argIdx)
	args = append(args, limit)

	sqlStr := fmt.Sprintf(`
		SELECT id, title, author, tropes, description, cover_url, book_id,
		       buy_url_amazon_us, buy_url_amazon_in, buy_url_flipkart,
		       rating, rating_count, rating_source,
		       subjects, language, series_name, series_number
		FROM books
		%s
		ORDER BY title
		LIMIT %s
	`, where, limitClause)

	rows, err := s.db.Query(sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("SearchBooks query: %w", err)
	}
	defer rows.Close()

	return scanBooks(rows)
}

// refresh fetches all books from Postgres and updates the in-memory cache.
func (s *DBService) refresh() {
	books, err := s.fetchAll()
	if err != nil {
		log.Printf("[DBService] refresh error: %v", err)
		return
	}
	s.mu.Lock()
	s.books = books
	s.lastFetched = time.Now()
	s.mu.Unlock()
}

// fetchAll queries every row from the books table.
func (s *DBService) fetchAll() ([]models.Book, error) {
	const sqlStr = `
		SELECT id, title, author, tropes, description, cover_url, book_id,
		       buy_url_amazon_us, buy_url_amazon_in, buy_url_flipkart,
		       rating, rating_count, rating_source,
		       subjects, language, series_name, series_number
		FROM books
		ORDER BY title
	`
	rows, err := s.db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("fetchAll query: %w", err)
	}
	defer rows.Close()

	return scanBooks(rows)
}

// scanBooks converts sql.Rows into a []models.Book slice.
func scanBooks(rows *sql.Rows) ([]models.Book, error) {
	var books []models.Book

	for rows.Next() {
		var (
			b           models.Book
			id          int
			description sql.NullString
			coverURL    sql.NullString
			bookID      sql.NullString
			amazonUS    sql.NullString
			amazonIN    sql.NullString
			flipkart    sql.NullString
			rating      sql.NullFloat64
			ratingCount sql.NullInt64
			ratingSource sql.NullString
			language    sql.NullString
			seriesName  sql.NullString
			seriesNum   sql.NullString
		)

		err := rows.Scan(
			&id,
			&b.Title,
			&b.Author,
			pq.Array(&b.Tropes),
			&description,
			&coverURL,
			&bookID,
			&amazonUS,
			&amazonIN,
			&flipkart,
			&rating,
			&ratingCount,
			&ratingSource,
			pq.Array(&b.Subjects),
			&language,
			&seriesName,
			&seriesNum,
		)
		if err != nil {
			return nil, fmt.Errorf("scanBooks row scan: %w", err)
		}

		// Unwrap nullable fields
		b.Description  = description.String
		b.CoverURL     = coverURL.String
		b.BookID       = bookID.String
		b.BuyURLAmazonUS = amazonUS.String
		b.BuyURLAmazonIN = amazonIN.String
		b.BuyURLFlipkart = flipkart.String
		b.Rating       = rating.Float64
		b.RatingCount  = int(ratingCount.Int64)
		b.RatingSource = ratingSource.String
		b.Language     = language.String
		b.SeriesName   = seriesName.String
		b.SeriesNumber = seriesNum.String

		// Ensure nil slices become empty slices (friendlier JSON output)
		if b.Tropes == nil {
			b.Tropes = []string{}
		}
		if b.Subjects == nil {
			b.Subjects = []string{}
		}

		books = append(books, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanBooks rows iteration: %w", err)
	}

	return books, nil
}
