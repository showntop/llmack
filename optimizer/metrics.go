package optimizer

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/showntop/llmack/program"
	"golang.org/x/text/unicode/norm"
)

// type Metric interface {
// 	Calculate(expected, actual []string) float64
// }

// type Accuracy struct{}

// func (a Accuracy) Calculate(expected, actual []string) float64 {
// 	var correct float64
// 	for i := range expected {
// 		if expected[i] == actual[i] {
// 			correct++
// 		}
// 	}
// 	return correct / float64(len(expected))
// }

// Metric ...
type Metric func(*Example, any) float64

// AccuracyMatch checks if outputs exactly match
func AccuracyMatch(output, expected any) float64 {
	if output == expected {
		return 1.0
	}
	return 0.0
}

// StringSimilarity calculates string similarity using Levenshtein distance
func StringSimilarity(output, expected any) float64 {
	str1, ok1 := output.(string)
	str2, ok2 := expected.(string)
	if !ok1 || !ok2 {
		return 0.0
	}

	// Normalize strings
	str1 = strings.TrimSpace(strings.ToLower(str1))
	str2 = strings.TrimSpace(strings.ToLower(str2))

	if str1 == str2 {
		return 1.0
	}

	// Calculate Levenshtein distance
	d := levenshteinDistance(str1, str2)
	maxLen := math.Max(float64(len(str1)), float64(len(str2)))

	return 1.0 - float64(d)/maxLen
}

// F1Score calculates F1 score for classification tasks
func F1Score(output, expected interface{}) float64 {
	pred, ok1 := output.(map[string]interface{})
	true, ok2 := expected.(map[string]interface{})
	if !ok1 || !ok2 {
		return 0.0
	}

	var tp, fp, fn float64
	for k, v := range pred {
		if trueVal, exists := true[k]; exists {
			if v == trueVal {
				tp++
			} else {
				fp++
			}
		} else {
			fp++
		}
	}

	for k := range true {
		if _, exists := pred[k]; !exists {
			fn++
		}
	}

	precision := tp / (tp + fp)
	recall := tp / (tp + fn)

	if precision+recall == 0 {
		return 0.0
	}

	return 2 * (precision * recall) / (precision + recall)
}

// Helper function to calculate Levenshtein distance
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			if s1[i-1] == s2[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = min(
					matrix[i-1][j]+1,   // deletion
					matrix[i][j-1]+1,   // insertion
					matrix[i-1][j-1]+1, // substitution
				)
			}
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(nums ...int) int {
	result := nums[0]
	for _, num := range nums[1:] {
		if num < result {
			result = num
		}
	}
	return result
}

