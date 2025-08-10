package services

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"sync"

	"npm-console/internal/core"
	"npm-console/internal/managers"
	"npm-console/pkg/logger"
)

// ConfigService implements configuration management functionality
type ConfigService struct {
	factory *managers.ManagerFactory
	logger  *logger.Logger
}

// NewConfigService creates a new config service
func NewConfigService() *ConfigService {
	return &ConfigService{
		factory: managers.GetGlobalFactory(),
		logger:  logger.GetDefault().WithField("service", "config"),
	}
}

// GetAllConfigs returns configuration for all available package managers
func (s *ConfigService) GetAllConfigs(ctx context.Context) ([]core.Config, error) {
	availableManagers := s.factory.GetAvailableManagers(ctx)
	
	var configs []core.Config
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errors []error

	// Get configs concurrently from all managers
	for name, manager := range availableManagers {
		wg.Add(1)
		go func(name string, mgr core.PackageManager) {
			defer wg.Done()
			
			config, err := mgr.GetConfig(ctx)
			if err != nil {
				s.logger.WithError(err).WithField("manager", name).Warn("Failed to get config")
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to get config for %s: %w", name, err))
				mu.Unlock()
				return
			}
			
			mu.Lock()
			configs = append(configs, *config)
			mu.Unlock()
		}(name, manager)
	}
	
	wg.Wait()
	
	// Sort by manager name for consistent output
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Manager < configs[j].Manager
	})
	
	// Log any errors but don't fail the entire operation
	if len(errors) > 0 {
		for _, err := range errors {
			s.logger.WithError(err).Warn("Config retrieval error")
		}
	}
	
	return configs, nil
}

// GetConfig returns configuration for a specific package manager
func (s *ConfigService) GetConfig(ctx context.Context, managerName string) (*core.Config, error) {
	manager, err := s.factory.GetManager(managerName)
	if err != nil {
		return nil, err
	}
	
	if !manager.IsAvailable(ctx) {
		return nil, core.NewManagerError(managerName, "get config", core.ErrManagerNotAvailable)
	}
	
	return manager.GetConfig(ctx)
}

// SetRegistry sets the registry URL for a specific manager
func (s *ConfigService) SetRegistry(ctx context.Context, managerName string, registryURL string) error {
	if err := s.ValidateRegistryURL(registryURL); err != nil {
		return err
	}
	
	manager, err := s.factory.GetManager(managerName)
	if err != nil {
		return err
	}
	
	if !manager.IsAvailable(ctx) {
		return core.NewManagerError(managerName, "set registry", core.ErrManagerNotAvailable)
	}
	
	err = manager.SetRegistry(ctx, registryURL)
	if err != nil {
		return err
	}
	
	s.logger.WithField("manager", managerName).WithField("registry", registryURL).Info("Registry updated")
	return nil
}

// SetRegistryForAll sets the registry URL for all available managers
func (s *ConfigService) SetRegistryForAll(ctx context.Context, registryURL string) error {
	if err := s.ValidateRegistryURL(registryURL); err != nil {
		return err
	}
	
	availableManagers := s.factory.GetAvailableManagers(ctx)
	
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errors []error
	var successCount int

	// Set registry concurrently for all managers
	for name, manager := range availableManagers {
		wg.Add(1)
		go func(name string, mgr core.PackageManager) {
			defer wg.Done()
			
			err := mgr.SetRegistry(ctx, registryURL)
			if err != nil {
				s.logger.WithError(err).WithField("manager", name).Error("Failed to set registry")
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to set registry for %s: %w", name, err))
				mu.Unlock()
				return
			}
			
			mu.Lock()
			successCount++
			mu.Unlock()
			
			s.logger.WithField("manager", name).WithField("registry", registryURL).Info("Registry updated")
		}(name, manager)
	}
	
	wg.Wait()
	
	s.logger.WithField("success_count", successCount).WithField("total_managers", len(availableManagers)).Info("Registry update completed")
	
	// Return error if any registry setting failed
	if len(errors) > 0 {
		return fmt.Errorf("failed to set registry for some managers: %v", errors)
	}
	
	return nil
}

// SetProxy sets the proxy configuration for a specific manager
func (s *ConfigService) SetProxy(ctx context.Context, managerName string, proxyURL string) error {
	if proxyURL != "" {
		if err := s.ValidateProxyURL(proxyURL); err != nil {
			return err
		}
	}
	
	manager, err := s.factory.GetManager(managerName)
	if err != nil {
		return err
	}
	
	if !manager.IsAvailable(ctx) {
		return core.NewManagerError(managerName, "set proxy", core.ErrManagerNotAvailable)
	}
	
	err = manager.SetProxy(ctx, proxyURL)
	if err != nil {
		return err
	}
	
	if proxyURL == "" {
		s.logger.WithField("manager", managerName).Info("Proxy removed")
	} else {
		s.logger.WithField("manager", managerName).WithField("proxy", proxyURL).Info("Proxy updated")
	}
	
	return nil
}

