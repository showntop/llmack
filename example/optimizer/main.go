package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/showntop/llmack/llm"
	mlog "github.com/showntop/llmack/log"
	"github.com/showntop/llmack/program"

	"github.com/joho/godotenv"
	azureoai "github.com/showntop/llmack/llm/azure-openai"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/llm/moonshot"
	oaic "github.com/showntop/llmack/llm/openai-c"
	"github.com/showntop/llmack/llm/zhipu"
	"github.com/showntop/llmack/optimizer"
)

var (
	LLMProvider = deepseek.Name
	LLMModel    = "deepseek-chat"
	// LLMProvider = oaic.Name
	// LLMModel    = "hunyuan"
	// LLMProvider = azureoai.Name
	// LLMModel    = "gpt-4o"
)

func init() {
	godotenv.Load()

	llm.WithConfigs(map[string]any{
		moonshot.Name: map[string]any{
			"api_key": os.Getenv("moonshot_api_key"),
		},
		azureoai.Name: map[string]any{
			"endpoint": os.Getenv("azure_openai_endpoint"),
			"api_key":  os.Getenv("azure_openai_api_key"),
			// "api_version": "2024-02-15-preview",
		},
		oaic.Name: map[string]any{
			"base_url": os.Getenv("hunyuan_base_url"),
			"api_key":  os.Getenv("hunyuan_api_key"),
		},
		"hunyuan": map[string]any{
			"api_key": os.Getenv("hunyuan_api_key"),
		},
		deepseek.Name: map[string]any{
			"api_key": os.Getenv("deepseek_api_key"),
		},
		zhipu.Name: map[string]any{
			"api_key": os.Getenv("zhipu_api_key"),
		},
	})
}

func loadDataset() []*optimizer.Example {
	file := "example/optimizer/train.jsonl"
	examples := []*optimizer.Example{}

	data, _ := os.ReadFile("core/example/optimizer/dataset/" + file)

	for _, line := range strings.Split(string(data), "\n") {
		if line != "" {
			var example *optimizer.Example
			json.Unmarshal([]byte(line), example)
			examples = append(examples, example)
		}
	}
	return examples

}

var trainset = []*optimizer.Example{
	optimizer.Examplex("question", "At My Window was released by which American singer-songwriter?", "answer", "John Townes Van Zandt").WithInputKeys("question"),
	optimizer.Examplex("question", "which American actor was Candace Kita guest starred with ", "answer", "Bill Murray").WithInputKeys("question"),
	optimizer.Examplex("question", "Which of these publications was most recently published, Who Put the Bomp or Self?", "answer", "Self").WithInputKeys("question"),
	optimizer.Examplex("question", "The Victorians - Their Story In Pictures is a documentary series written by an author born in what year?", "answer", "1950").WithInputKeys("question"),
	optimizer.Examplex("question", "Which magazine has published articles by Scott Shaw, Tae Kwon Do Times or Southwest Art?", "answer", "Tae Kwon Do Times").WithInputKeys("question"),
	optimizer.Examplex("question", "In what year was the club founded that played Manchester City in the 1972 FA Charity Shield", "answer", "1874").WithInputKeys("question"),
	optimizer.Examplex("question", "Which is taller, the Empire State Building or the Bank of America Tower?", "answer", "The Empire State Building").WithInputKeys("question"),
	optimizer.Examplex("question", "Which American actress who made their film debut in the 1995 teen drama \"Kids\" was the co-founder of Voto Latino?", "answer", "Rosario Dawson").WithInputKeys("question"),
	optimizer.Examplex("question", "Tombstone stared an actor born May 17, 1955 known as who?", "answer", "Bill Paxton").WithInputKeys("question"),
	optimizer.Examplex("question", "What is the code name for the German offensive that started this Second World War engagement on the Eastern Front (a few hundred kilometers from Moscow) between Soviet and German forces, which included 102nd Infantry Division?", "answer", "Operation Citadel").WithInputKeys("question"),
	optimizer.Examplex("question", "Who acted in the shot film The Shore and is also the youngest actress ever to play Ophelia in a Royal Shakespeare Company production of \"Hamlet.\" ?", "answer", "Kerry Condon").WithInputKeys("question"),
	optimizer.Examplex("question", "Which company distributed this 1977 American animated film produced by Walt Disney Productions for which Sherman Brothers wrote songs?", "answer", "Buena Vista Distribution").WithInputKeys("question"),
	optimizer.Examplex("question", "Samantha Cristoforetti and Mark Shuttleworth are both best known for being first in their field to go where? ", "answer", "space").WithInputKeys("question"),
	optimizer.Examplex("question", "Having the combination of excellent foot speed and bat speed helped Eric Davis, create what kind of outfield for the Los Angeles Dodgers? ", "answer", "\"Outfield of Dreams\"").WithInputKeys("question"),
	optimizer.Examplex("question", "Which Pakistani cricket umpire who won 3 consecutive ICC umpire of the year awards in 2009, 2010, and 2011 will be in the ICC World Twenty20?", "answer", "Aleem Sarwar Dar").WithInputKeys("question"),
	optimizer.Examplex("question", "The Organisation that allows a community to influence their operation or use and to enjoy the benefits arisingwas founded in what year?", "answer", "2010").WithInputKeys("question"),
	optimizer.Examplex("question", "\"Everything Has Changed\" is a song from an album released under which record label ?", "answer", "Big Machine Records").WithInputKeys("question"),
	optimizer.Examplex("question", "Who is older, Aleksandr Danilovich Aleksandrov or Anatoly Fomenko?", "answer", "Aleksandr Danilovich Aleksandrov").WithInputKeys("question"),
	optimizer.Examplex("question", "On the coast of what ocean is the birthplace of Diogal Sakho?", "answer", "Atlantic").WithInputKeys("question"),
	optimizer.Examplex("question", "This American guitarist best known for her work with the Iron Maidens is an ancestor of a composer who was known as what?", "answer", "The Waltz King").WithInputKeys("question"),
}

