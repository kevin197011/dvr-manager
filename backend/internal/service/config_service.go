package service

import (
	"log"
	"sync"

	"dvr-vod-system/internal/config"
	"dvr-vod-system/internal/repository"
)

// ConfigService 配置服务接口
type ConfigService interface {
	GetConfig() (*config.Config, error)
	UpdateConfig(cfg *config.Config) error
	GetDVRServers() []string
	UpdateDVRServers(servers []string) error
	ReloadConfig() error
}

// configService 配置服务实现
type configService struct {
	configRepo repository.ConfigRepository
	dvrRepo    repository.DVRRepository
	mu         sync.RWMutex
	onUpdate   func(*config.Config) // 配置更新回调
}

// NewConfigService 创建新的配置服务
func NewConfigService(configRepo repository.ConfigRepository, dvrRepo repository.DVRRepository) ConfigService {
	return &configService{
		configRepo: configRepo,
		dvrRepo:    dvrRepo,
	}
}

// GetConfig 获取完整配置
func (s *configService) GetConfig() (*config.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return nil, err
	}

	// 从 DVR repository 获取服务器列表
	if s.dvrRepo != nil {
		servers, err := s.dvrRepo.GetAll()
		if err == nil {
			cfg.DVRServers = servers
		}
	}

	return cfg, nil
}

// UpdateConfig 更新完整配置
func (s *configService) UpdateConfig(cfg *config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 保存配置到数据库
	if err := s.configRepo.SaveConfig(cfg); err != nil {
		log.Printf("[ERROR] 保存配置失败: %v", err)
		return err
	}

	// 保存 DVR 服务器列表
	if s.dvrRepo != nil && len(cfg.DVRServers) > 0 {
		if err := s.dvrRepo.Save(cfg.DVRServers); err != nil {
			log.Printf("[ERROR] 保存 DVR 服务器列表失败: %v", err)
			return err
		}
	}

	// 更新全局配置
	config.SetConfig(cfg)

	// 触发更新回调
	if s.onUpdate != nil {
		s.onUpdate(cfg)
	}

	log.Printf("[INFO] 配置已更新")
	return nil
}

// GetDVRServers 获取 DVR 服务器列表
func (s *configService) GetDVRServers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.dvrRepo != nil {
		servers, err := s.dvrRepo.GetAll()
		if err == nil {
			return servers
		}
	}

	return []string{}
}

// UpdateDVRServers 更新 DVR 服务器列表
func (s *configService) UpdateDVRServers(servers []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 保存到数据库
	if s.dvrRepo != nil {
		if err := s.dvrRepo.Save(servers); err != nil {
			log.Printf("[ERROR] 保存到数据库失败: %v", err)
			return err
		}
	}

	// 更新配置中的服务器列表
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		cfg = s.configRepo.GetDefaultConfig()
	}
	cfg.DVRServers = servers

	// 保存完整配置
	if err := s.configRepo.SaveConfig(cfg); err != nil {
		log.Printf("[WARN] 保存配置失败: %v", err)
	}

	// 更新全局配置
	config.SetConfig(cfg)

	// 触发更新回调
	if s.onUpdate != nil {
		s.onUpdate(cfg)
	}

	log.Printf("[INFO] DVR 服务器列表已更新，共 %d 个服务器", len(servers))
	return nil
}

// ReloadConfig 重新加载配置
func (s *configService) ReloadConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return err
	}

	// 从 DVR repository 获取服务器列表
	if s.dvrRepo != nil {
		servers, err := s.dvrRepo.GetAll()
		if err == nil {
			cfg.DVRServers = servers
		}
	}

	// 更新全局配置
	config.SetConfig(cfg)

	// 触发更新回调
	if s.onUpdate != nil {
		s.onUpdate(cfg)
	}

	log.Printf("[INFO] 配置已重新加载，共 %d 个 DVR 服务器", len(cfg.DVRServers))
	return nil
}
