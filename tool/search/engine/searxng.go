package engine

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"strings"

// 	"github.com/showntop/llmack/log"
// )

// // NewSearxng create searxng client
// func NewSearxng(url string) Searcher {
// 	return &Searxng{baseUrl: url}
// }

// // Searxng searxng instance
// type Searxng struct {
// 	baseUrl string
// }

// // Search Gets the raw HTML of a searx search result page
// // query : The query to search for.
// func (s *Searxng) Search(ctx context.Context, query string) ([]*Result, error) {

// 	url := s.baseUrl + "/search"

// 	payload := strings.NewReader(fmt.Sprintf(`{"q":"%s","gl":"cn"}`, query))

// 	req, err := http.NewRequest(http.MethodGet, url, payload)

// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux i686; rv:109.0) Gecko/20100101 Firefox/114.0")
// 	req.Header.Add("Content-Type", "application/json")

// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.InfoContextf(ctx, "searxng search result: %s", string(body))

// }
