package optimizer

// import (
// 	"context"

// 	"golang.org/x/exp/rand"
// )

// // MiproV2Optimizer ...
// type MiproV2Optimizer struct {
// }

// // NewMiproV2Optimizer ...
// func NewMiproV2Optimizer(opts ...Option) Optimizer {
// 	options := &Options{}
// 	for _, opt := range opts {
// 		opt(options)
// 	}
// 	return &MiproV2Optimizer{
// 		// Optimizer: Optimizer{
// 		// 	options: options,
// 		// 	model:   options.LLM,
// 		// 	// metrics: options.Prompt,
// 		// 	// strategy: options.MutationStrategy,
// 		// 	// config:   options.Config,
// 		// },
// 	}
// }

// // Optimize ...
// func (o *MiproV2Optimizer) Optimize(ctx context.Context, p any, optFuncs ...OptimizeOption) (*OptimizeResult, error) {

// 	// 	for i := 0; i < len(candidatePrompts); i++ {
// 	// 		expertIdentity := o.chatCompletion(ctx, templates.ExpertInstruction, map[string]any{
// 	// 			"task_description": promptx.Instruction,
// 	// 		})
// 	// 		fmt.Println("expert identity: ", expertIdentity)
// 	// 		intentKeywords := o.chatCompletion(ctx, templates.IntentInstruction, map[string]any{
// 	// 			"task_description": promptx.Instruction,
// 	// 			"base_instruction": promptx.Instruction,
// 	// 		})
// 	// 		fmt.Println("intent keywords: ", intentKeywords)
// 	// 		// 没有训练集
// 	// 		finalPrompt := `
// 	// Expert Profile:
// 	// {{expert_identity}}
// 	// Keywords:
// 	// {{intent_keywords}}

// 	// {{instruction}}
// 	// {{few_shot_examples}}

// 	// {{answer_format}}
// 	// 		`
// 	// 		vvvv, _ := prompt.Render(finalPrompt, map[string]any{
// 	// 			"instruction":       candidatePrompts[i],
// 	// 			"few_shot_examples": "",
// 	// 			"expert_identity":   expertIdentity,
// 	// 			"answer_format":     "",
// 	// 			"intent_keywords":   intentKeywords,
// 	// 		})
// 	// 		fmt.Println("============================== final prompt ==============================")
// 	// 		fmt.Println(vvvv)
// 	// 		fmt.Println("==========================================================================")

// 	// 	}

// 	// 生成样本数据 zero-shot
// 	// o.chatCompletion(ctx, promptx, map[string]any{})

// 	return promptx, nil
// }

// func (o *MiproV2Optimizer) sampleFewshot(trainset []Example, numSamples int) ([]Example, []int) {
// 	sampledIndices := randomSampleInts(len(trainset), numSamples)

// 	result := make([]Example, 0)
// 	for _, index := range sampledIndices {
// 		result = append(result, trainset[index])
// 	}
// 	return result, sampledIndices
// }

// // create few-shot examples
// func (o *MiproV2Optimizer) createNFewShotExamples(ctx context.Context, trainset []Example) []Example {
// 	candidateNum := 4
// 	maxLabeledCandidateNum := 10      // maximum number of labeled candidate
// 	maxBootstrappedCandidateNum := 10 // maximum number of labeled candidate

// 	candidates := make([]Example, 0)

// 	// 添加no-shot候选
// 	candidates = append(candidates, Example{})

// 	// 添加采样的fewshot候选
// 	sampledFewshot, sampledIndices := o.sampleFewshot(trainset, maxLabeledCandidateNum)
// 	_ = sampledIndices
// 	candidates = append(candidates, Prompt{Fewshot: sampledFewshot})

// 	// 添加引导式候选
// 	for i := 0; i < candidateNum-2; i++ {
// 		maxBootstrapped := rand.Intn(maxBootstrappedCandidateNum) + 1
// 		maxLabeled := rand.Intn(maxLabeledCandidateNum) + 1
// 		samples, indices := t.sample(trainset, predictions, evalResults, maxBootstrapped, maxLabeled)
// 		candidates = append(candidates, Example{input})
// 	}

// 	return candidates
// }

// // // createNFewshotDemoSets 创建多个fewshot示例集
// // func (t *DspyMiproV2Optimizer) createNFewshotDemoSets(
// // 	trainset []DatasetItem,
// // 	predictions []interface{},
// // 	evalResults []MetricResult,
// // ) ([]Prompt, [][]int) {

// // }

// // randomSampleInts 从给定范围内随机采样 k 个整数
// func randomSampleInts(n int, k int) []int {
// 	indices := make([]int, 0)
// 	for i := 0; i < k; i++ {
// 		indices = append(indices, rand.Intn(n))
// 	}
// 	return indices
// }
