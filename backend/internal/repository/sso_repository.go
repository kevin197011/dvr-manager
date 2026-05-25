package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"dvr-vod-system/pkg/db"
)

// SSO 提供商类型常量
const (
	SSOTypeOIDC = "oidc"
	SSOTypeSAML = "saml"
)

// ErrSSOProviderNotFound 提供商不存在
var ErrSSOProviderNotFound = errors.New("sso 提供商不存在")

// SSOProvider SSO 提供商记录
type SSOProvider struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`        // oidc / saml
	Name       string    `json:"name"`        // 展示名
	Enabled    bool      `json:"enabled"`     // 是否启用
	ConfigJSON string    `json:"config_json"` // 协议相关配置 JSON
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SSORepository SSO 提供商仓库接口
type SSORepository interface {
	List() ([]SSOProvider, error)
	ListEnabled() ([]SSOProvider, error)
	GetByID(id int64) (*SSOProvider, error)
	Create(p *SSOProvider) (*SSOProvider, error)
	Update(p *SSOProvider) error
	SetEnabled(id int64, enabled bool) error
	Delete(id int64) error
}

type ssoRepository struct {
	db *sql.DB
}

// NewSSORepository 创建 SSO 仓库
func NewSSORepository() SSORepository {
	return &ssoRepository{db: db.GetDB()}
}

func scanSSO(row interface {
	Scan(dest ...interface{}) error
}) (*SSOProvider, error) {
	var p SSOProvider
	var enabled int
	if err := row.Scan(&p.ID, &p.Type, &p.Name, &enabled, &p.ConfigJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.Enabled = enabled == 1
	return &p, nil
}

// List 列出全部提供商
func (r *ssoRepository) List() ([]SSOProvider, error) {
	rows, err := r.db.Query(
		`SELECT id, type, name, enabled, config_json, created_at, updated_at FROM sso_providers ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list sso providers: %w", err)
	}
	defer rows.Close()

	var list []SSOProvider
	for rows.Next() {
		p, err := scanSSO(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *p)
	}
	return list, rows.Err()
}

// ListEnabled 列出已启用提供商
func (r *ssoRepository) ListEnabled() ([]SSOProvider, error) {
	rows, err := r.db.Query(
		`SELECT id, type, name, enabled, config_json, created_at, updated_at
		 FROM sso_providers WHERE enabled = 1 ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list enabled sso providers: %w", err)
	}
	defer rows.Close()

	var list []SSOProvider
	for rows.Next() {
		p, err := scanSSO(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *p)
	}
	return list, rows.Err()
}

// GetByID 根据 ID 查询
func (r *ssoRepository) GetByID(id int64) (*SSOProvider, error) {
	row := r.db.QueryRow(
		`SELECT id, type, name, enabled, config_json, created_at, updated_at
		 FROM sso_providers WHERE id = ?`, id,
	)
	p, err := scanSSO(row)
	if err == sql.ErrNoRows {
		return nil, ErrSSOProviderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get sso provider: %w", err)
	}
	return p, nil
}

// Create 新增
func (r *ssoRepository) Create(p *SSOProvider) (*SSOProvider, error) {
	enabled := 0
	if p.Enabled {
		enabled = 1
	}
	res, err := r.db.Exec(
		`INSERT INTO sso_providers (type, name, enabled, config_json) VALUES (?, ?, ?, ?)`,
		p.Type, p.Name, enabled, p.ConfigJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("create sso provider: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id)
}

// Update 更新（不允许修改 type）
func (r *ssoRepository) Update(p *SSOProvider) error {
	enabled := 0
	if p.Enabled {
		enabled = 1
	}
	_, err := r.db.Exec(
		`UPDATE sso_providers SET name = ?, enabled = ?, config_json = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		p.Name, enabled, p.ConfigJSON, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update sso provider: %w", err)
	}
	return nil
}

// SetEnabled 启用/停用
func (r *ssoRepository) SetEnabled(id int64, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := r.db.Exec(
		`UPDATE sso_providers SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		v, id,
	)
	return err
}

// Delete 删除
func (r *ssoRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM sso_providers WHERE id = ?`, id)
	return err
}
