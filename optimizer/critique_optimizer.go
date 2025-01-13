package optimizer

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/prompt"
	"github.com/showntop/llmack/prompt/templates"
)

// CritiqueNOptimizer defines the CritiqueN optimizer
type CritiqueNOptimizer struct {
	options *Options
	model   *llm.Instance
	metrics []Metric
}

// NewCritiqueNOptimizer ...
func NewCritiqueNOptimizer(opts ...Option) *CritiqueNOptimizer {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return &CritiqueNOptimizer{
		options: options,
		model:   options.LLM,
		// metrics: options.Prompt,
		// strategy: options.MutationStrategy,
		// config:   options.Config,
	}
}

// Optimize ...
func (o *CritiqueNOptimizer) Optimize(ctx context.Context, target any, trainset []*Example) (*program.Program, error) {
	var targetx program.Program
	if x, ok := target.(*program.Promptx); ok {
		targetx = program.NewPredictorWithPrompt(x)
	} else if x, ok := target.(program.Promptx); ok {
		targetx = program.NewPredictorWithPrompt(&x)
	} else if x, ok := target.(program.Program); ok {
		targetx = x
	} else {
		panic("invalid optimize target, only support prompt or program")
	}

	// 生成不同风格候选指令
	var result string = ""
	err := program.Predictor().
		WithAdapter(&program.RawAdapter{}).
		WithInstruction(templates.MetaInstruction).
		ChatWith(map[string]any{
			"task_description":   targetx.Prompt().Description,
			"meta_prompts":       strings.Join(templates.ThinkingStyles[:5], "\n"),
			"num_variations":     2,
			"prompt_instruction": targetx.Prompt().Instruction,
		}).
		Result(ctx, &result)
	if err != nil {
		panic(err)
	}

	// extract results
	regx := regexp.MustCompile("(?s)<START>(.*?)<END>")
	matches := regx.FindAllStringSubmatch(result, -1)

	candidateInstructions := make([]string, len(matches))
	for i, match := range matches {
		candidateInstructions[i] = match[1]
	}

	if len(o.options.trainset) <= 0 { // 无法评估，随机选择一个 or 不支持
		panic("no trainset")
	}
	// evaluate
	// scores, failed, _ := o.evaluate(ctx, candidateInstructions, program.Outputs)
	scores, failed, _ := o.evaluate(ctx, candidateInstructions, nil)
	fmt.Println(scores)
	fmt.Println(failed)
	// 计算 scores 中值最大的 N 个下标
	topn := 2
	topNIndices := make([]int, topn)
	for i := 0; i < topn; i++ {
		maxIndex := 0
		maxScore := 0.0
		for j := 0; j < len(scores); j++ {
			if scores[j] > maxScore {
				maxIndex = j
				maxScore = scores[j]
			}
		}
		topNIndices[i] = maxIndex
		scores[maxIndex] = -1 // 将已选中的下标置为 -1
	}

	// refine instruction if
	refinedInstructions := make([]string, 0)
	for i := 0; i < len(topNIndices); i++ {
		sss := ""
		examples := failed[topNIndices[i]]
		for j := 0; j < len(examples); j++ {
			sss += "[Question] " + examples[j].Get("question").(string) + "\n"
			sss += "[Expected Answer] " + examples[j].Get("answer").(string) + "\n"
			sss += "[Wrong Answer] " + "" + "\n"
		}
		critiqueResult := o.chatCompletion(ctx, critiqueRefineMetaPrompt, map[string]any{
			"instruction": candidateInstructions[topNIndices[i]],
			"examples":    sss,
		})

		refinedPrompt := o.chatCompletion(ctx, critiqueRefinePrompt, map[string]any{
			"instruction":      candidateInstructions[topNIndices[i]],
			"examples":         sss,
			"critique":         critiqueResult,
			"steps_per_sample": 2,
		})
		fmt.Println("============================refinedPrompts============================")
		fmt.Println(i, refinedPrompt)
		fmt.Println("============================refinedPrompts============================")
		refinedInstructions = append(refinedInstructions, refinedPrompt)
	}

	// 评估
	// scores, failed, _ = o.evaluate(ctx, refinedInstructions, program.Outputs)
	scores, failed, _ = o.evaluate(ctx, refinedInstructions, nil)
	fmt.Println("scores2 ====================")
	fmt.Println(scores)
	fmt.Println(failed)

	topn = 2
	topNIndices = make([]int, topn)
	for i := 0; i < topn; i++ {
		maxIndex := 0
		maxScore := 0.0
		for j := 0; j < len(scores); j++ {
			if scores[j] > maxScore {
				maxIndex = j
				maxScore = scores[j]
			}
		}
		topNIndices[i] = maxIndex
		scores[maxIndex] = -1 // 将已选中的下标置为 -1
	}
	var topNInstructions []string
	for i := 0; i < len(topNIndices); i++ {
		topNInstructions = append(topNInstructions, refinedInstructions[topNIndices[i]])
	}
	fmt.Println("topNInstructions: ", topNInstructions)
	// finalInstruction := topNInstructions[0]
	// program.Instruction = finalInstruction
	// o.createNFewShotExamples(ctx, promptx)
	// 	for i := 0; i < len(candidateInstructions); i++ {
	// 		expertIdentity := o.chatCompletion(ctx, templates.ExpertInstruction, map[string]any{
	// 			"task_description": promptx.Instruction,
	// 		})
	// 		fmt.Println("expert identity: ", expertIdentity)
	// 		intentKeywords := o.chatCompletion(ctx, templates.IntentInstruction, map[string]any{
	// 			"task_description": promptx.Instruction,
	// 			"base_instruction": promptx.Instruction,
	// 		})
	// 		fmt.Println("intent keywords: ", intentKeywords)
	// 		// 没有训练集
	// 		finalPrompt := `
	// Expert Profile:
	// {{expert_identity}}
	// Keywords:
	// {{intent_keywords}}

	// {{instruction}}
	// {{few_shot_examples}}

	// {{answer_format}}
	// 		`
	// 		vvvv, _ := prompt.Render(finalPrompt, map[string]any{
	// 			"instruction":       candidatePrompts[i],
	// 			"few_shot_examples": "",
	// 			"expert_identity":   expertIdentity,
	// 			"answer_format":     "",
	// 			"intent_keywords":   intentKeywords,
	// 		})
	// 		fmt.Println("============================== final prompt ==============================")
	// 		fmt.Println(vvvv)
	// 		fmt.Println("==========================================================================")

	// 	}

	// 生成样本数据 zero-shot
	// o.chatCompletion(ctx, promptx, map[string]any{})

	return &targetx, nil
}

