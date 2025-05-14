package services

import (
	"errors"
	"time"

	"github.com/bookkeeper-ai/bookkeeper/database"
	"github.com/bookkeeper-ai/bookkeeper/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{
		db: database.DB,
	}
}

// Register 用户注册
func (s *AuthService) Register(username, email, password string) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (string, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", errors.New("用户不存在")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("密码错误")
	}

	// 更新最后登录时间
	s.db.Model(&user).Update("last_login_at", time.Now())

	// 生成 JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7天过期
	})

	// 签名 token
	tokenString, err := token.SignedString([]byte("your-secret-key")) // TODO: 从配置文件读取密钥
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证 token
func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil // TODO: 从配置文件读取密钥
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, errors.New("无效的 token")
}
