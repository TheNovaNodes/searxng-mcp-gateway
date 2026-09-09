package searxng

// SearchParams contains parameters for SearXNG query.
type SearchParams struct {
	Query      string
	MaxResults int
	Categories string
	Language   string
	SafeSearch int
	Engines    string
}

// ResultItem represents a single search result.
type ResultItem struct {
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	Content       string   `json:"content"`
	Engine        string   `json:"engine"`
	Engines       []string `json:"engines,omitempty"`
	Score         float64  `json:"score,omitempty"`
	Category      string   `json:"category,omitempty"`
	PublishedDate string   `json:"published_date,omitempty"`
}

// SearchResult represents the normalized output of search_web.
type SearchResult struct {
	Query               string          `json:"query"`
	Count               int             `json:"count"`
	Results             []ResultItem    `json:"results"`
	Answers             []string        `json:"answers,omitempty"`
	Suggestions         []string        `json:"suggestions,omitempty"`
	UnresponsiveEngines [][]interface{} `json:"unresponsive_engines,omitempty"`
	LatencyMs           float64         `json:"latency_ms"`
	Degraded            bool            `json:"degraded,omitempty"`
	Error               string          `json:"error,omitempty"`
}

// HealthInfo represents the diagnostic info for SearXNG.
type HealthInfo struct {
	Status              string          `json:"status"`
	SearXNGURL          string          `json:"searxng_url"`
	Reachable           bool            `json:"reachable"`
	StatusCode          int             `json:"status_code,omitempty"`
	LatencyMs           float64         `json:"latency_ms,omitempty"`
	ResultCount         int             `json:"result_count,omitempty"`
	UnresponsiveEngines [][]interface{} `json:"unresponsive_engines,omitempty"`
	Version             string          `json:"version,omitempty"`
	Error               string          `json:"error,omitempty"`
}
