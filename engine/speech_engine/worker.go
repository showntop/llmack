package engine

// Worker ...
type Worker struct {
	handle Handler
}

// NewWorker ...
func NewWorker(handle Handler) *Worker {
	return &Worker{handle: handle}
}

// LoopHandle ...
func (h *Worker) LoopHandle() error {
	// h.handle()
	return nil
}
