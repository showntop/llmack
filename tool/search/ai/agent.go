package ai

import (
	"context"
	"time"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/crawl"
	"github.com/showntop/llmack/tool/search"
	"github.com/showntop/llmack/tool/search/multi"
	"github.com/showntop/llmack/tool/search/serper"
)

// Config 配置
type Config struct {
}

// InvokeOptions ...
type InvokeOptions struct {
	TopK               int
	FetchPageContent   string
	ValidStartTime     float64
	TsnEnable          string
	QueryRewriteEnable string
	QueryRewriteMode   string
	RankModelType      string
}

// InvokeOption 搜索选项
type InvokeOption func(*InvokeOptions)

// Agent 搜索智能体
type Agent struct {
	intent        *IntentAgent
	searcher      search.Searcher
	imageSearcher search.Searcher
	videoSearcher search.Searcher
	crawler       crawl.Crawler
	ranker        *Ranker
}

// Stream ...
func (a *Agent) Stream(ctx context.Context, query string, o ...InvokeOption) (chan any, error) {
	events := make(chan any, 10)
	go a.exec(ctx, query, events)
	return events, nil
}

func (a *Agent) exec(ctx context.Context, query string, stream chan any) {
	log.InfoContextf(ctx, "search agent exec with query: %s", query)
	intent := a.intent.Invoke(ctx, query)
	if intent.Decision == "direct" {
		stream <- tool.Event{Name: "answer", Type: "answer", Data: intent.Answer}
		close(stream)
		return
	}
	if intent.Decision == "access" {
		result, err := a.crawler.Crawl(ctx, query)
		if err != nil {
			log.ErrorContextf(ctx, "search failed: %v", err)
			return
		}

		answers := streamx(ctx, answerPrompt, map[string]any{
			"query":          "总结网页" + query + "，网页的内容已经在下面的搜索结果中给出，你无需访问，可以直接使用。",
			"now":            time.Now().Format("2006-01-02 15:04:05"),
			"search_results": result,
		})

		for v := answers.Next(); v != nil; v = answers.Next() {
			stream <- tool.Event{Name: "answer", Type: "answer", Data: v.Delta.Message.Content()}
		}
		close(stream)
		return
	}
	if intent.Decision == "search" {
		a.handleHybridSearch(ctx, intent.Answer, stream)
		close(stream)
	}
}

func (a *Agent) handleHybridSearch(ctx context.Context, query string, stream chan any) {
	log.InfoContextf(ctx, "search agent handle hybrid search with query: %s", query)
	stream <- tool.Event{Name: "search", Type: "search", Data: "检索中..."}
	// search
	result, err := a.searcher.Search(ctx, query)
	if err != nil {
		log.ErrorContextf(ctx, "search failed: %v", err)
		return
	}
	stream <- tool.Event{Name: "sources", Type: "sources", Data: result}

	// answers
	answers := streamx(ctx, answerPrompt, map[string]any{
		"query":            query,
		"now":              time.Now().Format("2006-01-02 15:04:05"),
		"search_results":   result,
		"history_messages": "",
	})
	for v := answers.Next(); v != nil; v = answers.Next() {
		stream <- tool.Event{Name: "answer", Type: "answer", Data: v.Delta.Message.Content()}
	}

	// related
	related := streamRelated(ctx, query, map[string]any{
		"query":          query,
		"now":            time.Now().Format("2006-01-02 15:04:05"),
		"search_results": result,
	})
	for v := related.Next(); v != nil; v = related.Next() {
		stream <- tool.Event{Name: "related", Type: "related", Data: v.Delta.Message.Content()}
	}
	// images
	result, err = a.imageSearcher.Search(ctx, query)
	if err != nil {
		log.ErrorContextf(ctx, "search failed: %v", err)
		return
	}
	stream <- tool.Event{Name: "images", Type: "images", Data: result}

	// videos
	result, err = a.videoSearcher.Search(ctx, query)
	if err != nil {
		log.ErrorContextf(ctx, "search failed: %v", err)
		return
	}
	stream <- tool.Event{Name: "videos", Type: "videos", Data: result}

}

