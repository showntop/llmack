package image

import (
	"context"

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
	t := &tool.Tool{}
	t.Name = SiliconflowImageGenerate
	t.Kind = "code"
	t.Description = "Generate Images using Siliconflow"
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name: "prompt", Type: tool.String, Required: true, LLMDescrition: "The text prompt used to generate the image.", Default: "",
	}, tool.Parameter{
		Name: "image_size", Type: tool.String, Required: true, LLMDescrition: "Choose Image Size.", Default: "768x512",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		apiKey := tool.DefaultConfig.GetString("siliconflow.api_key")

		prompt := args["prompt"].(string)
		return vision.NewInstance(siliconflow.Name).GenerateImage(ctx, prompt, vision.WithApiKey(apiKey))
	}
	tool.Register(t)
}

func registMinimax() {
	t := &tool.Tool{}
	t.Name = MinimaxImageGenerate
	t.Kind = "code"
	t.Description = "Generate Images using Minimax"
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name: "prompt", Type: tool.String, Required: true, LLMDescrition: "The text prompt used to generate the image.", Default: "",
	}, tool.Parameter{
		Name: "image_size", Type: tool.String, Required: true, LLMDescrition: "Choose Image Size.", Default: "768x512",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		apiKey := tool.DefaultConfig.GetString("minimax.api_key")

		prompt := args["prompt"].(string)
		vision.NewInstance(minimax.Name).GenerateImage(ctx, prompt, vision.WithApiKey(apiKey))
		return "", nil
	}
	tool.Register(t)
}
