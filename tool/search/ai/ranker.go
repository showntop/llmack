package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/prompt"
	"github.com/showntop/llmack/rag/deepdoc/chunk"
	"github.com/showntop/llmack/tool/search"
)

const (
	// RankPrompt 排序任务描述
	RankPrompt = `
	任务描述：你将会看到一些网页的标题和部分摘要，以及一个用户的查询和查询时间。你的任务是根据这些信息，
	首先理解用户查询的意图，然后评估每个网页的有用性。并按有用性从高到低进行排序，选择有用性最高的{{topk}}个网页。
	网页有用性的评估标准主要包括：1）网页的时效性（比如用户查询2024年的最新新闻，网页返回的是2023年的新闻就是时效性差） 
	2）是否可以根据该网页信息直接、准确地回答用户的查询 3）详细程度
	用户查询：{{query}}	
	用户查询时间：{{time}}
	网页列表：{{abstracts}}
	输出格式：按照网页有用性从高到低，输出网页列表中有用性最高的{{topk}}个网页对应的编号，
	比如网页1、网页2、网页3，并且请确保网页列表的前后顺序不会影响你的排序结果。一定不要输出其他额外内容。
	输出：
	`
)

// Ranker 结构体定义
type Ranker struct {
	chunker     chunk.RecursiveCharacterChunker
	isRank      bool
	modelName   string
	modelKwargs map[string]interface{}
	rankPrompt  string

	// BGE排序相关配置
	headers       map[string]string
	ragCode       string
	namespaceCode string
	bgeModel      string

	isFetchPrompt string
}

// NewRanker 构造函数
func NewRanker() *Ranker {

	// 初始化tokenizer
	// currentPath := filepath.Dir(filepath.Clean("__FILE__"))
	// tokenizerPath := filepath.Join(currentPath, "hunyuan_tokenizer")
	// tokenizer := NewHunYuanTokenizer(
	// 	filepath.Join(tokenizerPath, "vocab.json"),
	// 	filepath.Join(tokenizerPath, "merges.txt"),
	// )

	rm := &Ranker{
		headers: map[string]string{
			// "Authorization": fmt.Sprintf("Bearer %s", modelConfig.TragToken),
			"Content-type": "application/json;charset=utf-8",
		},
	}

	// 初始化LLM
	// rm.llm = buildLLM(rm.modelName, rm.modelKwargs, rm.venusOpenConfig)

	return rm
}

// RankWithBGE BGE模型排序
func (r *Ranker) RankWithBGE(query string, docs []map[string]interface{}, topK int) []int {
	titleAbstract := make([]string, 0)
	for _, data := range docs {
		text := fmt.Sprintf("%s %s", data["title"], data["content"])
		titleAbstract = append(titleAbstract, text)
	}

	reqData := map[string]interface{}{
		"ragCode":       r.ragCode,
		"namespaceCode": r.namespaceCode,
		"model":         r.bgeModel,
		"query":         query,
		"documents":     titleAbstract,
	}

	jsonData, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", "http://api.trag.woa.com/v1/trag/retrieval/rerank",
		strings.NewReader(string(jsonData)))

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// 获取data字段并按relevanceScore排序
	data := result["data"].([]interface{})
	sort.Slice(data, func(i, j int) bool {
		return data[i].(map[string]interface{})["relevanceScore"].(float64) >
			data[j].(map[string]interface{})["relevanceScore"].(float64)
	})

	// 获取topK的index
	topKIndex := make([]int, 0)
	for i := 0; i < topK && i < len(data); i++ {
		topKIndex = append(topKIndex, int(data[i].(map[string]interface{})["index"].(float64)))
	}

	return topKIndex
}

// RankWithLLM LLM模型排序
func (r *Ranker) RankWithLLM(query string, docs []*search.Result, topK int, cutOff int) ([]int, error) {
	titleAbstract := make([]string, 0)
	for i, data := range docs {
		text := fmt.Sprintf("网页%d:%s %s", i, data.Title, data.Content)
		if len(text) > cutOff {
			text = text[:cutOff]
		}
		titleAbstract = append(titleAbstract, text)
	}
	log.InfoContextf(context.Background(), "titleAbstract: %+v", titleAbstract)
	content := strings.Join(titleAbstract, "\n\n")
	currentTime := time.Now().Format("2006-01-02")

	promptVars := map[string]interface{}{
		"topk":      topK,
		"query":     query,
		"time":      currentTime,
		"abstracts": content,
	}

	prompt, err := prompt.Render(RankPrompt, promptVars)
	if err != nil {
		return nil, err
	}
	// 调用LLM生成结果
	messages := append([]llm.Message{llm.SystemPromptMessage(" ")}, llm.UserTextPromptMessage(prompt))
	response, err := llm.NewInstance("openai").Invoke(context.Background(), messages, nil, llm.WithModel("hunyuan-standard"))
	if err != nil {
		log.ErrorContextf(context.Background(), "invoke llm with messages: %+v, error: %+v", messages, err)
		return nil, err
	}
	_ = response
	result := ""
	// result := strings.TrimSpace(response.Generations[0][0].Text)

	// 解析结果获取index
	var topKIndex []int
	parts := strings.Split(result, "、")
	for _, p := range parts {
		if idx, err := strconv.Atoi(strings.TrimPrefix(p, "网页")); err == nil {
			topKIndex = append(topKIndex, idx)
		}
	}

	if len(topKIndex) == 0 {
		topKIndex = make([]int, topK)
		for i := 0; i < topK; i++ {
			topKIndex[i] = i
		}
	}

	return topKIndex, nil
}

