package serper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool/search"
)

type options struct {
	kind string
	key  string
}

// Option ...
type Option func(*options)

// WithKind ...
func WithKind(k string) Option {
	return func(o *options) {
		o.kind = k
	}
}

// WithKey ...
func WithKey(k string) Option {
	return func(o *options) {
		o.key = k
	}
}

// NewSerper 创建serper
func NewSerper(optxs ...Option) search.Searcher {
	opts := &options{}
	for i := 0; i < len(optxs); i++ {
		optxs[i](opts)
	}
	if opts.kind == "" {
		opts.kind = "search"
	}
	return &Serper{apiKey: os.Getenv("serper_api_key"), kind: opts.kind, options: opts}
}

// Serper 使用serper搜索
type Serper struct {
	options *options
	apiKey  string
	kind    string
}

// Category ....
func (s *Serper) Category() string {
	return "all"
}

// Search 使用serper搜索
func (s *Serper) Search(ctx context.Context, query string) ([]*search.Result, error) {

	url := "https://google.serper.dev/" + s.kind
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{"q":"%s"}`, query))

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("X-API-KEY", os.Getenv("serper_api_key"))
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.InfoContextf(ctx, "serper search result: %s", string(body))
	var result SearchResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	fmt.Printf("resultresultresultresultresultresult: %+v \n", result)
	xxx := make([]*search.Result, 0, result.SearchParameters.Num)
	if result.SearchParameters.Type == "images" {
		for _, v := range result.Images {
			xxx = append(xxx, &search.Result{
				Link:  v.Link,
				Image: v.ImageURL,
				Video: "",
				Title: v.Title,
			})
		}
	} else if result.SearchParameters.Type == "videos" {
		for _, v := range result.Videos {
			xxx = append(xxx, &search.Result{
				Link:  v.Link,
				Video: extractYouTubeID(v.Link),
				Title: v.Title,
			})
		}
	} else {
		for _, v := range result.Organic {
			xxx = append(xxx, &search.Result{
				Time:    v.Time,
				Link:    v.Link,
				Image:   "",
				Video:   "",
				Title:   v.Title,
				Snippet: v.Snippet,
				Content: v.Content,
			})
		}
	}

	return xxx, nil
}

func extractYouTubeID(url string) string {
	regExp := regexp.MustCompile(`^.*(youtu.be\/|v\/|u\/\w\/|embed\/|watch\?v=|\&v=)([^#\&\?]*).*`)
	match := regExp.FindStringSubmatch(url)
	if len(match) > 2 && len(match[2]) == 11 {
		return match[2]
	}
	return ""
}

// SearchResult 搜索结果
type SearchResult struct {
	SearchParameters struct {
		Query  string `json:"q"`
		Type   string `json:"type"`
		Engine string `json:"engine"`
		Num    int    `json:"num"`
	} `json:"searchParameters"`
	Organic []Organic      `json:"organic"`
	Images  []Image        `json:"images"`
	Videos  []Video        `json:"videos"`
	Related []RelatedQuery `json:"relatedSearches"`
}

type RelatedQuery struct {
	Query string `json:"query"`
}

// Video 搜索结果
type Video struct {
	Link     string `json:"link"`
	Title    string `json:"title"`
	VideoURL string `json:"videoUrl"`
	ImageURL string `json:"imageUrl"`
}

// Image 搜索结果
type Image struct {
	Link     string `json:"link"`
	Title    string `json:"title"`
	ImageURL string `json:"imageUrl"`
}

// Organic 搜索结果
type Organic struct {
	Time    string `json:"time"`
	Link    string `json:"link"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	Content string `json:"content"`
}
