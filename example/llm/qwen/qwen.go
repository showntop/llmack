package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/qwen"
	"github.com/showntop/llmack/log"
)

func init() {
	godotenv.Load()
}

func main() {
	runWithCache()
}

func runWithCache() {
	ctx := context.Background()

	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("qwen_api_key"),
	})

	instance := llm.NewInstance(qwen.Name,
		llm.WithCache(llm.NewMemoCache()),
		llm.SetLogger(&log.WrapLogger{}),
	)

	resp, err := instance.Invoke(ctx,
		[]llm.Message{
			llm.NewUserTextMessage("你好"),
			llm.NewUserMultipartMessage(
				llm.MultipartContentImageURL("https://help-static-aliyun-doc.aliyuncs.com/file-manage-files/zh-CN/20241022/emyrja/dog_and_girl.jpeg"),
				llm.MultipartContentText("这是一张关于猫的照片吗"),
			),
		},
		llm.WithModel("qwen-vl-max-latest"),
		llm.WithStream(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Result())

	resp, err = instance.Invoke(ctx,
		[]llm.Message{llm.NewUserTextMessage("你好")},
		llm.WithModel("qwen-vl-max-latest"),
		llm.WithStream(true),
	)
	if err != nil {
		panic(err)
	}
	// stream := resp.Stream()
	// for v := stream.Next(); v != nil; v = stream.Next() {
	// 	fmt.Println(string(v.Delta.Message.Content().Data))
	// }
	fmt.Println(resp.Result())
}
