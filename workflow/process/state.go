package process

import (
	"time"

	wf "github.com/showntop/llmack/workflow"
)

type (
	// State TODO
	State struct {
		Created   time.Time
		Completed *time.Time
		// parent, parent element
		Parent wf.Node

		// current element
		Current wf.Node

		// next elements
		Nexts []wf.Node

		Action string
	}
)
