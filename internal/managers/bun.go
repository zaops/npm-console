package managers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"npm-console/internal/core"
	"npm-console/pkg/logger"
	"npm-console/pkg/utils"
)

// BunManager implements the PackageManager interface for bun
type BunManager struct {
	logger *logger.Logger
}

// NewBunManager creates a new Bun manager instance
func NewBunManager() *BunManager {
	return &BunManager{
		logger: logger.GetDefault().WithField("manager", "bun"),
	}
}

// Name returns the name of the package manager
func (b *BunManager) Name() string {
	return "bun"
}

// IsAvailable checks if bun is installed and available
func (b *BunManager) IsAvailable(ctx context.Context) bool {
	result := utils.ExecuteCommand(ctx, "bun", "--version")
	return result.Error == nil
}

// GetCacheInfo returns information about bun cache
func (b *BunManager) GetCacheInfo(ctx context.Context) (*core.CacheInfo, error) {
	// Bun doesn't have a direct cache dir command, use default path
	cachePath := b.getDefaultCachePath()

	// Expand path if needed
	expandedPath, err := utils.ExpandPath(cachePath)
	if err != nil {
		return nil, core.NewManagerError("bun", "expand cache path", err)
	}

	// Check if cache directory exists
	if !utils.PathExists(expandedPath) {
		return &core.CacheInfo{
			Manager:     "bun",
			Path:        expandedPath,
			Size:        0,
			FileCount:   0,
			LastUpdated: time.Time{},
		}, nil
	}

	// Get cache size and file count
	size, err := utils.GetDirSize(expandedPath)
	if err != nil {
		b.logger.WithError(err).Warn("Failed to get cache size")
		size = 0
	}

	fileCount, err := utils.GetFileCount(expandedPath)
	if err != nil {
		b.logger.WithError(err).Warn("Failed to get file count")
		fileCount = 0
	}

	// Get last modified time
	info, err := os.Stat(expandedPath)
	var lastUpdated time.Time
	if err == nil {
		lastUpdated = info.ModTime()
	}

	return &core.CacheInfo{
		Manager:     "bun",
		Path:        expandedPath,
		Size:        size,
		FileCount:   fileCount,
		LastUpdated: lastUpdated,
	}, nil
}

// ClearCache clears the bun cache
func (b *BunManager) ClearCache(ctx context.Context) error {
	// Bun doesn't have a built-in cache clean command, manually remove cache directory
	cachePath := b.getDefaultCachePath()
	expandedPath, err := utils.ExpandPath(cachePath)
	if err != nil {
		return core.NewManagerError("bun", "expand cache path", err)
	}

	if utils.PathExists(expandedPath) {
		if err := utils.RemoveDir(expandedPath); err != nil {
			return core.NewManagerError("bun", "remove cache directory", err)
		}
	}
	
	b.logger.Info("bun cache cleared successfully")
	return nil
}

// GetInstalledPackages returns packages installed in a specific project
func (b *BunManager) GetInstalledPackages(ctx context.Context, projectPath string) ([]core.Package, error) {
	// Check if package.json exists
	packageJsonPath := filepath.Join(projectPath, "package.json")
	if !utils.IsFile(packageJsonPath) {
		return nil, core.ErrProjectNotFound
	}

	// Bun doesn't have a list command yet, read from package.json
	return b.getPackagesFromPackageJson(packageJsonPath)
}

// GetGlobalPackages returns globally installed bun packages
func (b *BunManager) GetGlobalPackages(ctx context.Context) ([]core.Package, error) {
	// Bun doesn't support global packages in the traditional sense
	// Return empty list for now
	return []core.Package{}, nil
}

