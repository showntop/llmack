package optimizer

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/showntop/llmack/pkg/structx"
	"github.com/showntop/llmack/program"
)

// Candidate represents a candidate for optimization.
type Candidate struct {
	Program     program.Program
	Score       float64
	Instruction string
	Prefix      string
	Depth       int
}

// CoproOptimizer represents the teleprompter for optimizing signatures.
type CoproOptimizer struct {
	metric          Metric
	Breadth         int
	Depth           int
	InitTemperature float64
	TrackStats      bool
}

// NewCoproOptimizer creates a new CoproOptimizer.
func NewCoproOptimizer(breadth, depth int, metric Metric) *CoproOptimizer {
	return &CoproOptimizer{
		metric:  metric,
		Breadth: breadth,
		Depth:   depth,
	}
}

// Optimize optimizes the program.
func (co *CoproOptimizer) Optimize(ctx context.Context, student program.Program, trainset []*Example) program.Program {
	evaluate := NewEvaluator(3, trainset, co.metric)

	var candidates []map[string]any
	err := program.Predictor().
		WithInstruction(`You are an instruction optimizer for large language models.
		I will give you a few of fields (inputs and outputs) in English.
		Your task is to propose `+strconv.Itoa(co.Breadth)+` instructions that will lead a good language model to perform the task well.
		Don't be afraid to be creative.`).
		WithInputField(
			"initial_instruction", "The initial instructions before optimization").
		WithOutputField(
			"proposed_instruction", "The improved instructions for the language model").
		WithOutputField(
			"proposed_prefix_for_output_field", "The string at the end of the prompt, which will help the model start solving the task").
		ChatWith(map[string]any{
			"initial_instruction": "Answer the question and give the reasoning for the same."}).
		Result(ctx, &candidates)
	if err != nil {
		panic(err)
	}

	var evaluateCandidates = structx.Set[string]()
	for i := 0; i < co.Depth; i++ { // for each cadidate
		fmt.Println("================================ccoo======================================")
		fmt.Println("Depth: ", i, "/", co.Depth)
		fmt.Println("Candidates: ", candidates)
		fmt.Println("================================ccoo======================================")
		for j := range candidates {
			var lastOutput *program.Field
			for _, v := range student.Prompt().OutputFields {
				lastOutput = v
			}
			proposedInstruction := candidates[j]["proposed_instruction"].(string)
			proposedPrefixForOutputField := candidates[j]["proposed_prefix_for_output_field"].(string)
			student.Update(
				program.WithInstruction(proposedInstruction),
				program.WithOutput(lastOutput.Name, lastOutput.Description, proposedPrefixForOutputField),
			)
			score := evaluate.Score(ctx, student)

			replace := true
			ok := evaluateCandidates.Exist(proposedInstruction + proposedPrefixForOutputField)
			if ok && score <= evaluateCandidates.Get(proposedInstruction + proposedPrefixForOutputField).(any).(map[string]any)["score"].(float64) { // replace if better
				replace = false
			}
			if replace {
				evaluateCandidates.Set(proposedInstruction+proposedPrefixForOutputField, map[string]any{
					"score":       score,
					"program":     student,
					"instruction": proposedInstruction,
					"prefix":      proposedPrefixForOutputField,
					"depth":       i,
				})
			}
		}
		fmt.Println("================================Candidates======================================")
		fmt.Println("Depth: ", i, "/", co.Depth)
		fmt.Println("Candidates: ", evaluateCandidates.Values())
		fmt.Println("================================Candidates======================================")
		// 爬升
		values := evaluateCandidates.Values()
		sort.Slice(values, func(i, j int) bool {
			return values[i].(map[string]any)["score"].(float64) > values[j].(map[string]any)["score"].(float64)
		})
		count := len(values)
		if count > co.Breadth {
			count = co.Breadth
		}
		attempts := []string{}
		for i := 0; i < count; i++ {
			xxx := values[i].(map[string]any)
			attempts = append(attempts, fmt.Sprintf("Instruction #%d: %s", i, xxx["instruction"].(string)))
			attempts = append(attempts, fmt.Sprintf("Prefix #%d: %s", i, xxx["prefix"].(string)))
			attempts = append(attempts, fmt.Sprintf("Resulting Score #%d: %f", i, xxx["score"].(float64)))
		}

		candidates = candidates[:0] // reset candidates

		err := program.Predictor().
			WithInstruction(`You are an instruction optimizer for large language models. I will give some task instructions I've tried, along with their corresponding validation scores. The instructions are arranged in increasing order based on their scores, where higher scores indicate better quality.

    		Your task is to propose a new instruction that will lead a good language model to perform the task even better. Don't be afraid to be creative.`).
			WithInputFields(map[string]string{"attempted_instructions": "attempted_instructions"}).
			WithOutputField("proposed_instruction", "The improved instructions for the language model").
			WithOutputField("proposed_prefix_for_output_field", "The string at the end of the prompt, which will help the model start solving the task").
			ChatWith(map[string]any{"attempted_instructions": attempts}).
			Result(ctx, &candidates)
		if err != nil {
			panic(err)
		}
	}

	finalCandidates := evaluateCandidates.Values()
	sort.Slice(finalCandidates, func(i, j int) bool {
		item := finalCandidates[i].(map[string]any)
		item2 := finalCandidates[j].(map[string]any)
		return item["score"].(float64) > item2["score"].(float64)
	})

	instruction := finalCandidates[0].(map[string]any)["instruction"].(string)
	var lastOutput *program.Field
	for _, v := range student.Prompt().OutputFields {
		lastOutput = v
	}
	proposedPrefixForOutputField := finalCandidates[0].(map[string]any)["prefix"].(string)
	student.Update(
		program.WithInstruction(instruction),
		program.WithOutput(lastOutput.Name, lastOutput.Description, proposedPrefixForOutputField),
	)

	return student
}
