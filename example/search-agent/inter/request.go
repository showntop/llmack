package inter

// SearchRequest ...
type SearchRequest struct {
	Model      string `json:"model"`
	Source     string `json:"source"`
	Profile    string `json:"profile"`
	IsSearch   bool   `json:"isSearch"`
	IsShadcnUI bool   `json:"isShadcnUI"`
	Messages   []struct {
		ID          string `json:"id"`
		Content     string `json:"content"`
		Role        string `json:"role"`
		Attachments []struct {
			URL string `json:"url"`
		} `json:"attachments"`
		Type string `json:"type"`
	} `json:"messages"`
}
