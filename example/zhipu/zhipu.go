package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/zhipu"
)

func init() {
	godotenv.Load()
}

func main() {
	ctx := context.Background()
	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("zhipu_api_key"),
	})

	resp, err := llm.NewInstance(zhipu.Name).Invoke(ctx,
		[]llm.Message{
			// llm.UserTextPromptMessage("你好"),
			llm.NewUserMultipartMessage(
				llm.MultipartContentImageURL("https://help-static-aliyun-doc.aliyuncs.com/file-manage-files/zh-CN/20241022/emyrja/dog_and_girl.jpeg"),
				llm.MultipartContentText("这是一张关于猫的照片吗"),
			),
		},

		llm.WithModel("glm-4v-plus"))
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Result())
}
