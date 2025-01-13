package optimizer

import (
	"context"
	"fmt"
	"sync"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/program"
	pgp "github.com/showntop/llmack/program"

	// azoai "github.com/showntop/llmack/llm/azure-openai"
	oaic "github.com/showntop/llmack/llm/openai-c"
)

// Evaluator ...
type Evaluator struct {
	model *llm.Instance

	Program       string // prompt or agent
	Parameters    map[string]interface{}
	metric        Metric
	concurrentNum int
	trainSet      []*Example
}

// NewEvaluator ...
func NewEvaluator(concurrentNum int, trainset []*Example, metric Metric) *Evaluator {
	return &Evaluator{
		model: llm.NewInstance(oaic.Name),
		// model:         llm.NewInstance(azoai.Name),
		// Program:       program,
		concurrentNum: concurrentNum,
		metric:        metric,
		trainSet:      trainset,
	}
}

// Score ...
func (e *Evaluator) Score(ctx context.Context, program any) float64 {
	programx, _ := program.(pgp.Program)
	results := make(chan [2]any, len(e.trainSet))
	wg := sync.WaitGroup{}
	wg.Add(len(e.trainSet)) // Add the number of examples to the WaitGroup

	semaphore := make(chan struct{}, e.concurrentNum) // Semaphore to limit concurrency

	for _, ex := range e.trainSet {
		semaphore <- struct{}{} // Acquire a token
		go func(ex *Example) {
			defer wg.Done()                // Mark this goroutine as done when it finishes
			defer func() { <-semaphore }() // Release the token when done

			answer, score := e.evaluateOne(ctx, programx, ex)
			results <- [2]any{answer, score}
		}(ex)
	}

	go func() {
		wg.Wait()      // Wait for all goroutines to finish
		close(results) // Close the results channel
	}()

	scoreSum := 0.0
	// anwers := []string{}
	for score := range results { // Collect scores from the channel
		score := score[1].(float64)
		scoreSum += score
	}
	return scoreSum / float64(len(e.trainSet)) // Calculate the average score
}

func (e *Evaluator) evaluateOne(ctx context.Context, program program.Program, ex *Example) (any, float64) {
	prediction, err := program.Forward(ctx, ex.Inputs())
	if err != nil {
		fmt.Println("evaluate one error ", err)
	}
	score := e.metric(ex, prediction)
	return prediction, score
}
