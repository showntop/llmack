package inter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/showntop/llmack/example/search-agent/app"
)

// SSEHandler ...
type SSEHandler struct {
	app *app.Application
}

// NewSSEHandler ...
func NewSSEHandler(app *app.Application) *SSEHandler {
	return &SSEHandler{
		app: app,
	}
}

// Search ...
func (h *SSEHandler) Search(c *gin.Context) {
	ctx := c.Request.Context()
	var params SearchRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	if len(params.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty message"})
		return
	}
	events, err := h.app.Search(ctx, app.SearchCommand{
		Query: params.Messages[len(params.Messages)-1].Content,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for event := range events {
		c.SSEvent("data", event)
		c.Writer.Flush()
	}
}
