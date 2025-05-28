package browser

import (
	"context"

	"github.com/showntop/llmack/tool"
)

type Browser struct {
	tool.Tool
}

func (b *Browser) DoAction(ctx context.Context, args map[string]any) (string, error) {
	return "", nil
}

func (b *Browser) Tools() ([]*tool.Tool, error) {
	return []*tool.Tool{
		{
			Name:        "NavigateTo",
			Description: "Navigates to a URL.",
			Parameters: []tool.Parameter{
				{
					Name:          "url",
					LLMDescrition: "The URL to navigate to.",
					Type:          tool.String,
					Required:      true,
				},
				{
					Name:          "connect_url",
					LLMDescrition: "The connection URL from an existing session",
					Type:          tool.String,
					Required:      false,
				},
			},
			Invokex: b.Navigate,
		},
		{
			Name:        "Screenshot",
			Description: "Takes a screenshot of the current page.",
			Parameters: []tool.Parameter{
				{
					Name:          "path",
					LLMDescrition: "Where to save the screenshot",
					Type:          tool.String,
					Required:      true,
				},
				{
					Name:          "full_page",
					LLMDescrition: "Whether to capture the full page",
					Type:          tool.Boolean,
					Required:      false,
				},
				{
					Name:          "connect_url",
					LLMDescrition: "The connection URL from an existing session",
					Type:          tool.String,
					Required:      false,
				},
			},
			Invokex: b.Screenshot,
		},
		{
			Name:        "GetPageContent",
			Description: "Gets the content of the current page.",
			Parameters:  []tool.Parameter{},
			Invokex:     b.GetPageContent,
		},
	}, nil
}