// RankAbstractInfo 排序
func (r *Ranker) RankAbstractInfo(query string, abstractInfo []search.Result, topK int, rankModelType string) ([]search.Result, error) {
	abstractInfoList := make([]search.Result, 0)

	// 清理和转换数据
	for _, info := range abstractInfo {
		if info.Content == "" {
			log.Info("abstract info is empty")
			continue
		}

		// 清理标题中的特殊字符
		cleanTitle := strings.NewReplacer(
			"\ue40a", "",
			"\ue40b", "",
		).Replace(info.Title)

		// 清理内容中的特殊字符
		cleanContent := strings.NewReplacer(
			"\ue40a", "",
			"\ue40b", "",
			"\ufeff", "",
			"\u2003\u2003", "",
			"\ue629", "",
		).Replace(info.Content)

		abstractInfoList = append(abstractInfoList, search.Result{
			Title:   cleanTitle,
			Time:    info.Time,
			Snippet: info.Snippet,
			Link:    info.Link,
			Content: cleanContent,
			// ChunkIndex: 0,
		})
	}

	// 确保 topK 不超过可用文档数量
	if topK > len(abstractInfoList) {
		topK = len(abstractInfoList)
	}

	var topKIndex []int
	var err error
	// if rankModelType == "llm" {
	if true {
		log.InfoContextf(context.Background(), "使用llm进行排序，abstractInfoList: %+v", abstractInfoList)
		// topKIndex, err = r.RankWithLLM(query, abstractInfoList, topK, 1000)
	} else {
		log.InfoContextf(context.Background(), "使用bge进行排序")
		// topKIndex, err = r.RankWithBGE(query, abstractInfoList, topK)
	}
	if err != nil {
		return nil, fmt.Errorf("ranking failed: %v", err)
	}

	// 过滤越界的索引
	var validIndices []int
	for _, idx := range topKIndex {
		if idx < len(abstractInfoList) {
			validIndices = append(validIndices, idx)
		}
	}

	// 根据排序后的索引选择文档
	selectedAbstract := make([]search.Result, 0, len(validIndices))
	for _, idx := range validIndices {
		selectedAbstract = append(selectedAbstract, abstractInfoList[idx])
	}

	return selectedAbstract, nil
}

type ChunkData struct {
	URL        string `json:"url"`
	Time       string `json:"time"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Query      string `json:"query"`
	ChunkIndex int    `json:"chunk_index"`
}

// RerankPage 重排序
func (r *Ranker) RerankPage(query string, data []map[string]interface{}, contents []string, topK int, rankModelType string) ([]ChunkData, error) {
	// 更新content
	for i, content := range contents {
		if content != "" {
			data[i]["content"] = content
		}
	}

	// 切块并构建新数据
	var newData []ChunkData
	for _, item := range data {
		texts, err := r.chunker.Chunk(item["content"].(string))
		if err != nil {
			continue
		}

		// 限制每个网页最多取4个chunk
		maxChunks := 4
		if len(texts) < maxChunks {
			maxChunks = len(texts)
		}

		for i := 0; i < maxChunks; i++ {
			chunk := ChunkData{
				URL:        item["url"].(string),
				Time:       item["time"].(string),
				Title:      item["title"].(string),
				Content:    texts[i],
				Query:      item["query"].(string),
				ChunkIndex: i,
			}
			newData = append(newData, chunk)
		}
	}

	// 调整topK值
	if len(newData) < topK {
		topK = len(newData)
	}

	// 重排序
	var topKIndex []int
	var err error

	// if rankModelType == "llm" {
	if true {
		// topKIndex, err = r.RankWithLLM(query, newData, topK, 500)
	} else {
		// topKIndex, err = r.RankWithBGE(query, newData, topK)
	}

	if err != nil {
		return nil, fmt.Errorf("重排序失败: %v", err)
	}

	// 过滤越界的索引
	var validIndices []int
	for _, idx := range topKIndex {
		if idx < len(newData) {
			validIndices = append(validIndices, idx)
		}
	}

	// 构建最终结果
	var rerankOutput []ChunkData
	for _, idx := range validIndices {
		rerankOutput = append(rerankOutput, newData[idx])
	}

	return rerankOutput, nil
}