// EM calculates the exact match score between the prediction and the list of answers.
func EM(prediction string, answersList []string) float64 {
	maxScore := 0.0
	for _, ans := range answersList {
		score := emScore(prediction, ans)
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}

// F1 calculates the F1 score between the prediction and the list of answers.
func F1(prediction string, answersList []string) float64 {
	maxScore := 0.0
	for _, ans := range answersList {
		score := f1Score(prediction, ans)
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}

// HotPotF1 calculates the HotPot F1 score between the prediction and the list of answers.
func HotPotF1(prediction string, answersList []string) float64 {
	maxScore := 0.0
	for _, ans := range answersList {
		score := hotpotF1Score(prediction, ans)
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}

// normalizeText normalizes the input text by removing articles, punctuation, and extra whitespace.
func normalizeText(s string) string {
	s = norm.NFD.String(s)
	s = removeArticles(s)
	s = removePunc(s)
	s = strings.ToLower(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// removeArticles removes articles (a, an, the) from the text.
func removeArticles(text string) string {
	return regexp.MustCompile(`\b(a|an|the)\b`).ReplaceAllString(text, " ")
}

// removePunc removes punctuation from the text.
func removePunc(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) {
			return -1
		}
		return r
	}, text)
}

// emScore calculates the exact match score between the prediction and the ground truth.
func emScore(prediction, groundTruth string) float64 {
	if normalizeText(prediction) == normalizeText(groundTruth) {
		return 1.0
	}
	return 0.0
}

// f1Score calculates the F1 score between the prediction and the ground truth.
func f1Score(prediction, groundTruth string) float64 {
	predictionTokens := strings.Fields(normalizeText(prediction))
	groundTruthTokens := strings.Fields(normalizeText(groundTruth))

	common := intersection(predictionTokens, groundTruthTokens)
	numSame := len(common)

	if len(predictionTokens) == 0 && len(groundTruthTokens) == 0 {
		fmt.Println("\n#> F1 Metric: Rare edge case of len(predictionTokens) == len(groundTruthTokens) == 0.\n")
	}

	if numSame == 0 {
		return 0.0
	}

	precision := float64(numSame) / float64(len(predictionTokens))
	recall := float64(numSame) / float64(len(groundTruthTokens))
	f1 := (2 * precision * recall) / (precision + recall)

	return f1
}

// hotpotF1Score calculates the HotPot F1 score between the prediction and the ground truth.
func hotpotF1Score(prediction, groundTruth string) float64 {
	normalizedPrediction := normalizeText(prediction)
	normalizedGroundTruth := normalizeText(groundTruth)

	if contains([]string{"yes", "no", "noanswer"}, normalizedPrediction) && normalizedPrediction != normalizedGroundTruth {
		return 0.0
	}
	if contains([]string{"yes", "no", "noanswer"}, normalizedGroundTruth) && normalizedPrediction != normalizedGroundTruth {
		return 0.0
	}

	predictionTokens := strings.Fields(normalizedPrediction)
	groundTruthTokens := strings.Fields(normalizedGroundTruth)
	common := intersection(predictionTokens, groundTruthTokens)
	numSame := len(common)
	if numSame == 0 {
		return 0.0
	}
	precision := float64(numSame) / float64(len(predictionTokens))
	recall := float64(numSame) / float64(len(groundTruthTokens))
	f1 := (2 * precision * recall) / (precision + recall)
	return f1
}

// precisionScore calculates the precision score between the prediction and the ground truth.
func precisionScore(prediction, groundTruth string) float64 {
	predictionTokens := strings.Fields(normalizeText(prediction))
	groundTruthTokens := strings.Fields(normalizeText(groundTruth))

	common := intersection(predictionTokens, groundTruthTokens)
	numSame := len(common)

	if len(predictionTokens) == 0 && len(groundTruthTokens) == 0 {
		fmt.Println("\n#> F1 Metric: Rare edge case of len(predictionTokens) == len(groundTruthTokens) == 0.\n")
	}

	if numSame == 0 {
		return 0.0
	}

	precision := float64(numSame) / float64(len(predictionTokens))
	return precision
}

// intersection returns the common elements between two slices of strings.
func intersection(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}

	var result []string
	for _, item := range b {
		if _, ok := m[item]; ok {
			result = append(result, item)
		}
	}
	return result
}

// contains checks if a string is present in a slice of strings.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type semanticF1 struct {
	ctx context.Context
	cot program.Program
}

// NewSemanticF1 ...
func NewSemanticF1(ctx context.Context) *semanticF1 {
	sf1 := &semanticF1{ctx: ctx}
	sf1.cot = program.COT().
		WithInstruction(`Compare a system's response to the ground truth to compute recall and precision of key ideas.
		You will first enumerate key ideas in each response, discuss their overlap, and then report recall and precision.`).
		WithInputField("question", "question").
		WithInputField("ground_truth", "ground_truth").
		WithInputField("system_response", "system_response").
		WithOutputField("ground_truth_key_ideas", "enumeration of key ideas in the ground truth").
		WithOutputField("system_response_key_ideas", "enumeration of key ideas in the system response").
		WithOutputField("discussion", "discussion of the overlap between ground truth and system response").
		WithOutputField("recall", "fraction (out of 1.0) of ground truth covered by the system response", "", reflect.Float64).
		WithOutputField("precision", "fraction (out of 1.0) of system response covered by the ground truth", "", reflect.Float64)
	return sf1
}

// Metric ...
func (s semanticF1) Metric(ex *Example, prediction any) float64 {
	result, err := s.cot.Forward(s.ctx, map[string]any{
		"question":        ex.Get("question"),
		"ground_truth":    ex.Get("ground_truth"),
		"system_response": prediction.(map[string]any)["response"],
	})
	if err != nil {
		// panic(err)
		return 0.0
	}
	fmt.Println(result)
	precision, _ := strconv.ParseFloat(fmt.Sprint(result.(map[string]any)["precision"]), 64)
	recall, _ := strconv.ParseFloat(fmt.Sprint(result.(map[string]any)["recall"]), 64)

	return s.f1Score(precision, recall)
}

// SemanticF1 ...
func SemanticF1() Metric {
	return (NewSemanticF1(context.Background())).Metric
}

// f1Score calculates the F1 score given precision and recall
func (s semanticF1) f1Score(precision, recall float64) float64 {
	// Clamp precision and recall to the range [0.0, 1.0]
	precision = math.Max(0.0, math.Min(1.0, precision))
	recall = math.Max(0.0, math.Min(1.0, recall))

	// Calculate F1 score
	if precision+recall == 0 {
		return 0.0
	}
	return 2 * (precision * recall) / (precision + recall)
}
