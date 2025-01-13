package crawl

import "context"

// Result 爬取结果
type Result struct {
	Link    string
	Title   string
	Content string
}

// Crawler 爬虫接口
type Crawler interface {
	Crawl(ctx context.Context, url string) (*Result, error)
}
