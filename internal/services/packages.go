package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"npm-console/internal/core"
	"npm-console/internal/managers"
	"npm-console/pkg/logger"
	"npm-console/pkg/utils"
)

// PackageService implements package management functionality
type PackageService struct {
	factory *managers.ManagerFactory
	logger  *logger.Logger
}

// NewPackageService creates a new package service
func NewPackageService() *PackageService {
	return &PackageService{
		factory: managers.GetGlobalFactory(),
		logger:  logger.GetDefault().WithField("service", "packages"),
	}
}

// GetAllPackages returns all packages from all available managers for a project
func (s *PackageService) GetAllPackages(ctx context.Context, projectPath string) ([]core.Package, error) {
	if projectPath == "" {
		return nil, core.NewValidationError("projectPath", projectPath, "project path cannot be empty")
	}

	availableManagers := s.factory.GetAvailableManagers(ctx)
	
	var allPackages []core.Package
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errors []error

	// Get packages concurrently from all managers
	for name, manager := range availableManagers {
		wg.Add(1)
		go func(name string, mgr core.PackageManager) {
			defer wg.Done()
			
			packages, err := mgr.GetInstalledPackages(ctx, projectPath)
			if err != nil {
				// Don't treat "project not found" as an error for this manager
				if err == core.ErrProjectNotFound {
					s.logger.WithField("manager", name).Debug("No project found for this manager")
					return
				}
				
				s.logger.WithError(err).WithField("manager", name).Warn("Failed to get packages")
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to get packages from %s: %w", name, err))
				mu.Unlock()
				return
			}
			
			mu.Lock()
			allPackages = append(allPackages, packages...)
			mu.Unlock()
		}(name, manager)
	}
	
	wg.Wait()
	
	// Remove duplicates and sort
	uniquePackages := s.removeDuplicatePackages(allPackages)
	sort.Slice(uniquePackages, func(i, j int) bool {
		return uniquePackages[i].Name < uniquePackages[j].Name
	})
	
	// Log any errors but don't fail the entire operation
	if len(errors) > 0 {
		for _, err := range errors {
			s.logger.WithError(err).Warn("Package retrieval error")
		}
	}
	
	return uniquePackages, nil
}

// GetPackagesByManager returns packages for a specific manager
func (s *PackageService) GetPackagesByManager(ctx context.Context, managerName, projectPath string) ([]core.Package, error) {
	manager, err := s.factory.GetManager(managerName)
	if err != nil {
		return nil, err
	}
	
	if !manager.IsAvailable(ctx) {
		return nil, core.NewManagerError(managerName, "get packages", core.ErrManagerNotAvailable)
	}
	
	return manager.GetInstalledPackages(ctx, projectPath)
}

// GetGlobalPackages returns all global packages from all available managers
func (s *PackageService) GetGlobalPackages(ctx context.Context) ([]core.Package, error) {
	availableManagers := s.factory.GetAvailableManagers(ctx)
	
	var allPackages []core.Package
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errors []error

	// Get global packages concurrently from all managers
	for name, manager := range availableManagers {
		wg.Add(1)
		go func(name string, mgr core.PackageManager) {
			defer wg.Done()
			
			packages, err := mgr.GetGlobalPackages(ctx)
			if err != nil {
				s.logger.WithError(err).WithField("manager", name).Warn("Failed to get global packages")
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to get global packages from %s: %w", name, err))
				mu.Unlock()
				return
			}
			
			mu.Lock()
			allPackages = append(allPackages, packages...)
			mu.Unlock()
		}(name, manager)
	}
	
	wg.Wait()
	
	// Remove duplicates and sort
	uniquePackages := s.removeDuplicatePackages(allPackages)
	sort.Slice(uniquePackages, func(i, j int) bool {
		return uniquePackages[i].Name < uniquePackages[j].Name
	})
	
	// Log any errors but don't fail the entire operation
	if len(errors) > 0 {
		for _, err := range errors {
			s.logger.WithError(err).Warn("Global package retrieval error")
		}
	}
	
	return uniquePackages, nil
}

