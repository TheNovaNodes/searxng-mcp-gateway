package echelon

// EchelonItem represents an external search result item.
type EchelonItem struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Engine  string  `json:"engine"`
	Score   float64 `json:"score,omitempty"`
}

// ScrapeResult represents the normalized result of a web page scrape.
type ScrapeResult struct {
	URL        string `json:"url"`
	Provider   string `json:"provider"`
	Markdown   string `json:"markdown,omitempty"`
	Title      string `json:"title,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	Degraded   bool   `json:"degraded,omitempty"`
	Error      string `json:"error,omitempty"`
}
