package api

import (
	"net/http"

	"github.com/bookkeeper-ai/bookkeeper/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile 获取用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	user, err := h.userService.GetUserProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// UpdateProfile 更新用户信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateUserProfile(userID, req.Username, req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "个人信息更新成功"})
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdatePassword 更新密码
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userID := c.GetUint("userID")
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdatePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码更新成功"})
}

// UpdateAvatar 更新头像
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID := c.GetUint("userID")
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的头像"})
		return
	}

	if err := h.userService.UpdateAvatar(userID, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像更新成功"})
}
