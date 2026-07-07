package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"dvr-manager/pkg/db"
)

// User 用户实体（密码以哈希形式存储）
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Source       string    `json:"source"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ErrUserNotFound 用户不存在
var ErrUserNotFound = errors.New("用户不存在")

// ErrUserExists 用户已存在
var ErrUserExists = errors.New("用户已存在")

// UserRepository 用户仓库接口
type UserRepository interface {
	GetByUsername(username string) (*User, error)
	GetByID(id int64) (*User, error)
	List() ([]User, error)
	Create(username, passwordHash, role string) (*User, error)
	CreateSSO(username, role, source string) (*User, error)
	UpdatePassword(id int64, passwordHash string) error
	UpdateRole(id int64, role string) error
	Delete(id int64) error
	Count() (int, error)
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository() UserRepository {
	return &userRepository{db: db.GetDB()}
}

func (r *userRepository) scan(row interface {
	Scan(dest ...interface{}) error
}) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.Source, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByUsername 根据用户名查询
func (r *userRepository) GetByUsername(username string) (*User, error) {
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, role, source, created_at, updated_at FROM users WHERE username = ?`,
		username,
	)
	u, err := r.scan(row)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

// GetByID 根据 ID 查询
func (r *userRepository) GetByID(id int64) (*User, error) {
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, role, source, created_at, updated_at FROM users WHERE id = ?`,
		id,
	)
	u, err := r.scan(row)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// List 列出全部用户
func (r *userRepository) List() ([]User, error) {
	rows, err := r.db.Query(
		`SELECT id, username, password_hash, role, source, created_at, updated_at FROM users ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.Source, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Create 新增本地用户
func (r *userRepository) Create(username, passwordHash, role string) (*User, error) {
	res, err := r.db.Exec(
		`INSERT INTO users (username, password_hash, role, source) VALUES (?, ?, ?, 'local')`,
		username, passwordHash, role,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id)
}

// CreateSSO 创建 SSO 用户（无密码，password_hash 留占位禁止本地登录）
func (r *userRepository) CreateSSO(username, role, source string) (*User, error) {
	if source == "" {
		source = "sso"
	}
	res, err := r.db.Exec(
		`INSERT INTO users (username, password_hash, role, source) VALUES (?, '!', ?, ?)`,
		username, role, source,
	)
	if err != nil {
		return nil, ErrUserExists
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id)
}

// UpdatePassword 更新密码哈希
func (r *userRepository) UpdatePassword(id int64, passwordHash string) error {
	_, err := r.db.Exec(
		`UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		passwordHash, id,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

// UpdateRole 更新角色
func (r *userRepository) UpdateRole(id int64, role string) error {
	_, err := r.db.Exec(
		`UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		role, id,
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// Count 用户总数
func (r *userRepository) Count() (int, error) {
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
