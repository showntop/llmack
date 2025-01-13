package multi

import (
	"context"
	"sync"

	"github.com/showntop/llmack/tool/search"
	"github.com/showntop/llmack/tool/search/serper"
)

// Result 搜索结果
type Result struct {
	Time    string `json:"time"`
	Link    string `json:"link"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	Content string `json:"content"`
}

// Searcher 多个搜索
type Searcher struct {
	searchers []search.Searcher
}

// NewSearcher ...
func NewSearcher() search.Searcher {
	s := &Searcher{
		searchers: []search.Searcher{
			serper.NewSerper(),
		},
	}
	return s
}

// Category ...
func (s *Searcher) Category() string {
	return "all"
}

// Search ...
func (s *Searcher) Search(ctx context.Context, query string) ([]*search.Result, error) {
	results := make(chan *search.Result, 0)
	wg := sync.WaitGroup{}
	wg.Add(len(s.searchers))
	for _, searcher := range s.searchers {
		go func(s search.Searcher) {
			defer wg.Done()
			xxx, err := s.Search(ctx, query)
			if err != nil {
				panic(err)
			}
			for _, x := range xxx {
				results <- x
			}
		}(searcher)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	zzz := make([]*search.Result, 0)
	for results := range results {
		zzz = append(zzz, results)
	}
	return zzz, nil
}
