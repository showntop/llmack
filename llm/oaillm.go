package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

// base llm by openai input and output
type OAILLM struct {
	baseURL string
	apiKey  string

	client *http.Client
}

func NewOAILLM(url, apiKey string) *OAILLM {
	return &OAILLM{
		baseURL: url,
		apiKey:  apiKey,
		client:  http.DefaultClient,
	}
}

func (o *OAILLM) Invoke(ctx context.Context, messages []Message, optFuncs ...InvokeOption) (*Response, error) {
	options := &InvokeOptions{}
	for i := 0; i < len(optFuncs); i++ {
		optFuncs[i](options)
	}
	// validate
	if options.Model == "" {
		return nil, errors.New("model is required")
	}

	// chat completions
	body, err := o.ChatCompletions(ctx, o.buildRequest(messages, options))
	if err != nil {
		return nil, err
	}

	if options.Stream {
		return o.handleStreamResponse(body)
	}
	return o.handleStreamResponse(body)
}

// ChatCompletions ...
func (o *OAILLM) ChatCompletions(ctx context.Context, req *ChatCompletionRequest) (io.ReadCloser, error) {

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	log.Println("OAILLM ChatCompletions request payload ", string(payload)) // for debug
	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, errors.New(resp.Status + ": " + string(raw))
	}

	return resp.Body, nil
}

// buildRequest ...
func (o *OAILLM) buildRequest(messages []Message, options *InvokeOptions) *ChatCompletionRequest {
	request := &ChatCompletionRequest{}
	request.InvokeOptions = options
	// messages
	for _, m := range messages {
		msg := &ChatCompletionMessage{
			Role:       string(m.Role()),
			ToolCallID: m.ToolID(),
		}
		msg.ToolCalls = m.GetToolCalls()
		msg.Content = m.Content()
		msg.MultipartContent = m.MultipartContent()
		request.Messages = append(request.Messages, msg)
	}
	if len(options.Tools) <= 0 {
		return request
	}
	request.ToolChoice = "required"
	request.Tools = options.Tools
	return request
}

// handleStreamResponse ...
func (o *OAILLM) handleStreamResponse(body io.ReadCloser) (*Response, error) {
	response := NewStreamResponse()

	process := func() {
		defer body.Close()
		defer response.Stream().Close()

		reader := bufio.NewReader(body)
		for {
			line, err := readLine(reader)
			if err != nil {
				if err == io.EOF {
					break
				}
				break
			}

			chunk, err := buildChunkMessage(line) // TODO Unmarshal line
			if err != nil {
				continue
			}
			response.Stream().Push(chunk)
		}
	}

	go process()

	return response, nil
}

func readLine(reader *bufio.Reader) ([]byte, error) {

	var (
		headerData  = []byte("data: ")
		errorPrefix = []byte(`data: {"error":`)
	)

	var (
		emptyMessagesCount uint
		hasErrorPrefix     bool
	)

READ:
	rawLine, err := reader.ReadBytes('\n')
	if err != nil { // TODO error handle
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}
	noSpaceLine := bytes.TrimSpace(rawLine)
	if bytes.HasPrefix(noSpaceLine, errorPrefix) {
		hasErrorPrefix = true
	}

	if !bytes.HasPrefix(noSpaceLine, headerData) || hasErrorPrefix {
		if hasErrorPrefix {
			noSpaceLine = bytes.TrimPrefix(noSpaceLine, headerData)
		}
		// writeErr := stream.errAccumulator.Write(noSpaceLine)
		// if writeErr != nil {
		// 	return nil, writeErr
		// }
		emptyMessagesCount++
		if emptyMessagesCount > 300 {
			return nil, errors.New("too many empty stream messages")
		}
		goto READ
	}

	noPrefixLine := bytes.TrimPrefix(noSpaceLine, headerData)
	if string(noPrefixLine) == "[DONE]" {
		return nil, io.EOF
	}
	return noPrefixLine, nil
}
