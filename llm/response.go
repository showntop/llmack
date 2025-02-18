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

// MakeResult ...
func (resp *Response) MakeResult() *Response {
	resp.result = &Result{}
	return resp
}

// MakeStream ...
func (resp *Response) MakeStream() *Response {
	resp.stream = NewStream()
	return resp
}

// Stream ...
func (resp *Response) Stream() *Stream {
	return resp.stream
}

// Result ...
func (resp *Response) Result() *Result {
	resp.result = &Result{}
	// 合并 message
	text := ""
	for it := resp.stream.Next(); it != nil; it = resp.stream.Next() {
		text += it.Delta.Message.content
	}
	message := AssistantPromptMessage(text)
	resp.result.Message = message
	return resp.result
}

// Result ...
type Result struct {
	Model string `json:"model"`
	// Messages          []*PromptMessage
	Message           *assistantPromptMessage `json:"message"`
	Usage             *Usage                  `json:"usage"`
	SystemFingerprint string                  `json:"system_fingerprint"`
}

// String ...
func (r *Result) String() string {
	return r.Message.String()
}

// NewChunk ...
func NewChunk(i int, msg *assistantPromptMessage, useage *Usage) *Chunk {
	return &Chunk{
		Delta: &ChunkDelta{
			Message:      msg,
			FinishReason: "",
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
	chunk.Delta = &ChunkDelta{}
	chunk.Delta.Index = 0

	if mmm.Choices[0].Delta.ReasoningContent != "" {
		chunk.Delta.Message = AssistantReasoningMessage(mmm.Choices[0].Delta.ReasoningContent)
	} else if len(mmm.Choices[0].Delta.ToolCalls) > 0 {
		chunk.Delta.Message = AssistantPromptMessage(mmm.Choices[0].Delta.Content)
		chunk.Delta.Message.ToolCalls = mmm.Choices[0].Delta.ToolCalls
	} else {
		chunk.Delta.Message = AssistantPromptMessage(mmm.Choices[0].Delta.Content)
	}
	chunk.Delta.FinishReason = mmm.Choices[0].FinishReason

	// chunk.Choices = []*ChunkDelta{
	// 	{Index: 0, Message: chunk.Delta.Message, FinishReason: mmm.Choices[0].FinishReason},
	// }
	return chunk, nil
}

// Chunk ...
type Chunk struct {
	ID                string        `json:"id"`
	CreatedAt         int64         `json:"created_at"`
	Model             string        `json:"model"`
	Object            string        `json:"object"`
	SystemFingerprint string        `json:"system_fingerprint"`
	Choices           []*ChunkDelta `json:"choices"`
	Usage             *Usage        `json:"usage"`
	Delta             *ChunkDelta   `json:"delta"`
}

// ChunkDelta ...
type ChunkDelta struct {
	Index        int                     `json:"index"`
	FinishReason string                  `json:"finish_reason"`
	Logprobs     any                     `json:"logprobs"`
	Message      *assistantPromptMessage `json:"message"`
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

// Next ...
func (s *Stream) Next() *Chunk {
	select {
	case it := <-s.q:
		return it
		// default:
		// 	return nil
	}
}

// Close ...
func (s *Stream) Close() {
	close(s.q)
}

// Push ...
func (s *Stream) Push(it *Chunk) {
	s.q <- it
}
