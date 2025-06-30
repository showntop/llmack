package program

import (
	"context"
	"testing"

	"github.com/showntop/llmack/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLLMModel mocks the LLM model interface
type MockLLMModel struct {
	mock.Mock
}

func (m *MockLLMModel) Invoke(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
	args := m.Called(ctx, messages, opts)
	return args.Get(0).(*llm.Response), args.Error(1)
}

func TestFunCall(t *testing.T) {
	p := FunCall()
	assert.NotNil(t, p)
	assert.Equal(t, "funcall", p.Mode)
	assert.NotNil(t, p.invoker)
}

func TestFuncall_Invokex(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		setupMocks  func(*llm.Instance)
		wantError   bool
		errorString string
	}{
		{
			name:  "successful invocation without tools",
			query: "test query",
			setupMocks: func(m *llm.Instance) {
				m.Invoke(context.Background(), []llm.Message{})
			},
			wantError: false,
		},
		{
			name:  "max iteration reached",
			query: "test query that needs many iterations",
			setupMocks: func(m *llm.Instance) {
				m.Invoke(context.Background(), []llm.Message{})
			},
			wantError:   true,
			errorString: "failed to invoke query until max iteration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockModel := llm.New(llm.MockLLMModelName)
			tt.setupMocks(mockModel)

			p := FunCall()
			p.model = mockModel

			result := p.invoker.Invoke(context.Background(), []llm.Message{}, tt.query, nil)

			if tt.wantError {
				assert.Error(t, result.reponse.err)
				if tt.errorString != "" {
					assert.Equal(t, tt.errorString, result.reponse.err.Error())
				}
			} else {
				assert.NoError(t, result.reponse.err)
			}
		})
	}
}

func TestFuncall_BuildTools(t *testing.T) {
	rp := &funcall{predictor: NewPredictor()}

	tools := []any{"calculator", "weather"}
	messageTools := rp.buildTools(tools...)

	assert.Len(t, messageTools, 2)
	assert.Equal(t, "function", messageTools[0].Type)

	// Verify tool properties
	for _, tool := range messageTools {
		assert.NotEmpty(t, tool.Function.Name)
		assert.NotEmpty(t, tool.Function.Description)
		assert.NotNil(t, tool.Function.Parameters)

		// Type assert Parameters to map[string]any first
		params, ok := tool.Function.Parameters.(map[string]any)
		assert.True(t, ok, "Parameters should be map[string]any")

		properties := params["properties"].(map[string]any)
		assert.NotNil(t, properties)

		required := params["required"].([]string)
		assert.NotNil(t, required)
	}
}

func TestFuncall_InvokeTools(t *testing.T) {
	ctx := context.Background()
	rp := &funcall{predictor: NewPredictor()}

	toolCalls := []*llm.ToolCall{
		{
			ID: "test-id-1",
			Function: llm.ToolCallFunction{
				Name:      "calculator",
				Arguments: `{"operation": "add", "numbers": [1, 2]}`,
			},
		},
	}

	results, err := rp.invokeTools(ctx, toolCalls)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Contains(t, results, "test-id-1")
}