func coproOptimizer() {
	metric := func() optimizer.Metric {
		sf1 := optimizer.NewSemanticF1(context.Background())

		return func(e *optimizer.Example, prediction any) float64 {
			if _, ok := prediction.(map[string]any); !ok {
				return 0.0
			}
			actual := prediction.(map[string]any)["answer"].(string)
			fmt.Println("metrics metricsmetricsmetrics: ", actual, "|||", e.Get("answer").(string))
			examplex := optimizer.Examplex("question", e.Get("question"), "ground_truth", e.Get("answer"))
			predictionx := map[string]any{"response": actual}
			return sf1.Metric(examplex, predictionx)
		}
	}

	testp := program.NewPredictor("Answer the question and give the reasoning for the same.",
		map[string]string{"question": "question about something"},
		program.WithOutput("answer", "often between 1 and 5 words"),
	)

	opter := optimizer.NewCoproOptimizer(3, 2, metric())
	_ = opter.Optimize(context.Background(), testp, trainset)

}

func main() {
	// Initialize prompt
	// p := prompt.New(
	// 	prompt.WithTemplate("Given {{.input}}, provide {{.output}}"),
	// 	prompt.WithExamples([]prompt.Example{
	// 		{Input: "example1", Output: "result1"},
	// 		{Input: "example2", Output: "result2"},
	// 	}),
	// )

	metric := func(e *optimizer.Example, prediction any) float64 {
		if _, ok := prediction.(map[string]any); !ok {
			return 0.0
		}
		actual := prediction.(map[string]any)["answer"].(string)
		fmt.Println("metrics metricsmetricsmetrics: ", actual, "|||", e.Get("answer").(string))
		return optimizer.EM(actual, []string{e.Get("answer").(string)})
	}

	trainset := loadDataset() // [:2]

	// Configure optimizer
	// opter := optimizer.NewMiproOptimizer(
	opter := optimizer.NewCritiqueNOptimizer(
		optimizer.WithLLM(llm.NewInstance(LLMProvider, llm.WithLogger(&mlog.WrapLogger{})), LLMModel),
		// optimizer.WithLLM(llm.NewInstance(oaic.Name, llm.WithLogger(&log.WrapLogger{})), "hunyuan"),
		optimizer.WithMetric(metric),
		optimizer.WithTrainset(trainset),
		// optimizer.WithStrategy(mutation.NewTemplateModifier()),
	)

	target := &program.Promptx{
		Name:        "test",
		Instruction: "You are a mathematics expert. You will be given a mathematics problem which you need to solve",
		OutputFields: map[string]*program.Field{
			"answer": {
				Description: "The answer to the problem",
				Marker:      "ANS",
			},
		},
	}

	// Run optimization
	optimizedProgram, err := opter.Optimize(context.Background(), target, trainset)
	if err != nil {
		log.Fatal(err)
	}
	_ = optimizedProgram
	// fmt.Println(optimizedProgram.Prompt())

	// Use optimized prompt
	// result, err := opt.Model().Complete(optimizedPrompt.Render(map))
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func metric(ex *optimizer.Example, actualx any) float64 {
	expected := ex.Get("answer")
	actual := actualx.(string)
	fmt.Println("000 actual answer: ", actual)
	start, end := 0, 0
	if x := strings.Index(actual, "<ANS_START>"); x == -1 {
		x = strings.Index(actual, "</ANS_START>")
		start = x + len("</ANS_START>")
	} else {
		start = x + len("<ANS_START>")
	}

	if x := strings.Index(actual, "<ANS_END>"); x == -1 {
		if x = strings.Index(actual, "</ANS_END>"); x == -1 {
			end = strings.Index(actual, "</ANS_START>")
		} else {
			end = x
		}
	} else {
		end = x
	}
	if start == -1 || end == -1 {
		return 0
	}
	actual = actual[start:end]
	fmt.Println("111 actual answer: ", actual, expected)
	return optimizer.AccuracyMatch(actual, expected)
}

func testProgram(ctx context.Context, prompt string) (string, error) {
	messages := []llm.Message{
		llm.UserPromptMessage(prompt),
	}
	instance := llm.NewInstance(LLMProvider, llm.WithLogger(&mlog.WrapLogger{}))
	response, err := instance.Invoke(ctx, messages, nil,
		llm.WithStream(true), llm.WithModel(LLMModel))
	// llm.WithStream(true), llm.WithModel("hunyuan"))
	if err != nil {
		panic(err)
	}
	return response.Result().Message.Content().Data, nil
	// actual :=
	// score := e.metric(actual, ex.Answer)
	// return score
}
