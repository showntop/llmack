package crawl

import (
	"context"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/showntop/llmack/log"
)

// CollyCrawler 使用jina搜索
type CollyCrawler struct {
}

// NewCollyCrawler 创建jina爬虫
func NewCollyCrawler() Crawler {
	return &CollyCrawler{}
}

// CrawlMulti 爬取网页内容
func (s *CollyCrawler) CrawlMulti(ctx context.Context, urls []string) (map[string]*Result, error) {
	log.InfoContextf(ctx, "tool spider crawl urls: %+v", urls)
	// 并发爬取
	wg := sync.WaitGroup{}
	resultChan := make(chan *Result, len(urls))
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			content, err := s.Crawl(ctx, url)
			if err != nil {
				log.ErrorContextf(ctx, "tool spider crawl url %s failed: %v", url, err)
			}
			resultChan <- content
		}(url)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	outputs := make(map[string]*Result, len(urls))
	for item := range resultChan {
		outputs[item.Link] = item
	}
	log.InfoContextf(ctx, "tool spider crawl result: %+v", len(outputs))
	return outputs, nil
}

// Crawl ...
func (s *CollyCrawler) Crawl(ctx context.Context, url string) (*Result, error) {

	var title string
	var content string
	collector := colly.NewCollector(
		// colly.AllowedDomains("www.baidu.com"),
		colly.MaxDepth(1),
		colly.Async(true),
	)
	collector.OnHTML("title", func(e *colly.HTMLElement) {
		// log.InfoContextf(ctx, "title: %s", e.Text)
		title = e.Text
	})

	collector.OnHTML("body", func(e *colly.HTMLElement) {
		content = string(e.Text)
		log.InfoContextf(ctx, "tool spider crawl url %s response: %d", url, len(content))
	})
	collector.Visit(url)
	collector.Wait()

	return &Result{
		Title:   title,
		Content: content,
		Link:    url,
	}, nil
}
