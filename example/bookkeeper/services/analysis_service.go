package services

import (
	"fmt"
	"time"

	"github.com/bookkeeper-ai/bookkeeper/database"
	"github.com/bookkeeper-ai/bookkeeper/models"
	"gorm.io/gorm"
)

type AnalysisService struct {
	db *gorm.DB
}

func NewAnalysisService() *AnalysisService {
	return &AnalysisService{
		db: database.DB,
	}
}

// GetMonthlyAnalysis 获取月度消费分析
func (s *AnalysisService) GetMonthlyAnalysis(userID uint, year int, month int) (map[string]interface{}, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	var totalExpense float64
	var totalIncome float64
	var categoryStats []struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
		Count    int64   `json:"count"`
	}

	// 计算总支出
	if err := s.db.Model(&models.Transaction{}).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense).Error; err != nil {
		return nil, err
	}

	// 计算总收入
	if err := s.db.Model(&models.Transaction{}).
		Where("user_id = ? AND date BETWEEN ? AND ? AND amount > 0", userID, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalIncome).Error; err != nil {
		return nil, err
	}

	// 按类别统计
	if err := s.db.Model(&models.Transaction{}).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Select("category, SUM(amount) as total, COUNT(*) as count").
		Group("category").
		Scan(&categoryStats).Error; err != nil {
		return nil, err
	}

	// 获取预算信息
	var budget models.Budget
	if err := s.db.Where("user_id = ? AND start_date <= ? AND end_date >= ?", userID, startDate, endDate).
		First(&budget).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 生成消费建议
	suggestions := s.generateSuggestions(totalExpense, budget.Amount, categoryStats)

	return map[string]interface{}{
		"total_expense":  totalExpense,
		"total_income":   totalIncome,
		"category_stats": categoryStats,
		"budget":         budget,
		"suggestions":    suggestions,
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
	}, nil
}

// generateSuggestions 生成消费建议
func (s *AnalysisService) generateSuggestions(totalExpense float64, budget float64, categoryStats []struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Count    int64   `json:"count"`
}) []string {
	var suggestions []string

	// 检查是否超出预算
	if budget > 0 && totalExpense > budget {
		suggestions = append(suggestions, "本月支出已超出预算，建议控制消费")
	}

	// 分析各类别消费
	for _, stat := range categoryStats {
		if stat.Total > budget*0.3 { // 如果某个类别超过预算的30%
			suggestions = append(suggestions,
				fmt.Sprintf("%s类别的支出占比较大，建议关注该领域的消费", stat.Category))
		}
	}

	// 如果没有建议，添加积极反馈
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "您的消费习惯良好，继续保持！")
	}

	return suggestions
}

// GetBudgetRecommendation 获取预算建议
func (s *AnalysisService) GetBudgetRecommendation(userID uint) (map[string]interface{}, error) {
	// 获取过去3个月的消费数据
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)

	var monthlyStats []struct {
		Month  time.Time `json:"month"`
		Total  float64   `json:"total"`
		Income float64   `json:"income"`
	}

	if err := s.db.Model(&models.Transaction{}).
		Where("user_id = ? AND date >= ?", userID, threeMonthsAgo).
		Select("DATE_TRUNC('month', date) as month, SUM(amount) as total, SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END) as income").
		Group("DATE_TRUNC('month', date)").
		Order("month DESC").
		Scan(&monthlyStats).Error; err != nil {
		return nil, err
	}

	// 计算平均月支出和收入
	var totalExpense float64
	var totalIncome float64
	for _, stat := range monthlyStats {
		totalExpense += stat.Total
		totalIncome += stat.Income
	}

	avgExpense := totalExpense / float64(len(monthlyStats))
	avgIncome := totalIncome / float64(len(monthlyStats))

	// 生成预算建议
	recommendedBudget := avgExpense * 1.1 // 建议预算为平均支出的110%

	return map[string]interface{}{
		"average_expense":    avgExpense,
		"average_income":     avgIncome,
		"recommended_budget": recommendedBudget,
		"monthly_statistics": monthlyStats,
		"savings_potential":  avgIncome - avgExpense,
		"budget_adjustment":  recommendedBudget - avgExpense,
	}, nil
}
