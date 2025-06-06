package search

import (
	"context"
	"net/http"
	"sync"

	html2md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/showntop/llmack/log"
)

// CrawlWebpage 并发抓取网页内容
func CrawlWebpage(ctx context.Context, urls []string) ([]string, error) {

	var wg sync.WaitGroup
	resultsChan := make(chan string, len(urls))

	for _, urlx := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				result, err := crawlURL(ctx, url)
				if err != nil {
					resultsChan <- "null"
					return
				}

				resultsChan <- result
			}
		}(urlx)
	}

	// 等待所有任务完成或超时
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 收集结果
	results := make([]string, 0)
	for result := range resultsChan {
		results = append(results, result)
	}

	return results, nil
}

func crawlURL(ctx context.Context, url string) (string, error) {
	log.InfoContextf(ctx, "serper crawl detail about %s", url)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	res, err := http.DefaultClient.Do(req) // TODO timeout configure
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", err
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	// fmt.Println(doc.Html())
	// ttt := doc.Find("a").Remove().End().Text()
	// fmt.Println(ttt)
	// panic("test")

	// doc.Find("img").Each(func(i int, s *goquery.Selection) {
	// 	s.Remove()
	// })
	html, _ := doc.Find("body").Html()
	// Find the review items
	content, err := html2md.ConvertString(html)
	// fmt.Println(content)
	// panic("test")
	// 提前
	return content, err
}

func crawlURL2(ctx context.Context, url string) (string, error) {
	// select
	var content string
	c := colly.NewCollector()
	c.OnResponse(func(r *colly.Response) {
		// fmt.Println("Response:", string(r.Body))
		content, _ = html2md.ConvertString(string(r.Body))
		// fmt.Println("content:", content)
		// 去除所有图片
		// 去除script
		// fmt.Println(url)
	})
	c.Visit(url)
	c.Wait()

	return content, nil
}

// 提取正文
// func extractSmart(content string, contentFormat string, targetURL string, title string, logTraceInfo string,
// 	date string, abstractInfo string, onlySpiderInfo bool) (string, error) {
// 	// 提取正文
// 	if contentFormat == "html" {
// 		content, err := html2md.ConvertString(content)
// 		if err != nil {
// 			return "", err
// 		}
// 		return content, nil
// 	} else if contentFormat == "markdown" {
// 		return content, nil
// 	} else {
// 		return "", errors.New("不支持的contentFormat")
// 	}
// }
