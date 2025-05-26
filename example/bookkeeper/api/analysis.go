package api

import (
	"net/http"
	"strconv"

	"github.com/bookkeeper-ai/bookkeeper/services"
	"github.com/gin-gonic/gin"
)

type AnalysisHandler struct {
	analysisService *services.AnalysisService
}

func NewAnalysisHandler(analysisService *services.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService: analysisService,
	}
}

// GetMonthlyAnalysis 获取月度消费分析
func (h *AnalysisHandler) GetMonthlyAnalysis(c *gin.Context) {
	userID := uint(1) // TODO: 从认证中获取用户ID
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	month, _ := strconv.Atoi(c.DefaultQuery("month", "1"))

	analysis, err := h.analysisService.GetMonthlyAnalysis(userID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetBudgetRecommendation 获取预算建议
func (h *AnalysisHandler) GetBudgetRecommendation(c *gin.Context) {
	userID := uint(1) // TODO: 从认证中获取用户ID

	recommendation, err := h.analysisService.GetBudgetRecommendation(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recommendation)
}
