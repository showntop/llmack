package crawl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func init() {
	Register(Jina, NewJinaCrawler())
}

// JinaCrawler 使用jina搜索
type JinaCrawler struct {
}

// NewJinaCrawler 创建jina爬虫
func NewJinaCrawler() Crawler {
	return &JinaCrawler{}
}

// Crawl 爬取
func (c *JinaCrawler) Crawl(ctx context.Context, link string) (*Result, error) {
	accessURL := fmt.Sprintf("https://r.jina.ai/%s", link)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, accessURL, strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Code   int    `json:"code"`
		Status int    `json:"status"`
		Data   Result `json:"data"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}
