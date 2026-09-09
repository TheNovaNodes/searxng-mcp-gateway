package echelon

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	reScriptTag = regexp.MustCompile(`(?is)<script.*?>.*?</script>`)
	reStyleTag  = regexp.MustCompile(`(?is)<style.*?>.*?</style>`)
	reHTMLTags  = regexp.MustCompile(`<[^>]+>`)
	reSpaces    = regexp.MustCompile(`\n{3,}`)
)

// NativeScraper provides basic web page fetching without third-party APIs.
type NativeScraper struct {
	httpClient *http.Client
}

// NewNativeScraper creates a native scraper.
func NewNativeScraper(timeout time.Duration) *NativeScraper {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &NativeScraper{
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Scrape performs a standard HTTP GET and extracts clean text.
func (s *NativeScraper) Scrape(ctx context.Context, targetURL string) (*ScrapeResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TheNovaNodes-Agent/1.0; +https://github.com/TheNovaNodes)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain;q=0.9")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &ScrapeResult{
			URL:        targetURL,
			Provider:   "native",
			StatusCode: 0,
			Degraded:   true,
			Error:      err.Error(),
		}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024*2)) // 2MB limit
	if err != nil {
		return &ScrapeResult{
			URL:        targetURL,
			Provider:   "native",
			StatusCode: resp.StatusCode,
			Degraded:   true,
			Error:      err.Error(),
		}, err
	}

	bodyStr := string(bodyBytes)
	clean := reScriptTag.ReplaceAllString(bodyStr, "")
	clean = reStyleTag.ReplaceAllString(clean, "")
	clean = reHTMLTags.ReplaceAllString(clean, " ")
	clean = strings.ReplaceAll(clean, "&nbsp;", " ")
	clean = strings.ReplaceAll(clean, "&amp;", "&")
	clean = strings.ReplaceAll(clean, "&lt;", "<")
	clean = strings.ReplaceAll(clean, "&gt;", ">")
	clean = reSpaces.ReplaceAllString(strings.TrimSpace(clean), "\n\n")

	return &ScrapeResult{
		URL:        targetURL,
		Provider:   "native",
		Markdown:   clean,
		StatusCode: resp.StatusCode,
		Degraded:   resp.StatusCode != http.StatusOK,
	}, nil
}
