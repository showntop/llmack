package hunyuan

import (
	"context"
	"encoding/json"

	"github.com/showntop/llmack/llm"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	hunyuan "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/hunyuan/v20230901"
)

// Name ...
var Name = "hunyuan"

func init() {
	llm.Register(Name, &LLM{})
}

// LLM ...
type LLM struct {
	client *hunyuan.Client
}

// Invoke ...
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, options ...llm.InvokeOption) (*llm.Response, error) {
	if err := m.setupClient(); err != nil { // TODO sync.Once
		return nil, err
	}
	var opts llm.InvokeOptions
	for _, o := range options {
		o(&opts)
	}

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := hunyuan.NewChatCompletionsRequest()
	for _, m := range messages {
		role := string(m.Role())
		request.Messages = append(request.Messages, &hunyuan.Message{
			Role:    &role,
			Content: &m.Content().Data,
		})
	}
	// tools
	if len(opts.Tools) > 0 {
		fuction := "function"
		auto := "auto"
		request.ToolChoice = &auto
		request.Tools = make([]*hunyuan.Tool, len(opts.Tools))
		for i, t := range opts.Tools {
			raw, _ := json.Marshal(t.Function.Parameters)
			params := string(raw)
			request.Tools[i] = &hunyuan.Tool{
				Type: &fuction,
				Function: &hunyuan.ToolFunction{
					Name:        &t.Function.Name,
					Description: &t.Function.Description,
					Parameters:  &params,
				},
			}
		}
	}
	request.Stream = &opts.Stream
	request.Model = &opts.Model
	// 返回的resp是一个ChatCompletionsResponse的实例，与请求对象对应
	chatResponse, err := m.client.ChatCompletions(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	// 流式响应
	response := llm.NewStreamResponse()
	go func() {
		defer response.Stream().Close()

		currentTool := llm.ToolCall{}
		toolCalls := []llm.ToolCall{}
		for event := range chatResponse.Events {
			chunk := event.Data
			var data SSEData
			json.Unmarshal(chunk, &data)
			if len(data.Choices) <= 0 {
				continue
			}
			mmm := llm.AssistantPromptMessage(data.Choices[0].Delta.Content)
			for _, c := range data.Choices[0].Delta.ToolCalls {
				if c.Id == currentTool.ID {
					currentTool.ID = c.Id
					currentTool.Function.Name += c.Function.Name
					currentTool.Function.Arguments += c.Function.Arguments
				} else {
					toolCalls = append(toolCalls, currentTool)
					currentTool = llm.ToolCall{}
				}
			}

			// 直到finish
			if false {
				// mmm.ToolCalls = toolCalls
				response.Stream().Push(llm.NewChunk(0, mmm, nil))
			} else {
				response.Stream().Push(llm.NewChunk(0, mmm, nil))
			}

		}
	}()

	return response, nil
}

type SSEData struct {
	Note    string
	Choices []struct {
		Delta struct {
			Role      string
			Content   string
			ToolCalls []struct {
				Id       string
				Type     string
				Function struct {
					Name      string
					Arguments string
				}
			}
		}

		FinishReason string
	}
}

func (m *LLM) setupClient() error {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	credential := common.NewCredential(
		"",
		"",
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "hunyuan.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	// common.DefaultHttpClient = http.DefaultClient
	client, err := hunyuan.NewClient(credential, "ap-beijing", cpf)
	if err != nil {
		return err
	}
	m.client = client
	return nil
}
