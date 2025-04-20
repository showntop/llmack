package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/weather"
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})

	llm.WithConfigs(map[string]any{
		deepseek.Name: map[string]any{
			"api_key": os.Getenv("deepseek_api_key"),
		},
	})
	program.SetLLM(deepseek.Name, "deepseek-chat")
}

func main() {
	funcall()
	// funcallStream()
}

func funcall() {
	predictor := program.FunCall(
		program.WithLLM(deepseek.Name, "deepseek-chat"),
	).
		WithInstruction("你是一个计算器，根据用户问题计算结果").
		WithTools(getCalculator()).
		InvokeQuery(context.Background(), "1+1")
	if predictor.Error() != nil {
		panic(predictor.Error())
	}

	fmt.Println(predictor.Completion())
}

func funcallStream() {
	predictor := program.FunCall(
		program.WithLLM(deepseek.Name, "deepseek-chat"),
	).
		WithInstruction("你是一个旅行规划师，根据用户问题规划行程").
		WithTools(getWeather()).
		WithStream(true).
		InvokeQuery(context.Background(), "去北京三里屯")
	if predictor.Error() != nil {
		panic(predictor.Error())
	}

	for chunk := range predictor.Stream() {
		fmt.Println(chunk.Choices[0].Message.Content())
	}

}

func getWeather() string {
	return weather.QueryWeather
}

func getCalculator() string {
	calculator := &tool.Tool{}
	calculator.Name = "calculator"
	calculator.Kind = "code"
	calculator.Description = "Use this function to calculate the result of the user's question."
	calculator.Parameters = append(calculator.Parameters, tool.Parameter{
		Name:          "question",
		Type:          "string",
		LLMDescrition: "The user's question to calculate the result.",
	})
	calculator.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		return "2", nil
	}

	tool.Register(calculator)
	return calculator.Name
}

func cot() {
	var target struct {
		Out1 string `json:"out1"`
		Out2 bool   `json:"out2"`
	}
	err := program.COT().
		WithAdapter(&program.MarkableAdapter{}).
		WithInstruction("hello？").
		WithInputField("hello", "world").
		WithOutputField("out1", "description", "marker", reflect.String).
		WithOutputField("out2", "description", "marker", reflect.Bool).
		Result(context.Background(), &target)
	if err != nil {
		panic(err)
	}
	fmt.Println(target)
}
