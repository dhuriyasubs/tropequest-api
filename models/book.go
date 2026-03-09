package models

type Book struct {
	BookID      string   `json:"book_id,omitempty"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Tropes      []string `json:"tropes"`
	Description string   `json:"description,omitempty"`
	CoverURL    string   `json:"cover_url,omitempty"`
	BuyURL          string `json:"buy_url,omitempty"`
	BuyURLAmazonUS  string `json:"buy_url_amazon_us,omitempty"`
	BuyURLAmazonIN  string `json:"buy_url_amazon_in,omitempty"`
	BuyURLFlipkart  string `json:"buy_url_flipkart,omitempty"`
	Rating       float64  `json:"rating,omitempty"`
	RatingCount  int      `json:"rating_count,omitempty"`
	RatingSource string   `json:"rating_source,omitempty"`
	Subjects     []string `json:"subjects,omitempty"`
	Language     string   `json:"language,omitempty"`
	SeriesName   string   `json:"series_name,omitempty"`
	SeriesNumber string   `json:"series_number,omitempty"`
}

type SearchResult struct {
	Book          Book     `json:"book"`
	MatchCount    int      `json:"match_count"`
	MatchedTropes []string `json:"matched_tropes"`
	MatchScore    float64  `json:"match_score"`
}