// GetConfig returns the current bun configuration
func (b *BunManager) GetConfig(ctx context.Context) (*core.Config, error) {
	config := &core.Config{
		Manager:  "bun",
		Settings: make(map[string]string),
	}

	// Try to read bunfig.toml if it exists
	bunfigPath := filepath.Join(".", "bunfig.toml")
	if utils.IsFile(bunfigPath) {
		// For now, just note that bunfig.toml exists
		config.Settings["bunfig"] = bunfigPath
	}

	// Check for global bunfig.toml
	home, err := utils.GetHomeDir()
	if err == nil {
		globalBunfig := filepath.Join(home, ".bunfig.toml")
		if utils.IsFile(globalBunfig) {
			config.Settings["global-bunfig"] = globalBunfig
		}
	}

	// Bun uses npm registry by default
	config.Registry = "https://registry.npmjs.org/"

	return config, nil
}

// SetRegistry sets the bun registry URL
func (b *BunManager) SetRegistry(ctx context.Context, url string) error {
	// Bun doesn't have a config set command yet
	// This would need to be implemented by modifying bunfig.toml
	b.logger.WithField("registry", url).Warn("bun registry configuration not yet supported")
	return core.NewManagerError("bun", "set registry", fmt.Errorf("registry configuration not supported"))
}

// SetProxy sets the bun proxy configuration
func (b *BunManager) SetProxy(ctx context.Context, proxy string) error {
	// Bun doesn't have built-in proxy configuration
	// This would need to be implemented by modifying bunfig.toml
	b.logger.WithField("proxy", proxy).Warn("bun proxy configuration not yet supported")
	return core.NewManagerError("bun", "set proxy", fmt.Errorf("proxy configuration not supported"))
}

// GetProjects scans for bun projects
func (b *BunManager) GetProjects(ctx context.Context, rootPath string) ([]core.Project, error) {
	var projects []core.Project

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		// Look for bun.lockb files
		if info.Name() == "bun.lockb" && utils.IsFile(path) {
			projectPath := filepath.Dir(path)
			packageJsonPath := filepath.Join(projectPath, "package.json")
			
			if !utils.IsFile(packageJsonPath) {
				return nil // Continue walking
			}

			// Read package.json to get project name
			data, err := os.ReadFile(packageJsonPath)
			if err != nil {
				return nil // Continue walking
			}

			var packageJson struct {
				Name string `json:"name"`
			}
			
			if err := json.Unmarshal(data, &packageJson); err != nil {
				return nil // Continue walking
			}

			project := core.Project{
				Name:        packageJson.Name,
				Path:        projectPath,
				Managers:    []string{"bun"},
				PackageFile: packageJsonPath,
				LockFile:    path,
				NodeModules: filepath.Join(projectPath, "node_modules"),
			}

			projects = append(projects, project)
		}

		return nil
	})

	if err != nil {
		return nil, core.NewManagerError("bun", "scan projects", err)
	}

	return projects, nil
}

// getDefaultCachePath returns the default bun cache path for the current OS
func (b *BunManager) getDefaultCachePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "bun", "cache")
	case "darwin":
		home, _ := utils.GetHomeDir()
		return filepath.Join(home, "Library", "Caches", "bun")
	default: // linux and others
		home, _ := utils.GetHomeDir()
		return filepath.Join(home, ".cache", "bun")
	}
}

// getPackagesFromPackageJson reads packages from package.json
func (b *BunManager) getPackagesFromPackageJson(packageJsonPath string) ([]core.Package, error) {
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return nil, core.NewManagerError("bun", "read package.json", err)
	}

	var packageJson struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &packageJson); err != nil {
		return nil, core.NewManagerError("bun", "parse package.json", err)
	}

	var packages []core.Package
	projectPath := filepath.Dir(packageJsonPath)

	// Add regular dependencies
	for name, version := range packageJson.Dependencies {
		pkg := core.Package{
			Name:     name,
			Version:  version,
			Manager:  "bun",
			IsGlobal: false,
			Path:     filepath.Join(projectPath, "node_modules", name),
		}
		packages = append(packages, pkg)
	}

	// Add dev dependencies
	for name, version := range packageJson.DevDependencies {
		pkg := core.Package{
			Name:     name,
			Version:  version,
			Manager:  "bun",
			IsGlobal: false,
			Path:     filepath.Join(projectPath, "node_modules", name),
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}