// GetGlobalPackagesByManager returns global packages for a specific manager
func (s *PackageService) GetGlobalPackagesByManager(ctx context.Context, managerName string) ([]core.Package, error) {
	manager, err := s.factory.GetManager(managerName)
	if err != nil {
		return nil, err
	}
	
	if !manager.IsAvailable(ctx) {
		return nil, core.NewManagerError(managerName, "get global packages", core.ErrManagerNotAvailable)
	}
	
	return manager.GetGlobalPackages(ctx)
}

// SearchPackages searches for packages by name across all packages
func (s *PackageService) SearchPackages(ctx context.Context, query string) ([]core.Package, error) {
	if query == "" {
		return nil, core.NewValidationError("query", query, "search query cannot be empty")
	}
	
	// For now, we'll search through installed packages
	// In the future, this could be extended to search online registries
	globalPackages, err := s.GetGlobalPackages(ctx)
	if err != nil {
		return nil, err
	}
	
	query = strings.ToLower(query)
	var matchingPackages []core.Package
	
	for _, pkg := range globalPackages {
		if strings.Contains(strings.ToLower(pkg.Name), query) ||
		   strings.Contains(strings.ToLower(pkg.Description), query) {
			matchingPackages = append(matchingPackages, pkg)
		}
	}
	
	return matchingPackages, nil
}

// GetPackageInfo returns detailed information about a specific package
func (s *PackageService) GetPackageInfo(ctx context.Context, packageName string) (*core.PackageDetail, error) {
	if packageName == "" {
		return nil, core.NewValidationError("packageName", packageName, "package name cannot be empty")
	}
	
	// Try to find the package in global packages first
	globalPackages, err := s.GetGlobalPackages(ctx)
	if err != nil {
		return nil, err
	}
	
	for _, pkg := range globalPackages {
		if pkg.Name == packageName {
			// Convert Package to PackageDetail
			detail := &core.PackageDetail{
				Package: pkg,
				// Additional fields would be populated from package.json or registry
			}
			return detail, nil
		}
	}
	
	return nil, core.ErrPackageNotFound
}

// GetPackageStats returns statistics about packages
func (s *PackageService) GetPackageStats(ctx context.Context, projectPath string) (*PackageStats, error) {
	packages, err := s.GetAllPackages(ctx, projectPath)
	if err != nil {
		return nil, err
	}
	
	stats := &PackageStats{
		ByManager: make(map[string]int),
	}
	
	for _, pkg := range packages {
		stats.TotalPackages++
		stats.ByManager[pkg.Manager]++
		
		if pkg.IsGlobal {
			stats.GlobalPackages++
		} else {
			stats.LocalPackages++
		}
	}
	
	return stats, nil
}

// GetGlobalPackageStats returns statistics about global packages
func (s *PackageService) GetGlobalPackageStats(ctx context.Context) (*PackageStats, error) {
	packages, err := s.GetGlobalPackages(ctx)
	if err != nil {
		return nil, err
	}
	
	stats := &PackageStats{
		ByManager: make(map[string]int),
	}
	
	for _, pkg := range packages {
		stats.TotalPackages++
		stats.GlobalPackages++
		stats.ByManager[pkg.Manager]++
	}
	
	return stats, nil
}

// removeDuplicatePackages removes duplicate packages based on name and manager
func (s *PackageService) removeDuplicatePackages(packages []core.Package) []core.Package {
	seen := make(map[string]bool)
	var unique []core.Package
	
	for _, pkg := range packages {
		key := fmt.Sprintf("%s@%s:%s", pkg.Name, pkg.Version, pkg.Manager)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, pkg)
		}
	}
	
	return unique
}

// ValidateManagerName validates if a manager name is valid and available
func (s *PackageService) ValidateManagerName(ctx context.Context, managerName string) error {
	if err := s.factory.ValidateManager(managerName); err != nil {
		return err
	}
	
	if !s.factory.IsManagerAvailable(ctx, managerName) {
		return core.NewManagerError(managerName, "validate", core.ErrManagerNotAvailable)
	}
	
	return nil
}

