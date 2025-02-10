package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/prompt"
	"github.com/sirupsen/logrus"
)

// Spider 爬虫
type Spider struct{}

// Needed 判断是否需要爬取网页
func (s *Spider) Needed(ctx context.Context, query, abstract string) (bool, error) {
	p := `
		任务描述：你将会看到网页摘要和用户查询。你的任务是根据这些信息，首先理解用户查询的意图，然后评估网页摘要是否可以回答用户的查询。
		评估标准：1）如果网页摘要仅仅是标题或者很简略，则输出0表示需要更详细的内容才能回答。2）如果网页摘要比较具体，可以直接、准确、详细地回答用户查询，则输出1。
		用户查询：{{query}}
		网页摘要：{{abstract}}
		输出格式：仅仅输出0表示需要更详细的网页内容才能回答，输出1表示当前网页摘要可以满足回答用户的查询。一定不要输出其他额外内容。
		输出：
	`
	p, err := prompt.Render(p, map[string]interface{}{
		"query":    query,
		"abstract": abstract,
	})
	if err != nil {
		return false, fmt.Errorf("render prompt failed: %v", err)
	}

	model := llm.NewInstance("openai")
	response, err := model.Invoke(ctx, []llm.Message{
		llm.SystemPromptMessage(" "), llm.UserTextPromptMessage(p),
	}, nil,
		llm.WithModel("hunyuan-standard"),
	)
	if err != nil {
		return false, fmt.Errorf("invoke llm failed: %v", err)
	}
	return response.Result().Message.Content() == "0", nil
}

// Crawl 爬取网页
func (s *Spider) Crawl(ctx context.Context, urls []string) ([]string, error) {
	startTime := time.Now()
	// 并发抓取网页内容
	var wg sync.WaitGroup
	results := make([]string, 0)
	resultsChan := make(chan string, len(urls))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, urlx := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				result, err := s.crawlURL(ctx, url)

				if err != nil {
					logrus.Errorf("spider error occurred: %v", err)
					return
				}

				resultsChan <- result
			}
		}(urlx)
	}

	// 等待所有goroutine完成或超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(resultsChan)
	case <-time.After(time.Second): // 1秒超时
		logrus.Error("spider timed out after 1 seconds")
		cancel()
	}

	// 收集结果
	for result := range resultsChan {
		results = append(results, result)
	}

	logrus.Infof("spider time: %v", time.Since(startTime))
	return results, nil
}

func (s *Spider) crawlURL(ctx context.Context, url string) (string, error) {
	var content string
	c := colly.NewCollector()
	// Find and visit all links
	// c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	// 	e.Request.Visit(e.Attr("href"))
	// })
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Response:", string(r.Body))
		content = string(r.Body)
	})
	c.Visit(url)
	c.Wait()

	// return extractSmart(content, contentFormat, targetURL, title, logTraceInfo,
	// 	date, abstractInfo, onlySpiderInfo)

	// return extractDefault(content, contentFormat, string(respBody), targetURL,
	// 	title, logTraceInfo, abstractInfo, date, onlySpiderInfo)
	return content, nil
}

// NewSpider 创建爬虫
func NewSpider() *Spider {
	return &Spider{}
}
