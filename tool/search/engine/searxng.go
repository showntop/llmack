package engine

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/showntop/llmack/log"
)

func cleanWhitespace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

// NewSearxng creates a new Searxng client
func NewSearxng(url string) Searcher {
	return &Searxng{baseUrl: strings.TrimSuffix(url, "/")}
}

type Searxng struct {
	baseUrl string
}

func (s *Searxng) Search(ctx context.Context, query string) ([]*Result, error) {
	url := s.baseUrl + "/search"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("q", query)
	// _ = writer.WriteField("format", "json")

	if err := writer.Close(); err != nil {
		return nil, err
	}

	log.InfoContextf(ctx, "searxng search url: %s query:%s", url, query)
	req, err := http.NewRequest(http.MethodPost, url, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux i686; rv:109.0) Gecko/20100101 Firefox/114.0")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html failed: %w", err)
	}
	var results []*Result
	n := 1
	doc.Find(`#results #urls article`).Each(func(i int, resultDiv *goquery.Selection) {
		header := resultDiv.Find("h3, h4").First()
		if header.Length() == 0 {
			return
		}

		link, exists := header.Find("a").Attr("href")
		if !exists {
			return
		}

		title := cleanWhitespace(header.Text())
		snippet := cleanWhitespace(resultDiv.Find("p").First().Text())

		var sources []string
		resultDiv.Find(`.pull-right span, .engines span`).Each(func(j int, s *goquery.Selection) {
			sources = append(sources, cleanWhitespace(s.Text()))
		})

		results = append(results, &Result{
			Link:    link,
			Title:   title,
			Snippet: snippet,
			// Extra:   map[string]interface{}{"sources": sources},
		})
		n++
	})
	log.InfoContextf(ctx, "searxng search completed with %d results", len(results))
	return results, nil
}
