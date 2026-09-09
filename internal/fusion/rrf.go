package fusion

import (
	"math"
	"net/url"
	"sort"
	"strings"
)

// Item represents an item to be fused.
type Item struct {
	Title   string   `json:"title"`
	URL     string   `json:"url"`
	Content string   `json:"content"`
	Engine  string   `json:"engine"`
	Engines []string `json:"engines,omitempty"`
	Score   float64  `json:"score,omitempty"`
}

// NormalizeURL cleans up a URL by stripping fragments and trailing slashes.
func NormalizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	// Remove fragment
	if idx := strings.Index(rawURL, "#"); idx != -1 {
		rawURL = rawURL[:idx]
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return strings.TrimRight(rawURL, "/")
	}

	parsed.Fragment = ""
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Path != "/" {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
	}

	res := parsed.String()
	return strings.TrimRight(res, "/")
}

// ReciprocalRankFusion combines multiple ranked lists of Items using RRF.
func ReciprocalRankFusion(lists [][]Item, k int) []Item {
	if k <= 0 {
		k = 60
	}

	scores := make(map[string]float64)
	itemsByURL := make(map[string]Item)

	for _, list := range lists {
		seenInList := make(map[string]bool)
		for rank, item := range list {
			cleanURL := NormalizeURL(item.URL)
			if cleanURL == "" {
				continue
			}

			if seenInList[cleanURL] {
				continue
			}
			seenInList[cleanURL] = true

			eng := item.Engine

			existing, exists := itemsByURL[cleanURL]
			if !exists {
				scores[cleanURL] = 0.0
				newItem := item
				newItem.URL = cleanURL
				if eng != "" {
					newItem.Engines = []string{eng}
				} else {
					newItem.Engines = []string{}
				}
				itemsByURL[cleanURL] = newItem
			} else {
				if eng != "" {
					alreadyHas := false
					for _, e := range existing.Engines {
						if e == eng {
							alreadyHas = true
							break
						}
					}
					if !alreadyHas {
						existing.Engines = append(existing.Engines, eng)
					}
				}
				if len(item.Content) > len(existing.Content) {
					existing.Content = item.Content
				}
				itemsByURL[cleanURL] = existing
			}

			scores[cleanURL] += 1.0 / float64(k+rank+1)
		}
	}

	urls := make([]string, 0, len(scores))
	for u := range scores {
		urls = append(urls, u)
	}

	sort.Slice(urls, func(i, j int) bool {
		return scores[urls[i]] > scores[urls[j]]
	})

	var result []Item
	for _, u := range urls {
		item := itemsByURL[u]
		item.Score = math.Round(scores[u]*10000.0) / 10000.0
		result = append(result, item)
	}

	return result
}
