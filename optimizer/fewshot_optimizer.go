package optimizer

// import (
// 	"context"
// 	"math/rand"
// 	"strings"
// )

// type DatasetItem struct {
// 	Inputs  string
// 	Outputs string
// }

// type MetricResult struct {
// 	Score float64
// }

// type Prompt struct {
// 	Messages []string
// 	Fewshot  []DatasetItem
// }

// type FewShotTrainerReport struct {
// 	Scores     []map[string]float64
// 	Choices    []map[string]interface{}
// 	BestParams map[string]interface{}
// 	BestScore  float64
// }

// // FewShotOptimizer ...
// type FewShotOptimizer struct {
// 	RandomSeed           int
// 	MaxBootstrappedDemos int
// 	MaxLabeledDemos      int
// 	SuccessScore         float64
// 	NumCandidates        int
// }

// // NewFewShotOptimizer ...
// func NewFewShotOptimizer(randomSeed, maxBootstrappedDemos, maxLabeledDemos, numCandidates int, successScore float64) Optimizer {
// 	rand.Seed(int64(randomSeed))
// 	return &FewShotOptimizer{
// 		RandomSeed:           randomSeed,
// 		MaxBootstrappedDemos: maxBootstrappedDemos,
// 		MaxLabeledDemos:      maxLabeledDemos,
// 		SuccessScore:         successScore,
// 		NumCandidates:        numCandidates,
// 	}
// }

// // Optimize ...
// func (f *FewShotOptimizer) Optimize(ctx context.Context, p any, optFuncs ...OptimizeOption) (*OptimizeResult, error) {
// 	report := FewShotTrainerReport{
// 		Scores:     []map[string]float64{},
// 		Choices:    []map[string]interface{}{},
// 		BestParams: map[string]interface{}{},
// 		BestScore:  -1.0,
// 	}

// 	messagesStr := ""
// 	for _, message := range prompt.Messages {
// 		messagesStr += message
// 	}

// 	if !strings.Contains(messagesStr, "{_FEWSHOT_}") {
// 		prompt = f.GenerateFewshotPlaceholder(prompt)
// 	}

// 	bestScore := -1.0
// 	var bestFewshot []DatasetItem

// 	preds, evalResults := f.Evaluate(trainset, prompt)

// 	fewshotCandidates, fewshotCandidateIndices := f.CreateNFewshotDemoSets(trainset, preds, evalResults)

// 	for i, candidate := range fewshotCandidates {
// 		tempPrompt := prompt
// 		tempPrompt.Fewshot = candidate
// 		indices := fewshotCandidateIndices[i]
// 		trainsetWithoutFewshot := make([]DatasetItem, 0)
// 		for idx, item := range trainset {
// 			if !contains(indices, idx) {
// 				trainsetWithoutFewshot = append(trainsetWithoutFewshot, item)
// 			}
// 		}
// 		_, _, globalResult := f.Evaluate(trainsetWithoutFewshot, tempPrompt)

// 		report.Scores = append(report.Scores, map[string]float64{"step": float64(i), "score": globalResult.Score})
// 		report.Choices = append(report.Choices, map[string]interface{}{"step": i, "fewshot": candidate})

// 		if globalResult.Score > bestScore {
// 			bestScore = globalResult.Score
// 			bestFewshot = candidate
// 		}
// 	}

// 	prompt.Fewshot = bestFewshot
// 	report.BestParams = map[string]interface{}{"fewshot": bestFewshot}
// 	report.BestScore = bestScore
// 	return prompt, report
// }

// func (f *FewShotOptimizer) CreateNFewshotDemoSets(trainset []DatasetItem, predictions []string, evalResults []MetricResult) ([][]DatasetItem, [][]int) {
// 	candidates := [][]DatasetItem{}
// 	candidateIndices := [][]int{}

// 	// Add no-shot candidate
// 	candidates = append(candidates, []DatasetItem{})
// 	candidateIndices = append(candidateIndices, []int{})

// 	// Add sampled-fewshot candidate
// 	sampledFewshot, sampledIndices := f.SampleFewshot(trainset, f.MaxLabeledDemos)
// 	candidates = append(candidates, sampledFewshot)
// 	candidateIndices = append(candidateIndices, sampledIndices)

// 	// Add bootstrapped candidates
// 	for i := 0; i < f.NumCandidates-2; i++ {
// 		maxBootstrapped := rand.Intn(f.MaxBootstrappedDemos) + 1
// 		maxLabeled := rand.Intn(f.MaxLabeledDemos) + 1
// 		shuffledTrainset := make([]DatasetItem, len(trainset))
// 		copy(shuffledTrainset, trainset)
// 		rand.Shuffle(len(shuffledTrainset), func(i, j int) {
// 			shuffledTrainset[i], shuffledTrainset[j] = shuffledTrainset[j], shuffledTrainset[i]
// 		})
// 		samples, indices := f.Sample(shuffledTrainset, predictions, evalResults, maxBootstrapped, maxLabeled)
// 		candidates = append(candidates, samples)
// 		candidateIndices = append(candidateIndices, indices)
// 	}
// 	return candidates, candidateIndices
// }

