package managers

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"npm-console/internal/core"
	"npm-console/pkg/logger"
	"npm-console/pkg/utils"
)

// NPMManager implements the PackageManager interface for npm
type NPMManager struct {
	logger *logger.Logger
}

// NewNPMManager creates a new NPM manager instance
func NewNPMManager() *NPMManager {
	return &NPMManager{
		logger: logger.GetDefault().WithField("manager", "npm"),
	}
}

// Name returns the name of the package manager
func (n *NPMManager) Name() string {
	return "npm"
}

// IsAvailable checks if npm is installed and available
func (n *NPMManager) IsAvailable(ctx context.Context) bool {
	result := utils.ExecuteCommand(ctx, "npm", "--version")
	return result.Error == nil
}

// GetCacheInfo returns information about npm cache
func (n *NPMManager) GetCacheInfo(ctx context.Context) (*core.CacheInfo, error) {
	// Get npm cache directory
	result := utils.ExecuteCommand(ctx, "npm", "config", "get", "cache")
	if result.Error != nil {
		return nil, core.NewManagerError("npm", "get cache path", result.Error)
	}

	cachePath := strings.TrimSpace(result.Stdout)
	if cachePath == "" {
		// Use default cache path
		cachePath = n.getDefaultCachePath()
	}

	// Expand path if needed
	expandedPath, err := utils.ExpandPath(cachePath)
	if err != nil {
		return nil, core.NewManagerError("npm", "expand cache path", err)
	}

	// Check if cache directory exists
	if !utils.PathExists(expandedPath) {
		return &core.CacheInfo{
			Manager:     "npm",
			Path:        expandedPath,
			Size:        0,
			FileCount:   0,
			LastUpdated: time.Time{},
		}, nil
	}

	// Get cache size and file count
	size, err := utils.GetDirSize(expandedPath)
	if err != nil {
		n.logger.WithError(err).Warn("Failed to get cache size")
		size = 0
	}

	fileCount, err := utils.GetFileCount(expandedPath)
	if err != nil {
		n.logger.WithError(err).Warn("Failed to get file count")
		fileCount = 0
	}

	// Get last modified time
	info, err := os.Stat(expandedPath)
	var lastUpdated time.Time
	if err == nil {
		lastUpdated = info.ModTime()
	}

	return &core.CacheInfo{
		Manager:     "npm",
		Path:        expandedPath,
		Size:        size,
		FileCount:   fileCount,
		LastUpdated: lastUpdated,
	}, nil
}

// ClearCache clears the npm cache
func (n *NPMManager) ClearCache(ctx context.Context) error {
	result := utils.ExecuteCommand(ctx, "npm", "cache", "clean", "--force")
	if result.Error != nil {
		return core.NewManagerError("npm", "clear cache", result.Error)
	}
	
	n.logger.Info("npm cache cleared successfully")
	return nil
}

