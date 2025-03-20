package app

// SearchEvent ...
type SearchEvent struct {
	Status  string `json:"status"`
	Data    string `json:"data"`
	Answer  string `json:"answer"`
	Sources any    `json:"sources"`
	Related any    `json:"related"`
	Images  any    `json:"images"`
	Videos  any    `json:"videos"`
}

// SearchEventStream ...
type SearchEventStream chan *SearchEvent

// NewSearchEventStream ...
func NewSearchEventStream() SearchEventStream {
	ch := make(chan *SearchEvent, 1)
	return SearchEventStream(ch)
}

// Push ...
func (s SearchEventStream) Push(e *SearchEvent) {
	s <- e
}

// Close ...
func (s SearchEventStream) Close() {
	close(s)
}

// Next ...
func (s SearchEventStream) Next() *SearchEvent {
	return <-s
}
