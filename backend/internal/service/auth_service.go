package service

import (
	"errors"
	"os"
)

// User 用户模型
type User struct {
	Username string
	Password string
	Role     string
}

// AuthService 认证服务接口
type AuthService interface {
	Authenticate(username, password string) (*User, error)
	GetUser(username string) (*User, error)
}

// authService 认证服务实现
type authService struct {
	users map[string]*User
}

// NewAuthService 创建新的认证服务
func NewAuthService() AuthService {
	// 从环境变量读取默认管理员账号
	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}

	// 从环境变量读取默认普通用户账号
	userUsername := os.Getenv("USER_USERNAME")
	if userUsername == "" {
		userUsername = "user"
	}
	userPassword := os.Getenv("USER_PASSWORD")
	if userPassword == "" {
		userPassword = "user123"
	}

	users := map[string]*User{
		adminUsername: {
			Username: adminUsername,
			Password: adminPassword,
			Role:     "admin",
		},
		userUsername: {
			Username: userUsername,
			Password: userPassword,
			Role:     "user",
		},
	}

	return &authService{
		users: users,
	}
}

// Authenticate 验证用户名和密码
func (s *authService) Authenticate(username, password string) (*User, error) {
	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("用户不存在")
	}

	if user.Password != password {
		return nil, errors.New("密码错误")
	}

	return user, nil
}

// GetUser 获取用户信息
func (s *authService) GetUser(username string) (*User, error) {
	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("用户不存在")
	}
	return user, nil
}
