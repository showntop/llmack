package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

// 测试套件结构体
type SearxngTestSuite struct {
	suite.Suite
	server      *httptest.Server
	testCases   []TestCase
	mockHandler http.HandlerFunc // 添加mock处理器字段
	timeout     time.Duration    // 添加超时配置字段
}

// 测试用例结构
type TestCase struct {
	name        string
	query       string
	params      url.Values
	mockHandler http.HandlerFunc
	expectErr   bool
	validate    func(*testing.T, []*Result)
}

// 初始化测试套件
func (s *SearxngTestSuite) SetupSuite() {
	// 初始化模拟服务器
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.mockHandler != nil {
			s.mockHandler(w, r)
			return
		}
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}))

	// 初始化测试用例
	s.testCases = []TestCase{
		{
			name:  "basic search",
			query: "golang",
			validate: func(t *testing.T, results []*Result) {
				require.Greater(t, len(results), 0)
				require.Contains(t, results[0].Title, "Go Programming Language")
			},
		},
		// 添加更多测试用例...
	}
}

// 测试搜索功能
func (s *SearxngTestSuite) TestSearch() {
	// 设置全局HTTP客户端超时
	// originalClient := http.DefaultClient
	http.DefaultClient = &http.Client{Timeout: 5 * time.Second}
	// t.Cleanup(func() { http.DefaultClient = originalClient })

	// se := NewSearxng(s.server.URL)
	se := NewSearxng("http://9.134.217.159:8080")

	for _, tc := range s.testCases {
		s.Run(tc.name, func() {
			results, err := se.Search(context.Background(), tc.query)
			if tc.expectErr {
				require.Error(s.T(), err)
				return
			}
			require.NoError(s.T(), err)
			tc.validate(s.T(), results)
		})
	}
}

// 测试并发安全
func (s *SearxngTestSuite) TestConcurrentSearch() {
	se := NewSearxng(s.server.URL)
	eg := new(errgroup.Group)

	for i := 0; i < 10; i++ {
		eg.Go(func() error {
			_, err := se.Search(context.Background(), "concurrency test")
			return err
		})
	}
	require.NoError(s.T(), eg.Wait())
}

// 测试错误处理
func (s *SearxngTestSuite) TestErrorHandling() {
	testCases := []struct {
		name        string
		handler     http.HandlerFunc
		expectError string
	}{
		{
			name: "timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.Write([]byte("{}"))
			},
			expectError: "context deadline exceeded",
		},
		// 添加更多错误测试用例...
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			se := NewSearxng(s.server.URL)
			_, err := se.Search(context.Background(), "test")
			require.ErrorContains(s.T(), err, tc.expectError)
		})
	}
}

// 基准测试
func BenchmarkSearch(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回模拟响应
	}))
	defer server.Close()

	se := NewSearxng(server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := se.Search(context.Background(), "benchmark")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 运行测试套件
func TestSearxngSuite(t *testing.T) {
	suite.Run(t, new(SearxngTestSuite))
}
