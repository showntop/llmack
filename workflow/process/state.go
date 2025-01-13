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

		// element error (if any)
		// Err error

		// input variables that were sent to resume the session
		Input *wf.Vars

		// scope
		Scope *wf.Vars

		// element execution results
		Outputs *wf.Vars

		Action string
	}
)
