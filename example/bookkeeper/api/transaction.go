package api

import (
	"net/http"
	"time"

	"github.com/bookkeeper-ai/bookkeeper/database"
	"github.com/bookkeeper-ai/bookkeeper/models"
	"github.com/bookkeeper-ai/bookkeeper/services"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	imageService *services.ImageService
}

func NewTransactionHandler(imageService *services.ImageService) *TransactionHandler {
	return &TransactionHandler{
		imageService: imageService,
	}
}

// CreateTransaction 创建新的交易记录
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// GetTransactions 获取交易记录列表
func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	var transactions []models.Transaction
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := database.DB.Model(&models.Transaction{})

	if startDate != "" && endDate != "" {
		start, _ := time.Parse("2006-01-02", startDate)
		end, _ := time.Parse("2006-01-02", endDate)
		query = query.Where("date BETWEEN ? AND ?", start, end)
	}

	if err := query.Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetTransactionAnalysis 获取交易分析
func (h *TransactionHandler) GetTransactionAnalysis(c *gin.Context) {
	var result []struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
	}

	if err := database.DB.Model(&models.Transaction{}).
		Select("category, sum(amount) as total").
		Group("category").
		Scan(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UploadReceipt 上传并分析消费小票
func (h *TransactionHandler) UploadReceipt(c *gin.Context) {
	file, err := c.FormFile("receipt")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
		return
	}

	filepath, result, err := h.imageService.UploadAndAnalyze(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建交易记录
	transaction := models.Transaction{
		Amount:      result["amount"].(float64),
		Category:    result["category"].(string),
		Description: result["merchant"].(string),
		Date:        result["date"].(time.Time),
		ImageURL:    filepath,
		UserID:      1, // TODO: 从认证中获取用户 ID
	}

	if err := database.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction": transaction,
		"analysis":    result,
	})
}
