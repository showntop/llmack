package services

import (
	"context"
	"log"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/qwen"

	"github.com/bookkeeper-ai/bookkeeper/config"
)

type ChatService struct {
	llmClient *llm.Instance
}

func NewChatService(cfg *config.Config) (*ChatService, error) {
	// 初始化 LLM 客户端
	// llmClient := llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))
	llmClient := llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max"))

	return &ChatService{
		llmClient: llmClient,
	}, nil
}

func (s *ChatService) SendMessage(ctx context.Context, message string, images []string) (string, error) {
	multipartContents := make([]*llm.MultipartContent, 0)
	multipartContents = append(multipartContents, llm.MultipartContentText(message))
	for _, image := range images {
		multipartContents = append(multipartContents, llm.MultipartContentImageURL(image))
	}

	// 构建消息
	messages := []llm.Message{
		llm.NewSystemMessage("你是一个智能记账助手，根据用户的输入识别其中的内容，并根据内容生成记账凭证。"),
		llm.NewUserMultipartMessage(
			multipartContents...,
		),
	}

	// 调用 LLM
	response, err := s.llmClient.Invoke(ctx, messages, llm.WithStream(true))
	if err != nil {
		log.Printf("Error invoking LLM: %v", err)
		return "", err
	}
	return response.Result().Message.Content(), nil
}