// func (f *FewShotOptimizer) SampleFewshot(trainset []DatasetItem, numSamples int) ([]DatasetItem, []int) {
// 	sampledIndices := rand.Perm(len(trainset))[:numSamples]
// 	sampledItems := make([]DatasetItem, numSamples)
// 	for i, idx := range sampledIndices {
// 		sampledItems[i] = trainset[idx]
// 	}
// 	return sampledItems, sampledIndices
// }

// func (f *FewShotOptimizer) Sample(trainset []DatasetItem, predictions []string, evalResults []MetricResult, maxBootstrappedDemos, maxLabeledDemos int) ([]DatasetItem, []int) {
// 	var bootstrappedSamples []DatasetItem
// 	var labeledSamples []DatasetItem
// 	var bootstrappedIndices []int
// 	var labeledIndices []int

// 	// Sample bootstrapped demos
// 	var successIndices []int
// 	for i, result := range evalResults {
// 		if result.Score >= f.SuccessScore {
// 			successIndices = append(successIndices, i)
// 		}
// 	}

// 	if len(successIndices) > 0 {
// 		bootstrappedIndices = f.RandomSample(successIndices, min(maxBootstrappedDemos, len(successIndices)), false, nil)
// 		bootstrappedSamples = make([]DatasetItem, len(bootstrappedIndices))
// 		for i, idx := range bootstrappedIndices {
// 			bootstrappedSamples[i] = DatasetItem{Inputs: trainset[idx].Inputs, Outputs: predictions[idx]}
// 		}
// 	}

// 	// Sample labeled demos
// 	var failedIndices []int
// 	for i, result := range evalResults {
// 		if result.Score < f.SuccessScore {
// 			failedIndices = append(failedIndices, i)
// 		}
// 	}

// 	if len(failedIndices) > 0 {
// 		weights := make([]float64, len(failedIndices))
// 		for i, idx := range failedIndices {
// 			weights[i] = 1 - evalResults[idx].Score
// 		}
// 		labeledIndices = f.RandomSample(failedIndices, min(maxLabeledDemos, len(failedIndices)), false, weights)
// 		labeledSamples = make([]DatasetItem, len(labeledIndices))
// 		for i, idx := range labeledIndices {
// 			labeledSamples[i] = DatasetItem{Inputs: trainset[idx].Inputs, Outputs: trainset[idx].Outputs}
// 		}
// 	}

// 	return append(bootstrappedSamples, labeledSamples...), append(bootstrappedIndices, labeledIndices...)
// }

// func (f *FewShotOptimizer) RandomSample(dataset []int, numShots int, replace bool, weights []float64) []int {
// 	if len(dataset) == 0 {
// 		return []int{}
// 	}

// 	if !replace && numShots > len(dataset) {
// 		numShots = len(dataset)
// 	}

// 	if weights != nil {
// 		totalWeight := 0.0
// 		for _, weight := range weights {
// 			totalWeight += weight
// 		}
// 		if totalWeight == 0 {
// 			panic("Sum of weights cannot be zero.")
// 		}
// 		for i := range weights {
// 			weights[i] /= totalWeight
// 		}
// 	}

// 	indices := make([]int, numShots)
// 	for i := range indices {
// 		if weights != nil {
// 			indices[i] = f.weightedRandomChoice(dataset, weights)
// 		} else {
// 			indices[i] = dataset[rand.Intn(len(dataset))]
// 		}
// 	}
// 	return indices
// }

// func (f *FewShotOptimizer) weightedRandomChoice(dataset []int, weights []float64) int {
// 	r := rand.Float64()
// 	sum := 0.0
// 	for i, weight := range weights {
// 		sum += weight
// 		if r < sum {
// 			return dataset[i]
// 		}
// 	}
// 	return dataset[len(dataset)-1]
// }

// func contains(slice []int, item int) bool {
// 	for _, v := range slice {
// 		if v == item {
// 			return true
// 		}
// 	}
// 	return false
// }

// // func main() {
// // 	// Example usage
// // 	trainer := NewFewShotTrainer(42, 5, 5, 10, 1.0)
// // 	prompt := Prompt{Messages: []string{"Hello, world!"}}
// // 	trainset := []DatasetItem{{Inputs: "input1", Outputs: "output1"}}
// // 	valset := []DatasetItem{{Inputs: "input2", Outputs: "output2"}}
// // 	updatedPrompt, report := trainer.Train(prompt, trainset, valset)
// // 	_ = updatedPrompt
// // 	_ = report
// // }
