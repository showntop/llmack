package spider

import (
	"context"
	"sync"

	"github.com/gocolly/colly/v2"

	"github.com/showntop/llmack/log"
)

// Spider 爬虫
type Spider struct {
}

// NewSpider 构造函数
func NewSpider() *Spider {
	return &Spider{}
}

// Crawl 爬取网页内容
func (s *Spider) Crawl(ctx context.Context, urls []string) (map[string]string, error) {
	log.InfoContextf(ctx, "tool spider crawl urls: %+v", urls)
	type item struct {
		url     string
		content string
	}
	// 并发爬取
	wg := sync.WaitGroup{}
	resultChan := make(chan *item, len(urls))
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			content, err := s.crawl(ctx, url)
			if err != nil {
				log.ErrorContextf(ctx, "tool spider crawl url %s failed: %v", url, err)
			}
			resultChan <- &item{
				url:     url,
				content: content,
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	outputs := make(map[string]string, len(urls))
	for item := range resultChan {
		outputs[item.url] = item.content
	}
	log.InfoContextf(ctx, "tool spider crawl result: %+v", len(outputs))
	return outputs, nil
}

func (s *Spider) crawl(ctx context.Context, url string) (string, error) {
	var content string
	collector := colly.NewCollector(
		// colly.AllowedDomains("www.baidu.com"),
		colly.MaxDepth(1),
		colly.Async(true),
	)
	// collector.OnHTML("title", func(e *colly.HTMLElement) {
	// 	log.InfoContextf(ctx, "title: %s", e.Text)
	// })

	collector.OnResponse(func(r *colly.Response) {
		content = string(r.Body)
		log.InfoContextf(ctx, "tool spider crawl url %s response: %d", url, len(content))
	})
	collector.Visit(url)
	collector.Wait()

	return content, nil
}
