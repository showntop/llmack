package api

import (
	"net/http"

	"github.com/bookkeeper-ai/bookkeeper/services"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

type ChatRequest struct {
	Message string   `json:"message" binding:"required"`
	Medias  []string `json:"medias"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.chatService.SendMessage(c.Request.Context(), req.Message, req.Medias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response: response,
	})
}