// GetInstalledPackages returns packages installed in a specific project
func (n *NPMManager) GetInstalledPackages(ctx context.Context, projectPath string) ([]core.Package, error) {
	// Check if package.json exists
	packageJsonPath := filepath.Join(projectPath, "package.json")
	if !utils.IsFile(packageJsonPath) {
		return nil, core.ErrProjectNotFound
	}

	// Read package.json
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return nil, core.NewManagerError("npm", "read package.json", err)
	}

	var packageJson struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &packageJson); err != nil {
		return nil, core.NewManagerError("npm", "parse package.json", err)
	}

	var packages []core.Package

	// Add regular dependencies
	for name, version := range packageJson.Dependencies {
		pkg := core.Package{
			Name:     name,
			Version:  version,
			Manager:  "npm",
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
			Manager:  "npm",
			IsGlobal: false,
			Path:     filepath.Join(projectPath, "node_modules", name),
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// GetGlobalPackages returns globally installed npm packages
func (n *NPMManager) GetGlobalPackages(ctx context.Context) ([]core.Package, error) {
	result := utils.ExecuteCommand(ctx, "npm", "list", "-g", "--depth=0", "--json")
	if result.Error != nil {
		return nil, core.NewManagerError("npm", "list global packages", result.Error)
	}

	var npmList struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal([]byte(result.Stdout), &npmList); err != nil {
		return nil, core.NewManagerError("npm", "parse npm list output", err)
	}

	var packages []core.Package
	for name, info := range npmList.Dependencies {
		pkg := core.Package{
			Name:     name,
			Version:  info.Version,
			Manager:  "npm",
			IsGlobal: true,
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// GetConfig returns the current npm configuration
func (n *NPMManager) GetConfig(ctx context.Context) (*core.Config, error) {
	config := &core.Config{
		Manager:  "npm",
		Settings: make(map[string]string),
	}

	// Get registry
	result := utils.ExecuteCommand(ctx, "npm", "config", "get", "registry")
	if result.Error == nil {
		config.Registry = strings.TrimSpace(result.Stdout)
	}

	// Get proxy
	result = utils.ExecuteCommand(ctx, "npm", "config", "get", "proxy")
	if result.Error == nil && result.Stdout != "null" {
		config.Proxy = strings.TrimSpace(result.Stdout)
	}

	// Get other common settings
	settings := []string{"cache", "prefix", "userconfig", "globalconfig"}
	for _, setting := range settings {
		result = utils.ExecuteCommand(ctx, "npm", "config", "get", setting)
		if result.Error == nil {
			config.Settings[setting] = strings.TrimSpace(result.Stdout)
		}
	}

	return config, nil
}

// SetRegistry sets the npm registry URL
func (n *NPMManager) SetRegistry(ctx context.Context, url string) error {
	result := utils.ExecuteCommand(ctx, "npm", "config", "set", "registry", url)
	if result.Error != nil {
		return core.NewManagerError("npm", "set registry", result.Error)
	}

	n.logger.WithField("registry", url).Info("npm registry updated")
	return nil
}

// SetProxy sets the npm proxy configuration
func (n *NPMManager) SetProxy(ctx context.Context, proxy string) error {
	if proxy == "" {
		// Remove proxy
		result := utils.ExecuteCommand(ctx, "npm", "config", "delete", "proxy")
		if result.Error != nil {
			return core.NewManagerError("npm", "remove proxy", result.Error)
		}
		result = utils.ExecuteCommand(ctx, "npm", "config", "delete", "https-proxy")
		if result.Error != nil {
			return core.NewManagerError("npm", "remove https-proxy", result.Error)
		}
	} else {
		// Set proxy
		result := utils.ExecuteCommand(ctx, "npm", "config", "set", "proxy", proxy)
		if result.Error != nil {
			return core.NewManagerError("npm", "set proxy", result.Error)
		}
		result = utils.ExecuteCommand(ctx, "npm", "config", "set", "https-proxy", proxy)
		if result.Error != nil {
			return core.NewManagerError("npm", "set https-proxy", result.Error)
		}
	}

	n.logger.WithField("proxy", proxy).Info("npm proxy updated")
	return nil
}

// GetProjects scans for npm projects
func (n *NPMManager) GetProjects(ctx context.Context, rootPath string) ([]core.Project, error) {
	var projects []core.Project

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if info.Name() == "package.json" && utils.IsFile(path) {
			projectPath := filepath.Dir(path)
			
			// Read package.json to get project name
			data, err := os.ReadFile(path)
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
				Managers:    []string{"npm"},
				PackageFile: path,
				NodeModules: filepath.Join(projectPath, "node_modules"),
			}

			// Check for lock files
			if utils.IsFile(filepath.Join(projectPath, "package-lock.json")) {
				project.LockFile = filepath.Join(projectPath, "package-lock.json")
			}

			projects = append(projects, project)
		}

		return nil
	})

	if err != nil {
		return nil, core.NewManagerError("npm", "scan projects", err)
	}

	return projects, nil
}

// getDefaultCachePath returns the default npm cache path for the current OS
func (n *NPMManager) getDefaultCachePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "npm-cache")
	case "darwin":
		home, _ := utils.GetHomeDir()
		return filepath.Join(home, ".npm")
	default: // linux and others
		home, _ := utils.GetHomeDir()
		return filepath.Join(home, ".npm")
	}
}
