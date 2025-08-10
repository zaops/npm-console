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

// YarnManager implements the PackageManager interface for yarn
type YarnManager struct {
	logger *logger.Logger
}

// NewYarnManager creates a new Yarn manager instance
func NewYarnManager() *YarnManager {
	return &YarnManager{
		logger: logger.GetDefault().WithField("manager", "yarn"),
	}
}

// Name returns the name of the package manager
func (y *YarnManager) Name() string {
	return "yarn"
}

// IsAvailable checks if yarn is installed and available
func (y *YarnManager) IsAvailable(ctx context.Context) bool {
	result := utils.ExecuteCommand(ctx, "yarn", "--version")
	return result.Error == nil
}

// GetCacheInfo returns information about yarn cache
func (y *YarnManager) GetCacheInfo(ctx context.Context) (*core.CacheInfo, error) {
	// Try to get yarn cache directory
	result := utils.ExecuteCommand(ctx, "yarn", "cache", "dir")
	var cachePath string
	
	if result.Error != nil {
		// Fallback to config get
		result = utils.ExecuteCommand(ctx, "yarn", "config", "get", "cache-folder")
		if result.Error != nil {
			// Use default cache path
			cachePath = y.getDefaultCachePath()
		} else {
			cachePath = strings.TrimSpace(result.Stdout)
		}
	} else {
		cachePath = strings.TrimSpace(result.Stdout)
	}

	// Expand path if needed
	expandedPath, err := utils.ExpandPath(cachePath)
	if err != nil {
		return nil, core.NewManagerError("yarn", "expand cache path", err)
	}

	// Check if cache directory exists
	if !utils.PathExists(expandedPath) {
		return &core.CacheInfo{
			Manager:     "yarn",
			Path:        expandedPath,
			Size:        0,
			FileCount:   0,
			LastUpdated: time.Time{},
		}, nil
	}

	// Get cache size and file count
	size, err := utils.GetDirSize(expandedPath)
	if err != nil {
		y.logger.WithError(err).Warn("Failed to get cache size")
		size = 0
	}

	fileCount, err := utils.GetFileCount(expandedPath)
	if err != nil {
		y.logger.WithError(err).Warn("Failed to get file count")
		fileCount = 0
	}

	// Get last modified time
	info, err := os.Stat(expandedPath)
	var lastUpdated time.Time
	if err == nil {
		lastUpdated = info.ModTime()
	}

	return &core.CacheInfo{
		Manager:     "yarn",
		Path:        expandedPath,
		Size:        size,
		FileCount:   fileCount,
		LastUpdated: lastUpdated,
	}, nil
}

// ClearCache clears the yarn cache
func (y *YarnManager) ClearCache(ctx context.Context) error {
	result := utils.ExecuteCommand(ctx, "yarn", "cache", "clean")
	if result.Error != nil {
		return core.NewManagerError("yarn", "clear cache", result.Error)
	}
	
	y.logger.Info("yarn cache cleared successfully")
	return nil
}

// GetInstalledPackages returns packages installed in a specific project
func (y *YarnManager) GetInstalledPackages(ctx context.Context, projectPath string) ([]core.Package, error) {
	// Check if package.json exists
	packageJsonPath := filepath.Join(projectPath, "package.json")
	if !utils.IsFile(packageJsonPath) {
		return nil, core.ErrProjectNotFound
	}

	// Try yarn list first
	result := utils.ExecuteCommandWithTimeout(30*time.Second, "yarn", "list", "--json", "--depth=0")
	if result.Error == nil {
		return y.parseYarnListOutput(result.Stdout)
	}

	// Fallback to reading package.json
	return y.getPackagesFromPackageJson(packageJsonPath)
}

// GetGlobalPackages returns globally installed yarn packages
func (y *YarnManager) GetGlobalPackages(ctx context.Context) ([]core.Package, error) {
	result := utils.ExecuteCommand(ctx, "yarn", "global", "list", "--json", "--depth=0")
	if result.Error != nil {
		return nil, core.NewManagerError("yarn", "list global packages", result.Error)
	}

	return y.parseYarnListOutput(result.Stdout)
}

// GetConfig returns the current yarn configuration
func (y *YarnManager) GetConfig(ctx context.Context) (*core.Config, error) {
	config := &core.Config{
		Manager:  "yarn",
		Settings: make(map[string]string),
	}

	// Get registry
	result := utils.ExecuteCommand(ctx, "yarn", "config", "get", "registry")
	if result.Error == nil {
		config.Registry = strings.TrimSpace(result.Stdout)
	}

	// Get proxy
	result = utils.ExecuteCommand(ctx, "yarn", "config", "get", "proxy")
	if result.Error == nil && result.Stdout != "undefined" {
		config.Proxy = strings.TrimSpace(result.Stdout)
	}

	// Get other common settings
	settings := []string{"cache-folder", "global-folder", "yarn-offline-mirror"}
	for _, setting := range settings {
		result = utils.ExecuteCommand(ctx, "yarn", "config", "get", setting)
		if result.Error == nil && result.Stdout != "undefined" {
			config.Settings[setting] = strings.TrimSpace(result.Stdout)
		}
	}

	return config, nil
}

