package tool

import (
	"context"
	"fmt"
)

var NilTool = &Tool{
	Name:        "nil",
	Description: "nil tool",
	Kind:        "code",
	invoke: func(ctx context.Context, args string) (string, error) {
		return "", fmt.Errorf("not implement")
	},
}