// Invoke 生成搜索结果
func (a *Agent) Invoke(ctx context.Context, query string, o ...InvokeOption) (*search.Result, error) {
	// log.InfoContextf(ctx, "search agent invoke with query: %s", query)
	// opts := InvokeOptions{
	// 	QueryRewriteEnable: "no",
	// 	TopK:               10,
	// }

	// for _, opt := range o {
	// 	opt(&opts)
	// }

	// intent := a.intent.Invoke(ctx, query)
	// if intent.Decision == "direct" {

	// }
	// if intent.Decision == "access" {

	// }
	// if intent.Decision == "access" {

	// }

	// // 并发搜索，混合检索
	// queries := append([]string{query}, query)
	// // perQueryTopK := a.setting.SearchConfig.SearchTopK / len(queries)

	// var wg sync.WaitGroup
	// resultChan := make(chan *search.Result, len(queries))

	// for _, q := range queries {
	// 	wg.Add(1)
	// 	go func(query string) {
	// 		defer wg.Done()
	// 		results, err := a.searcher.Search(ctx, query)
	// 		if err != nil {
	// 			log.ErrorContextf(ctx, "search failed: %v", err)
	// 			return
	// 		}
	// 		_ = results
	// 		resultChan <- results
	// 	}(q)
	// }

	// go func() {
	// 	wg.Wait()
	// 	close(resultChan)
	// }()

	// // 收集结果
	// related := []RelatedQuery{}
	// var abstracts []Result
	// for results := range resultChan {
	// 	abstracts = append(abstracts, results.Organic...)
	// 	related = append(related, results.Related...)
	// }
	// deduplicateByURL := func(entries []Result) []Result {
	// 	seen := make(map[string]bool)
	// 	unique := []Result{}
	// 	for _, entry := range entries {
	// 		if !seen[entry.Link] {
	// 			seen[entry.Link] = true
	// 			entry.Content = strings.ReplaceAll(entry.Content+entry.Title+entry.Snippet, "\n", "")
	// 			unique = append(unique, entry)
	// 		}
	// 	}
	// 	return unique
	// }

	// // 去重
	// abstracts = deduplicateByURL(abstracts)
	// log.InfoContextf(ctx, "search results count: %d", len(abstracts))

	// if len(abstracts) == 0 {
	// 	return nil, nil
	// }
	// // 排序
	// result, _ := a.ranker.RankAbstractInfo(query, abstracts, opts.TopK, opts.RankModelType)
	// log.InfoContextf(ctx, "search ranked results: %v", result)
	// // 处理网页内容获取
	// // return a.handlePageContent(ctx, query, result, opts)
	// return &search.Result{Organic: result, Related: related}, nil
	return nil, nil
}

// func (a *Agent) handlePageContent(ctx context.Context, query string, result []*search.Result, opts InvokeOptions) ([]*search.Result, error) {
// 	log.InfoContextf(ctx, "fetch page content is: %s", opts.FetchPageContent)

// 	if opts.FetchPageContent == "auto" {
// 		selectContent := ""
// 		for i, data := range result {
// 			text := fmt.Sprintf("%d、%s%s", i+1, data.Title, data.Content)
// 			if len(text) > 500 {
// 				text = text[:500]
// 			}
// 			selectContent += text
// 		}

// 		needed, err := a.spider.Needed(context.Background(), query, selectContent)
// 		if err != nil {
// 			// return nil, fmt.Errorf("spider needed failed: %v", err)
// 		}
// 		log.InfoContextf(ctx, "auto fetch is: %v", needed)

// 		if needed {
// 			log.InfoContext(ctx, "start crawl page content")
// 			urls := []string{}
// 			for _, data := range result {
// 				urls = append(urls, data.Link)
// 			}
// 			fetchContent, err := a.spider.Crawl(context.Background(), urls)
// 			if err != nil {
// 				return nil, fmt.Errorf("spider crawl failed: %v", err)
// 			}
// 			_ = fetchContent
// 			return result, nil
// 			// if a.ranker.RerankPage(query, result, fetchContent, opts.TopK, opts.RankModelType)
// 		}
// 		return result, nil
// 	} else if opts.FetchPageContent == "yes" {
// 		log.InfoContext(ctx, "start crawl page content")
// 		urls := []string{}
// 		for _, data := range result {
// 			urls = append(urls, data.Link)
// 		}
// 		fetchContent, err := a.spider.Crawl(context.Background(), urls)
// 		if err != nil {
// 			return nil, fmt.Errorf("spider crawl failed: %v", err)
// 		}
// 		_ = fetchContent
// 		return result, nil

// 		// return a.ranker.RerankPage(query, result, fetchContent, opts.TopK, opts.RankModelType)
// 	}

// 	return result, nil
// }

// NewAgent 创建搜索智能体
func NewAgent(setting *Config) *Agent {
	return &Agent{
		searcher:      multi.NewSearcher(),
		intent:        &IntentAgent{},
		crawler:       crawl.NewCollyCrawler(),
		imageSearcher: serper.NewSerper(serper.WithKind("images")),
		videoSearcher: serper.NewSerper(serper.WithKind("videos")),
	}
}
