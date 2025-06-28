package image

import (
	"context"
	"encoding/json"

	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/vision"

	"github.com/showntop/llmack/vision/minimax"
	"github.com/showntop/llmack/vision/siliconflow"
)

const MinimaxImageGenerate = "MinimaxImageGenerate"
const SiliconflowImageGenerate = "SiliconflowImageGenerate"

func init() {
	registMinimax()
	registSiliconflow()
}

func registSiliconflow() {
	t := tool.New(
		tool.WithName(SiliconflowImageGenerate),
		tool.WithKind("code"),
		tool.WithDescription("Generate Images using Siliconflow"),
		tool.WithParameters(
			tool.Parameter{
				Name: "prompt", Type: tool.String, Required: true, LLMDescrition: "The text prompt used to generate the image.", Default: "",
			},
			tool.Parameter{
				Name: "image_size", Type: tool.String, Required: true, LLMDescrition: "Choose Image Size.", Default: "768x512",
			},
		),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params struct {
				Prompt    string `json:"prompt"`
				ImageSize string `json:"image_size"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}
			apiKey := tool.DefaultConfig.GetString("siliconflow.api_key")
			return vision.NewInstance(siliconflow.Name).GenerateImage(ctx, params.Prompt, vision.WithApiKey(apiKey))
		}),
	)
	tool.Register(t)
}

func registMinimax() {
	t := tool.New(
		tool.WithName(MinimaxImageGenerate),
		tool.WithKind("code"),
		tool.WithDescription("Generate Images using Minimax"),
		tool.WithParameters(
			tool.Parameter{
				Name: "prompt", Type: tool.String, Required: true, LLMDescrition: "The text prompt used to generate the image.", Default: "",
			},
			tool.Parameter{
				Name: "image_size", Type: tool.String, Required: true, LLMDescrition: "Choose Image Size.", Default: "768x512",
			},
		),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params struct {
				Prompt    string `json:"prompt"`
				ImageSize string `json:"image_size"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}
			apiKey := tool.DefaultConfig.GetString("minimax.api_key")
			result, err := vision.NewInstance(minimax.Name).GenerateImage(ctx, params.Prompt, vision.WithApiKey(apiKey))
			if err != nil {
				return "", err
			}
			return result, nil
		}),
	)
	tool.Register(t)
}
