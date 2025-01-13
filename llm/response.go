package llm

import (
	"bufio"
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
	message := AssistantPromptMessage("")
	for it := resp.stream.Next(); it != nil; it = resp.stream.Next() {
		message.content.Data += it.Delta.Message.content.Data
	}
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
	return r.Message.content.Data
}

// NewChunk ...
func NewChunk(i int, msg *assistantPromptMessage, useage *Usage) *Chunk {
	return &Chunk{
		Delta: &ChunkDelta{
			Index:        i,
			Message:      msg,
			Usage:        useage,
			FinishReason: "",
			Done:         false,
		},
	}
}

// Chunk ...
type Chunk struct {
	Model             string           `json:"model"`
	Messages          []*PromptMessage `json:"-"`
	SystemFingerprint string           `json:"system_fingerprint"`

	Delta *ChunkDelta `json:"delta"`
}

// ChunkDelta ...
type ChunkDelta struct {
	Index        int                     `json:"index"`
	Message      *assistantPromptMessage `json:"message"`
	Usage        *Usage                  `json:"usage"`
	FinishReason string                  `json:"finish_reason"`
	Done         bool                    `json:"done"`
}

// Usage ...
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Currency         string
	Latency          float64
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
