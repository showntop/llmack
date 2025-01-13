package workflow

// Edge TODO
type Edge struct {
	ID          string `json:"id"` // 线段id
	Name        string `json:"name,omitempty"`
	Source      string `json:"source"`
	Target      string `json:"target"`
	ExpressName string `json:"expr_name,omitempty"`
	Express     string `json:"expr_content,omitempty"`
	// Metadata meta `json:"metadata"`
}
