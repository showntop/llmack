package tool

import (
	"context"
	"fmt"
)

var NilTool = &Tool{
	Name:        "nil",
	Description: "nil tool",
	Kind:        "code",
	Invokex: func(ctx context.Context, args map[string]any) (string, error) {
		return "", fmt.Errorf("not implement")
	},
}