// SetRegistry sets the yarn registry URL
func (y *YarnManager) SetRegistry(ctx context.Context, url string) error {
	result := utils.ExecuteCommand(ctx, "yarn", "config", "set", "registry", url)
	if result.Error != nil {
		return core.NewManagerError("yarn", "set registry", result.Error)
	}

	y.logger.WithField("registry", url).Info("yarn registry updated")
	return nil
}

// SetProxy sets the yarn proxy configuration
func (y *YarnManager) SetProxy(ctx context.Context, proxy string) error {
	if proxy == "" {
		// Remove proxy
		result := utils.ExecuteCommand(ctx, "yarn", "config", "delete", "proxy")
		if result.Error != nil {
			return core.NewManagerError("yarn", "remove proxy", result.Error)
		}
		result = utils.ExecuteCommand(ctx, "yarn", "config", "delete", "https-proxy")
		if result.Error != nil {
			return core.NewManagerError("yarn", "remove https-proxy", result.Error)
		}
	} else {
		// Set proxy
		result := utils.ExecuteCommand(ctx, "yarn", "config", "set", "proxy", proxy)
		if result.Error != nil {
			return core.NewManagerError("yarn", "set proxy", result.Error)
		}
		result = utils.ExecuteCommand(ctx, "yarn", "config", "set", "https-proxy", proxy)
		if result.Error != nil {
			return core.NewManagerError("yarn", "set https-proxy", result.Error)
		}
	}

	y.logger.WithField("proxy", proxy).Info("yarn proxy updated")
	return nil
}

// GetProjects scans for yarn projects
func (y *YarnManager) GetProjects(ctx context.Context, rootPath string) ([]core.Project, error) {
	var projects []core.Project

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		// Look for yarn.lock files
		if info.Name() == "yarn.lock" && utils.IsFile(path) {
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
				Managers:    []string{"yarn"},
				PackageFile: packageJsonPath,
				LockFile:    path,
				NodeModules: filepath.Join(projectPath, "node_modules"),
			}

			projects = append(projects, project)
		}

		return nil
	})

	if err != nil {
		return nil, core.NewManagerError("yarn", "scan projects", err)
	}

	return projects, nil
}

// getDefaultCachePath returns the default yarn cache path for the current OS
func (y *YarnManager) getDefaultCachePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Yarn", "Cache")
	case "darwin":
		home, _ := utils.GetHomeDir()
		return filepath.Join(home, "Library", "Caches", "Yarn")
	default: // linux and others
		home, _ := utils.GetHomeDir()
		return filepath.Join(home, ".cache", "yarn")
	}
}

// parseYarnListOutput parses yarn list JSON output
func (y *YarnManager) parseYarnListOutput(output string) ([]core.Package, error) {
	var packages []core.Package
	
	// Yarn outputs multiple JSON objects, one per line
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var yarnOutput struct {
			Type string `json:"type"`
			Data struct {
				Trees []struct {
					Name     string `json:"name"`
					Children []struct {
						Name string `json:"name"`
					} `json:"children"`
				} `json:"trees"`
			} `json:"data"`
		}

		if err := json.Unmarshal([]byte(line), &yarnOutput); err != nil {
			continue // Skip invalid JSON lines
		}

		if yarnOutput.Type == "tree" {
			for _, tree := range yarnOutput.Data.Trees {
				// Parse package name and version
				parts := strings.Split(tree.Name, "@")
				if len(parts) >= 2 {
					name := strings.Join(parts[:len(parts)-1], "@")
					version := parts[len(parts)-1]
					
					pkg := core.Package{
						Name:     name,
						Version:  version,
						Manager:  "yarn",
						IsGlobal: false,
					}
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages, nil
}

// getPackagesFromPackageJson fallback method to read packages from package.json
func (y *YarnManager) getPackagesFromPackageJson(packageJsonPath string) ([]core.Package, error) {
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return nil, core.NewManagerError("yarn", "read package.json", err)
	}

	var packageJson struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &packageJson); err != nil {
		return nil, core.NewManagerError("yarn", "parse package.json", err)
	}

	var packages []core.Package
	projectPath := filepath.Dir(packageJsonPath)

	// Add regular dependencies
	for name, version := range packageJson.Dependencies {
		pkg := core.Package{
			Name:     name,
			Version:  version,
			Manager:  "yarn",
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
			Manager:  "yarn",
			IsGlobal: false,
			Path:     filepath.Join(projectPath, "node_modules", name),
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}
