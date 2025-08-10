package managers

import (
	"context"
	"fmt"
	"sync"

	"npm-console/internal/core"
	"npm-console/pkg/logger"
)

// ManagerFactory manages package manager instances
type ManagerFactory struct {
	managers map[string]core.PackageManager
	logger   *logger.Logger
	mu       sync.RWMutex
}

// NewManagerFactory creates a new manager factory
func NewManagerFactory() *ManagerFactory {
	factory := &ManagerFactory{
		managers: make(map[string]core.PackageManager),
		logger:   logger.GetDefault().WithField("component", "manager-factory"),
	}

	// Register all available managers
	factory.registerManagers()
	
	return factory
}

// registerManagers registers all available package managers
func (f *ManagerFactory) registerManagers() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Register npm manager
	f.managers["npm"] = NewNPMManager()
	
	// Register pnpm manager
	f.managers["pnpm"] = NewPNPMManager()
	
	// Register yarn manager
	f.managers["yarn"] = NewYarnManager()
	
	// Register bun manager
	f.managers["bun"] = NewBunManager()

	f.logger.Info("Registered package managers", "count", len(f.managers))
}

// GetManager returns a package manager by name
func (f *ManagerFactory) GetManager(name string) (core.PackageManager, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	manager, exists := f.managers[name]
	if !exists {
		return nil, fmt.Errorf("package manager '%s' not found", name)
	}

	return manager, nil
}

// GetAllManagers returns all registered package managers
func (f *ManagerFactory) GetAllManagers() map[string]core.PackageManager {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]core.PackageManager)
	for name, manager := range f.managers {
		result[name] = manager
	}

	return result
}

// GetAvailableManagers returns only the managers that are available on the system
func (f *ManagerFactory) GetAvailableManagers(ctx context.Context) map[string]core.PackageManager {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[string]core.PackageManager)
	for name, manager := range f.managers {
		if manager.IsAvailable(ctx) {
			result[name] = manager
		}
	}

	return result
}

// GetManagerNames returns the names of all registered managers
func (f *ManagerFactory) GetManagerNames() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.managers))
	for name := range f.managers {
		names = append(names, name)
	}

	return names
}

// GetAvailableManagerNames returns the names of available managers
func (f *ManagerFactory) GetAvailableManagerNames(ctx context.Context) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var names []string
	for name, manager := range f.managers {
		if manager.IsAvailable(ctx) {
			names = append(names, name)
		}
	}

	return names
}

// IsManagerAvailable checks if a specific manager is available
func (f *ManagerFactory) IsManagerAvailable(ctx context.Context, name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	manager, exists := f.managers[name]
	if !exists {
		return false
	}

	return manager.IsAvailable(ctx)
}

// RegisterManager registers a custom package manager
func (f *ManagerFactory) RegisterManager(name string, manager core.PackageManager) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.managers[name]; exists {
		return fmt.Errorf("package manager '%s' already registered", name)
	}

	f.managers[name] = manager
	f.logger.Info("Registered custom package manager", "name", name)
	
	return nil
}

// UnregisterManager unregisters a package manager
func (f *ManagerFactory) UnregisterManager(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.managers[name]; !exists {
		return fmt.Errorf("package manager '%s' not found", name)
	}

	delete(f.managers, name)
	f.logger.Info("Unregistered package manager", "name", name)
	
	return nil
}

// GetManagerCount returns the number of registered managers
func (f *ManagerFactory) GetManagerCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.managers)
}

// GetAvailableManagerCount returns the number of available managers
func (f *ManagerFactory) GetAvailableManagerCount(ctx context.Context) int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	count := 0
	for _, manager := range f.managers {
		if manager.IsAvailable(ctx) {
			count++
		}
	}

	return count
}

// ValidateManager checks if a manager name is valid
func (f *ManagerFactory) ValidateManager(name string) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if _, exists := f.managers[name]; !exists {
		return core.NewValidationError("manager", name, "unknown package manager")
	}

	return nil
}

// GetManagerInfo returns basic information about all managers
func (f *ManagerFactory) GetManagerInfo(ctx context.Context) []ManagerInfo {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var info []ManagerInfo
	for name, manager := range f.managers {
		managerInfo := ManagerInfo{
			Name:      name,
			Available: manager.IsAvailable(ctx),
		}
		info = append(info, managerInfo)
	}

	return info
}

// ManagerInfo represents basic information about a package manager
type ManagerInfo struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// Global factory instance
var globalFactory *ManagerFactory
var factoryOnce sync.Once

// GetGlobalFactory returns the global manager factory instance
func GetGlobalFactory() *ManagerFactory {
	factoryOnce.Do(func() {
		globalFactory = NewManagerFactory()
	})
	return globalFactory
}

// GetManager is a convenience function to get a manager from the global factory
func GetManager(name string) (core.PackageManager, error) {
	return GetGlobalFactory().GetManager(name)
}

// GetAllManagers is a convenience function to get all managers from the global factory
func GetAllManagers() map[string]core.PackageManager {
	return GetGlobalFactory().GetAllManagers()
}

// GetAvailableManagers is a convenience function to get available managers from the global factory
func GetAvailableManagers(ctx context.Context) map[string]core.PackageManager {
	return GetGlobalFactory().GetAvailableManagers(ctx)
}