func (o *CritiqueNOptimizer) createNFewShotExamples(ctx context.Context, promptx *prompt.Prompt) (string, error) {
	critique := o.chatCompletion(ctx, critiqueExampleGeneratePrompt, map[string]any{
		"task_description": promptx.Description,
		"prompt":           promptx.InitialInstruction,
		"num_examples":     3,
	})
	result := o.chatCompletion(ctx, critiqueExampleOptimizationPrompt, map[string]any{
		"prompt":           promptx.InitialInstruction,
		"examples":         "",
		"gt_example":       "",
		"critique":         critique,
		"task_description": promptx.Description,
		"num_examples":     3,
	})
	fmt.Println("xxxxxxxxxresult: ", result)

	// extract examples from response
	regx := regexp.MustCompile("(?s)<START>(.*?)<END>")
	matches := regx.FindAllStringSubmatch(result, -1)

	items := make([]string, len(matches))
	for i, match := range matches {
		items[i] = match[1]
	}
	examples := make([]*Example, len(items))
	for i, item := range items {
		fmt.Println("item: ", item)
		// examples[i] = Example{
		// 	Question: item,
		// }
		examples[i] = Examplex("question", item)
	}
	fmt.Println("examples: ", examples)

	return critique, nil
}

func (o *CritiqueNOptimizer) evaluate(ctx context.Context, instructions []string, output any) ([]float64, [][]*Example, [][]string) {
	scores := make([]float64, len(instructions), len(instructions))
	failed := make([][]*Example, len(instructions), len(instructions))
	failedAnswers := make([][]string, len(instructions), len(instructions))
	if len(instructions) <= 0 {
		return scores, failed, failedAnswers
	}

	bestInstruction := instructions[0]
	fmt.Println("==================evaluate===================")
	for i := 0; i < len(instructions); i++ {
		correctNum := 0
		totalNum := 0
		incorrected := []*Example{}
		answers := []string{}
		for j := 0; j < len(o.options.trainset); j++ {
			eval := NewEvaluator(5,
				[]*Example{o.options.trainset[j]}, o.options.metric)
			score := eval.Score(ctx, instructions[i]+"\n"+output.(string))
			if int(score) == 1 {
				correctNum++
			} else {
				incorrected = append(incorrected, o.options.trainset[j])
				answers = append(answers)
			}
			totalNum++
		}
		scores[i] = float64(correctNum) / float64(totalNum)

		failed[i] = incorrected
		failedAnswers[i] = answers
	}
	fmt.Println("==================evaluate===================")
	fmt.Println("bestInstruction:", bestInstruction)
	return scores, failed, failedAnswers
}

func (o *CritiqueNOptimizer) chatCompletion(ctx context.Context, template string, inputs map[string]any) string {
	userPrompt, _ := prompt.Render(template, inputs)

	messages := []llm.Message{
		llm.SystemPromptMessage(" "),
		llm.UserPromptMessage(userPrompt),
	}

	response, err := o.model.Invoke(ctx, messages, nil,
		llm.WithModel(o.options.LLMModel),
		llm.WithStream(true),
	)
	if err != nil {
		panic(err)
	}
	return response.Result().Message.Content().Data
}
