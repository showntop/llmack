package services

import (
	"errors"
	"mime/multipart"

	"github.com/bookkeeper-ai/bookkeeper/database"
	"github.com/bookkeeper-ai/bookkeeper/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{
		db: database.DB,
	}
}

// GetUserProfile 获取用户信息
func (s *UserService) GetUserProfile(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserProfile 更新用户信息
func (s *UserService) UpdateUserProfile(userID uint, username, email string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	// 检查用户名是否已被其他用户使用
	if username != user.Username {
		var existingUser models.User
		if err := s.db.Where("username = ? AND id != ?", username, userID).First(&existingUser).Error; err == nil {
			return errors.New("用户名已被使用")
		}
		user.Username = username
	}

	// 检查邮箱是否已被其他用户使用
	if email != user.Email {
		var existingUser models.User
		if err := s.db.Where("email = ? AND id != ?", email, userID).First(&existingUser).Error; err == nil {
			return errors.New("邮箱已被使用")
		}
		user.Email = email
	}

	return s.db.Save(&user).Error
}

// UpdatePassword 更新密码
func (s *UserService) UpdatePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.db.Save(&user).Error
}

// UpdateAvatar 更新头像
func (s *UserService) UpdateAvatar(userID uint, file *multipart.FileHeader) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	// TODO: 实现文件上传和存储逻辑
	// 1. 验证文件类型和大小
	// 2. 生成唯一文件名
	// 3. 保存文件
	// 4. 更新用户头像URL

	return nil
}
