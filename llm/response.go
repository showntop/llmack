package llm

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
)

// Response ...
type Response struct {
	stream *Stream
	result *Result
}

// NewStreamResponse ...
func NewStreamResponse() *Response {
	return &Response{
		stream: NewStream(),
	}
}

// Stream ...
func (resp *Response) Stream() *Stream {
	return resp.stream
}

// Result ...
func (resp *Response) Result() *Result {
	if resp.result != nil {
		return resp.result
	}
	resp.result = &Result{}
	// 合并 message
	text := ""
	reasoning := ""
	toolcalls := []*ToolCall{}
	currentTool := &ToolCall{Index: -1}
	for it := resp.stream.Take(); it != nil; it = resp.stream.Take() {
		deltaMessage := it.Choices[0].Delta
		text += deltaMessage.content
		reasoning += deltaMessage.ReasoningContent
		for _, toolcall := range deltaMessage.ToolCalls {
			if currentTool.Index != toolcall.Index {
				currentTool = toolcall
				toolcalls = append(toolcalls, currentTool)
			} else {
				currentTool.ID += toolcall.ID
				currentTool.Type += toolcall.Type
				currentTool.Function.Name += toolcall.Function.Name
				currentTool.Function.Arguments += toolcall.Function.Arguments
			}
		}
	}
	message := NewAssistantMessage(text)
	message.ReasoningContent = reasoning
	message.ToolCalls = toolcalls
	resp.result.Message = message
	return resp.result
}

// Result ...
type Result struct {
	Model string `json:"model"`
	// Messages          []*PromptMessage
	Message           *AssistantMessage `json:"message"`
	Usage             *Usage            `json:"usage"`
	SystemFingerprint string            `json:"system_fingerprint"`
}

// String ...
func (r *Result) String() string {
	return r.Message.String()
}

// NewChunk ...
func NewChunk(i int, msg *AssistantMessage, useage *Usage) *Chunk {
	return &Chunk{
		Choices: []*ChunkChoice{
			{
				Index:        i,
				Delta:        msg,
				FinishReason: "",
			},
		},
	}
}

// buildChunkMessage ...
func buildChunkMessage(line []byte) (*Chunk, error) {
	var mmm struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		Model   string `json:"model"`
		Choices []struct {
			Index int `json:"index"`
			Delta struct {
				ReasoningContent string      `json:"reasoning_content"`
				Content          string      `json:"content"`
				ToolCalls        []*ToolCall `json:"tool_calls"`
			} `json:"delta"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}
	if err := json.Unmarshal(line, &mmm); err != nil {
		return nil, err
	}

	if len(mmm.Choices) < 0 {
		return nil, errors.New("no choices")
	}
	chunk := &Chunk{}
	chunk.Usage = &mmm.Usage
	chunk.ID = mmm.ID
	chunk.CreatedAt = mmm.Created
	chunk.Model = mmm.Model
	chunk.Object = mmm.Object
	// chunk.SystemFingerprint = mmm.SystemFingerprint
	chunk.Choices = make([]*ChunkChoice, 1, 1)
	chunk.Choices[0] = &ChunkChoice{}
	chunk.Choices[0].Index = 0

	if mmm.Choices[0].Delta.ReasoningContent != "" {
		chunk.Choices[0].Delta = NewAssistantReasoningMessage(mmm.Choices[0].Delta.ReasoningContent)
	} else if len(mmm.Choices[0].Delta.ToolCalls) > 0 {
		chunk.Choices[0].Delta = NewAssistantMessage(mmm.Choices[0].Delta.Content)
		chunk.Choices[0].Delta.ToolCalls = mmm.Choices[0].Delta.ToolCalls
	} else {
		chunk.Choices[0].Delta = NewAssistantMessage(mmm.Choices[0].Delta.Content)
	}
	chunk.Choices[0].FinishReason = mmm.Choices[0].FinishReason

	choices := []*ChunkChoice{}
	for i := 0; i < len(mmm.Choices); i++ {
		choices = append(choices, &ChunkChoice{
			Index:        i,
			Delta:        chunk.Choices[0].Delta,
			FinishReason: mmm.Choices[i].FinishReason,
		})
	}
	chunk.Choices = choices
	return chunk, nil
}

// Chunk ...
type Chunk struct {
	ID                string         `json:"id"`
	CreatedAt         int64          `json:"created_at"`
	Model             string         `json:"model"`
	Object            string         `json:"object"`
	SystemFingerprint string         `json:"system_fingerprint"`
	Choices           []*ChunkChoice `json:"choices"`
	Usage             *Usage         `json:"usage"`
	// Delta             *ChunkDelta   `json:"-"`
}

// ChunkChoice ...
type ChunkChoice struct {
	Index        int               `json:"index"`
	FinishReason string            `json:"finish_reason"`
	Logprobs     any               `json:"logprobs"`
	Delta        *AssistantMessage `json:"delta"`
}

// Usage ...
type Usage struct {
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	Currency         string `json:"currency"`
	Latency          int    `json:"latency"`
}

// Stream ...
type Stream struct {
	q chan *Chunk
}

// NewStream ...
func NewStream() *Stream {
	return &Stream{
		q: make(chan *Chunk, 10),
	}
}

// Read ...
func (s *Stream) Read(body io.ReadCloser, fn func(chunk []byte) (*Chunk, bool)) {
	go func() {
		defer body.Close()
		defer s.Close()

		scanner := bufio.NewScanner(body)
		for scanner.Scan() {
			item, finished := fn(scanner.Bytes())
			if finished {
				break
			}
			if item == nil {
				continue
			}
			s.Push(item)
		}
		// s.push error
		// return scanner.Err()
	}()
}

// Take ...
func (s *Stream) Take() *Chunk {
	select {
	case it := <-s.q:
		return it
		// default:
		// 	return nil
	}
}

// Next ...
func (s *Stream) Next() <-chan *Chunk {
	return s.q
}

// Close ...
func (s *Stream) Close() {
	close(s.q)
}

// Push ...
func (s *Stream) Push(it *Chunk) {
	s.q <- it
}