// PackageStats represents package statistics
type PackageStats struct {
	TotalPackages  int            `json:"total_packages"`
	LocalPackages  int            `json:"local_packages"`
	GlobalPackages int            `json:"global_packages"`
	ByManager      map[string]int `json:"by_manager"`
}

// InstallPackage installs a package using the specified manager
func (s *PackageService) InstallPackage(ctx context.Context, packageName, managerName string, global bool) error {
	if packageName == "" {
		return core.NewValidationError("packageName", packageName, "package name cannot be empty")
	}

	if managerName == "" {
		return core.NewValidationError("managerName", managerName, "manager name cannot be empty")
	}

	manager, err := s.factory.GetManager(managerName)
	if err != nil {
		return err
	}

	if !manager.IsAvailable(ctx) {
		return core.NewManagerError(managerName, "install", core.ErrManagerNotAvailable)
	}

	s.logger.Info(fmt.Sprintf("Installing package %s with %s (global: %v)", packageName, managerName, global))

	// 使用命令行执行安装
	var cmd []string
	switch managerName {
	case "npm":
		if global {
			cmd = []string{"npm", "install", "-g", packageName}
		} else {
			cmd = []string{"npm", "install", packageName}
		}
	case "pnpm":
		if global {
			cmd = []string{"pnpm", "add", "-g", packageName}
		} else {
			cmd = []string{"pnpm", "add", packageName}
		}
	case "yarn":
		if global {
			cmd = []string{"yarn", "global", "add", packageName}
		} else {
			cmd = []string{"yarn", "add", packageName}
		}
	case "bun":
		if global {
			cmd = []string{"bun", "add", "-g", packageName}
		} else {
			cmd = []string{"bun", "add", packageName}
		}
	default:
		return fmt.Errorf("unsupported package manager: %s", managerName)
	}

	// 执行命令
	result := utils.ExecuteCommand(ctx, cmd[0], cmd[1:]...)
	if result.Error != nil {
		return core.NewManagerError(managerName, "install", result.Error)
	}

	if result.ExitCode != 0 {
		return core.NewManagerError(managerName, "install", fmt.Errorf("command failed with exit code %d: %s", result.ExitCode, result.Stderr))
	}

	return nil
}

// UninstallPackage uninstalls a package using the specified manager
func (s *PackageService) UninstallPackage(ctx context.Context, packageName, managerName string, global bool) error {
	if packageName == "" {
		return core.NewValidationError("packageName", packageName, "package name cannot be empty")
	}

	if managerName == "" {
		return core.NewValidationError("managerName", managerName, "manager name cannot be empty")
	}

	manager, err := s.factory.GetManager(managerName)
	if err != nil {
		return err
	}

	if !manager.IsAvailable(ctx) {
		return core.NewManagerError(managerName, "uninstall", core.ErrManagerNotAvailable)
	}

	s.logger.Info(fmt.Sprintf("Uninstalling package %s with %s (global: %v)", packageName, managerName, global))

	// 使用命令行执行卸载
	var cmd []string
	switch managerName {
	case "npm":
		if global {
			cmd = []string{"npm", "uninstall", "-g", packageName}
		} else {
			cmd = []string{"npm", "uninstall", packageName}
		}
	case "pnpm":
		if global {
			cmd = []string{"pnpm", "remove", "-g", packageName}
		} else {
			cmd = []string{"pnpm", "remove", packageName}
		}
	case "yarn":
		if global {
			cmd = []string{"yarn", "global", "remove", packageName}
		} else {
			cmd = []string{"yarn", "remove", packageName}
		}
	case "bun":
		if global {
			cmd = []string{"bun", "remove", "-g", packageName}
		} else {
			cmd = []string{"bun", "remove", packageName}
		}
	default:
		return fmt.Errorf("unsupported package manager: %s", managerName)
	}

	// 执行命令
	result := utils.ExecuteCommand(ctx, cmd[0], cmd[1:]...)
	if result.Error != nil {
		return core.NewManagerError(managerName, "uninstall", result.Error)
	}

	if result.ExitCode != 0 {
		return core.NewManagerError(managerName, "uninstall", fmt.Errorf("command failed with exit code %d: %s", result.ExitCode, result.Stderr))
	}

	return nil
}