// SetProxyForAll sets the proxy configuration for all available managers
func (s *ConfigService) SetProxyForAll(ctx context.Context, proxyURL string) error {
	if proxyURL != "" {
		if err := s.ValidateProxyURL(proxyURL); err != nil {
			return err
		}
	}
	
	availableManagers := s.factory.GetAvailableManagers(ctx)
	
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errors []error
	var successCount int

	// Set proxy concurrently for all managers
	for name, manager := range availableManagers {
		wg.Add(1)
		go func(name string, mgr core.PackageManager) {
			defer wg.Done()
			
			err := mgr.SetProxy(ctx, proxyURL)
			if err != nil {
				s.logger.WithError(err).WithField("manager", name).Error("Failed to set proxy")
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to set proxy for %s: %w", name, err))
				mu.Unlock()
				return
			}
			
			mu.Lock()
			successCount++
			mu.Unlock()
			
			if proxyURL == "" {
				s.logger.WithField("manager", name).Info("Proxy removed")
			} else {
				s.logger.WithField("manager", name).WithField("proxy", proxyURL).Info("Proxy updated")
			}
		}(name, manager)
	}
	
	wg.Wait()
	
	s.logger.WithField("success_count", successCount).WithField("total_managers", len(availableManagers)).Info("Proxy update completed")
	
	// Return error if any proxy setting failed
	if len(errors) > 0 {
		return fmt.Errorf("failed to set proxy for some managers: %v", errors)
	}
	
	return nil
}

// TestRegistry tests connectivity to a registry URL
func (s *ConfigService) TestRegistry(ctx context.Context, managerName string, registryURL string) error {
	if err := s.ValidateRegistryURL(registryURL); err != nil {
		return err
	}
	
	// For now, just validate the URL format
	// In the future, this could make an actual HTTP request to test connectivity
	s.logger.WithField("manager", managerName).WithField("registry", registryURL).Info("Registry test passed (URL validation)")
	
	return nil
}

// TestProxy tests proxy connectivity
func (s *ConfigService) TestProxy(ctx context.Context, managerName string, proxyURL string) error {
	if err := s.ValidateProxyURL(proxyURL); err != nil {
		return err
	}
	
	// For now, just validate the URL format
	// In the future, this could make an actual HTTP request through the proxy
	s.logger.WithField("manager", managerName).WithField("proxy", proxyURL).Info("Proxy test passed (URL validation)")
	
	return nil
}

// GetConfigSummary returns a summary of configuration across all managers
func (s *ConfigService) GetConfigSummary(ctx context.Context) (*ConfigSummary, error) {
	configs, err := s.GetAllConfigs(ctx)
	if err != nil {
		return nil, err
	}
	
	summary := &ConfigSummary{
		Managers:   make(map[string]ConfigManagerSummary),
		Registries: make(map[string][]string),
		Proxies:    make(map[string][]string),
	}
	
	for _, config := range configs {
		summary.ManagerCount++
		
		managerSummary := ConfigManagerSummary{
			Registry: config.Registry,
			Proxy:    config.Proxy,
			Settings: len(config.Settings),
		}
		summary.Managers[config.Manager] = managerSummary
		
		// Group managers by registry
		if config.Registry != "" {
			summary.Registries[config.Registry] = append(summary.Registries[config.Registry], config.Manager)
		}
		
		// Group managers by proxy
		if config.Proxy != "" {
			summary.Proxies[config.Proxy] = append(summary.Proxies[config.Proxy], config.Manager)
		}
	}
	
	return summary, nil
}

// ValidateRegistryURL validates a registry URL
func (s *ConfigService) ValidateRegistryURL(registryURL string) error {
	if registryURL == "" {
		return core.NewValidationError("registry", registryURL, "registry URL cannot be empty")
	}
	
	parsedURL, err := url.Parse(registryURL)
	if err != nil {
		return core.NewValidationError("registry", registryURL, "invalid URL format")
	}
	
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return core.NewValidationError("registry", registryURL, "registry URL must use http or https scheme")
	}
	
	if parsedURL.Host == "" {
		return core.NewValidationError("registry", registryURL, "registry URL must have a host")
	}
	
	return nil
}

// ValidateProxyURL validates a proxy URL
func (s *ConfigService) ValidateProxyURL(proxyURL string) error {
	if proxyURL == "" {
		return nil // Empty proxy URL is valid (means no proxy)
	}
	
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return core.NewValidationError("proxy", proxyURL, "invalid URL format")
	}
	
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return core.NewValidationError("proxy", proxyURL, "proxy URL must use http or https scheme")
	}
	
	if parsedURL.Host == "" {
		return core.NewValidationError("proxy", proxyURL, "proxy URL must have a host")
	}
	
	return nil
}

// ValidateManagerName validates if a manager name is valid and available
func (s *ConfigService) ValidateManagerName(ctx context.Context, managerName string) error {
	if err := s.factory.ValidateManager(managerName); err != nil {
		return err
	}
	
	if !s.factory.IsManagerAvailable(ctx, managerName) {
		return core.NewManagerError(managerName, "validate", core.ErrManagerNotAvailable)
	}
	
	return nil
}

// ConfigSummary represents a summary of configuration across all managers
type ConfigSummary struct {
	ManagerCount int                            `json:"manager_count"`
	Managers     map[string]ConfigManagerSummary `json:"managers"`
	Registries   map[string][]string            `json:"registries"`
	Proxies      map[string][]string            `json:"proxies"`
}

// ConfigManagerSummary represents configuration summary for a specific manager
type ConfigManagerSummary struct {
	Registry string `json:"registry"`
	Proxy    string `json:"proxy"`
	Settings int    `json:"settings"`
}
