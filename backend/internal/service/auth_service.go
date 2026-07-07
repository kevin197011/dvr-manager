package service

import (
	"errors"
	"log"
	"os"
	"strings"

	"dvr-manager/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// User 用户模型（对外返回，不包含密码）
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Source   string `json:"source"`
}

// AuthService 认证服务接口
type AuthService interface {
	Authenticate(username, password string) (*User, error)
	GetUser(username string) (*User, error)
	ChangePassword(username, oldPassword, newPassword string) error

	// SSO 登录使用：根据 (username, source) 查找用户，不存在则自动创建 (role=user)
	FindOrCreateSSOUser(username, source string) (*User, error)

	// 管理员接口
	ListUsers() ([]User, error)
	CreateUser(username, password, role string) (*User, error)
	ResetPassword(id int64, newPassword string) error
	UpdateUserRole(id int64, role string) error
	DeleteUser(id int64) error
	GetUserByID(id int64) (*User, error)
}

type authService struct {
	repo repository.UserRepository
}

// NewAuthService 创建认证服务（数据库存储 + bcrypt）
func NewAuthService(repo repository.UserRepository) AuthService {
	s := &authService{repo: repo}
	s.seedDefaultUsers()
	return s
}

// seedDefaultUsers 在用户表为空时，使用环境变量创建默认账号
func (s *authService) seedDefaultUsers() {
	count, err := s.repo.Count()
	if err != nil {
		log.Printf("[AUTH] count users error: %v", err)
		return
	}
	if count > 0 {
		return
	}

	adminUsername := strings.TrimSpace(os.Getenv("ADMIN_USERNAME"))
	if adminUsername == "" {
		adminUsername = "admin"
	}
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}

	userUsername := strings.TrimSpace(os.Getenv("USER_USERNAME"))
	if userUsername == "" {
		userUsername = "user"
	}
	userPassword := os.Getenv("USER_PASSWORD")
	if userPassword == "" {
		userPassword = "user123"
	}

	if _, err := s.CreateUser(adminUsername, adminPassword, "admin"); err != nil {
		log.Printf("[AUTH] seed admin failed: %v", err)
	} else {
		log.Printf("[AUTH] seeded default admin user: %s", adminUsername)
	}
	if userUsername != adminUsername {
		if _, err := s.CreateUser(userUsername, userPassword, "user"); err != nil {
			log.Printf("[AUTH] seed user failed: %v", err)
		} else {
			log.Printf("[AUTH] seeded default normal user: %s", userUsername)
		}
	}
}

func hashPassword(p string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func toUser(u *repository.User) *User {
	if u == nil {
		return nil
	}
	return &User{ID: u.ID, Username: u.Username, Role: u.Role, Source: u.Source}
}

// Authenticate 验证用户名密码（仅本地用户）
func (s *authService) Authenticate(username, password string) (*User, error) {
	u, err := s.repo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	// SSO 用户禁止本地密码登录
	if u.Source != "" && u.Source != "local" {
		return nil, errors.New("该账号通过 SSO 登录，无法使用密码登录")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return toUser(u), nil
}

// FindOrCreateSSOUser SSO 登录入口：按用户名查找；不存在则以默认角色 user 创建
func (s *authService) FindOrCreateSSOUser(username, source string) (*User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("SSO 用户名为空")
	}
	if source == "" {
		source = "sso"
	}
	u, err := s.repo.GetByUsername(username)
	if err == nil {
		return toUser(u), nil
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	created, err := s.repo.CreateSSO(username, "user", source)
	if err != nil {
		return nil, err
	}
	return toUser(created), nil
}

// GetUser 获取用户
func (s *authService) GetUser(username string) (*User, error) {
	u, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	return toUser(u), nil
}

// GetUserByID 根据 ID 获取
func (s *authService) GetUserByID(id int64) (*User, error) {
	u, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return toUser(u), nil
}

// ChangePassword 修改自己的密码（需校验旧密码）
func (s *authService) ChangePassword(username, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("新密码长度至少 6 位")
	}
	u, err := s.repo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}
	if oldPassword == newPassword {
		return errors.New("新密码与原密码相同")
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(u.ID, hash)
}

// ListUsers 用户列表
func (s *authService) ListUsers() ([]User, error) {
	list, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	out := make([]User, 0, len(list))
	for i := range list {
		out = append(out, *toUser(&list[i]))
	}
	return out, nil
}

// CreateUser 新增用户
func (s *authService) CreateUser(username, password, role string) (*User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}
	if len(password) < 6 {
		return nil, errors.New("密码长度至少 6 位")
	}
	if role != "admin" && role != "user" {
		return nil, errors.New("角色必须为 admin 或 user")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	u, err := s.repo.Create(username, hash, role)
	if err != nil {
		return nil, err
	}
	return toUser(u), nil
}

// ResetPassword 管理员重置某个用户的密码
func (s *authService) ResetPassword(id int64, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("新密码长度至少 6 位")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(id, hash)
}

// UpdateUserRole 修改用户角色
func (s *authService) UpdateUserRole(id int64, role string) error {
	if role != "admin" && role != "user" {
		return errors.New("角色必须为 admin 或 user")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.UpdateRole(id, role)
}

// DeleteUser 删除用户
func (s *authService) DeleteUser(id int64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
