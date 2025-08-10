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

// PNPMManager implements the PackageManager interface for pnpm
type PNPMManager struct {
	logger *logger.Logger
}

// NewPNPMManager creates a new PNPM manager instance
func NewPNPMManager() *PNPMManager {
	return &PNPMManager{
		logger: logger.GetDefault().WithField("manager", "pnpm"),
	}
}

// Name returns the name of the package manager
func (p *PNPMManager) Name() string {
	return "pnpm"
}

// IsAvailable checks if pnpm is installed and available
func (p *PNPMManager) IsAvailable(ctx context.Context) bool {
	result := utils.ExecuteCommand(ctx, "pnpm", "--version")
	return result.Error == nil
}

// GetCacheInfo returns information about pnpm store
func (p *PNPMManager) GetCacheInfo(ctx context.Context) (*core.CacheInfo, error) {
	// Get pnpm store path
	result := utils.ExecuteCommand(ctx, "pnpm", "store", "path")
	if result.Error != nil {
		return nil, core.NewManagerError("pnpm", "get store path", result.Error)
	}

	storePath := strings.TrimSpace(result.Stdout)
	if storePath == "" {
		// Use default store path
		storePath = p.getDefaultStorePath()
	}

	// Expand path if needed
	expandedPath, err := utils.ExpandPath(storePath)
	if err != nil {
		return nil, core.NewManagerError("pnpm", "expand store path", err)
	}

	// Check if store directory exists
	if !utils.PathExists(expandedPath) {
		return &core.CacheInfo{
			Manager:     "pnpm",
			Path:        expandedPath,
			Size:        0,
			FileCount:   0,
			LastUpdated: time.Time{},
		}, nil
	}

	// Get store size and file count
	size, err := utils.GetDirSize(expandedPath)
	if err != nil {
		p.logger.WithError(err).Warn("Failed to get store size")
		size = 0
	}

	fileCount, err := utils.GetFileCount(expandedPath)
	if err != nil {
		p.logger.WithError(err).Warn("Failed to get file count")
		fileCount = 0
	}

	// Get last modified time
	info, err := os.Stat(expandedPath)
	var lastUpdated time.Time
	if err == nil {
		lastUpdated = info.ModTime()
	}

	return &core.CacheInfo{
		Manager:     "pnpm",
		Path:        expandedPath,
		Size:        size,
		FileCount:   fileCount,
		LastUpdated: lastUpdated,
	}, nil
}

// ClearCache clears the pnpm store
func (p *PNPMManager) ClearCache(ctx context.Context) error {
	result := utils.ExecuteCommand(ctx, "pnpm", "store", "prune")
	if result.Error != nil {
		return core.NewManagerError("pnpm", "prune store", result.Error)
	}
	
	p.logger.Info("pnpm store pruned successfully")
	return nil
}

// GetInstalledPackages returns packages installed in a specific project
func (p *PNPMManager) GetInstalledPackages(ctx context.Context, projectPath string) ([]core.Package, error) {
	// Check if package.json exists
	packageJsonPath := filepath.Join(projectPath, "package.json")
	if !utils.IsFile(packageJsonPath) {
		return nil, core.ErrProjectNotFound
	}

	// Use pnpm list to get installed packages
	result := utils.ExecuteCommandWithTimeout(30*time.Second, "pnpm", "list", "--json", "--depth=0")
	if result.Error != nil {
		// Fallback to reading package.json
		return p.getPackagesFromPackageJson(packageJsonPath)
	}

	var pnpmList []struct {
		Name         string            `json:"name"`
		Dependencies map[string]struct {
			Version string `json:"version"`
			Path    string `json:"path"`
		} `json:"dependencies"`
		DevDependencies map[string]struct {
			Version string `json:"version"`
			Path    string `json:"path"`
		} `json:"devDependencies"`
	}

	if err := json.Unmarshal([]byte(result.Stdout), &pnpmList); err != nil {
		return nil, core.NewManagerError("pnpm", "parse pnpm list output", err)
	}

	var packages []core.Package
	
	if len(pnpmList) > 0 {
		item := pnpmList[0]
		
		// Add regular dependencies
		for name, info := range item.Dependencies {
			pkg := core.Package{
				Name:     name,
				Version:  info.Version,
				Manager:  "pnpm",
				IsGlobal: false,
				Path:     info.Path,
			}
			packages = append(packages, pkg)
		}

		// Add dev dependencies
		for name, info := range item.DevDependencies {
			pkg := core.Package{
				Name:     name,
				Version:  info.Version,
				Manager:  "pnpm",
				IsGlobal: false,
				Path:     info.Path,
			}
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// GetGlobalPackages returns globally installed pnpm packages
func (p *PNPMManager) GetGlobalPackages(ctx context.Context) ([]core.Package, error) {
	result := utils.ExecuteCommand(ctx, "pnpm", "list", "-g", "--depth=0", "--json")
	if result.Error != nil {
		return nil, core.NewManagerError("pnpm", "list global packages", result.Error)
	}

	var pnpmList []struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
			Path    string `json:"path"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal([]byte(result.Stdout), &pnpmList); err != nil {
		return nil, core.NewManagerError("pnpm", "parse pnpm global list output", err)
	}

	var packages []core.Package
	
	if len(pnpmList) > 0 {
		for name, info := range pnpmList[0].Dependencies {
			pkg := core.Package{
				Name:     name,
				Version:  info.Version,
				Manager:  "pnpm",
				IsGlobal: true,
				Path:     info.Path,
			}
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// GetConfig returns the current pnpm configuration
func (p *PNPMManager) GetConfig(ctx context.Context) (*core.Config, error) {
	config := &core.Config{
		Manager:  "pnpm",
		Settings: make(map[string]string),
	}

	// Get registry
	result := utils.ExecuteCommand(ctx, "pnpm", "config", "get", "registry")
	if result.Error == nil {
		config.Registry = strings.TrimSpace(result.Stdout)
	}

	// Get proxy
	result = utils.ExecuteCommand(ctx, "pnpm", "config", "get", "proxy")
	if result.Error == nil && result.Stdout != "undefined" {
		config.Proxy = strings.TrimSpace(result.Stdout)
	}

	// Get store directory
	result = utils.ExecuteCommand(ctx, "pnpm", "store", "path")
	if result.Error == nil {
		config.Settings["store-dir"] = strings.TrimSpace(result.Stdout)
	}

	// Get other common settings
	settings := []string{"cache-dir", "state-dir", "global-dir"}
	for _, setting := range settings {
		result = utils.ExecuteCommand(ctx, "pnpm", "config", "get", setting)
		if result.Error == nil && result.Stdout != "undefined" {
			config.Settings[setting] = strings.TrimSpace(result.Stdout)
		}
	}

	return config, nil
}

// SetRegistry sets the pnpm registry URL
func (p *PNPMManager) SetRegistry(ctx context.Context, url string) error {
	result := utils.ExecuteCommand(ctx, "pnpm", "config", "set", "registry", url)
	if result.Error != nil {
		return core.NewManagerError("pnpm", "set registry", result.Error)
	}

	p.logger.WithField("registry", url).Info("pnpm registry updated")
	return nil
}

// SetProxy sets the pnpm proxy configuration
func (p *PNPMManager) SetProxy(ctx context.Context, proxy string) error {
	if proxy == "" {
		// Remove proxy
		result := utils.ExecuteCommand(ctx, "pnpm", "config", "delete", "proxy")
		if result.Error != nil {
			return core.NewManagerError("pnpm", "remove proxy", result.Error)
		}
		result = utils.ExecuteCommand(ctx, "pnpm", "config", "delete", "https-proxy")
		if result.Error != nil {
			return core.NewManagerError("pnpm", "remove https-proxy", result.Error)
		}
	} else {
		// Set proxy
		result := utils.ExecuteCommand(ctx, "pnpm", "config", "set", "proxy", proxy)
		if result.Error != nil {
			return core.NewManagerError("pnpm", "set proxy", result.Error)
		}
		result = utils.ExecuteCommand(ctx, "pnpm", "config", "set", "https-proxy", proxy)
		if result.Error != nil {
			return core.NewManagerError("pnpm", "set https-proxy", result.Error)
		}
	}

	p.logger.WithField("proxy", proxy).Info("pnpm proxy updated")
	return nil
}

// GetProjects scans for pnpm projects
func (p *PNPMManager) GetProjects(ctx context.Context, rootPath string) ([]core.Project, error) {
	var projects []core.Project

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if info.Name() == "pnpm-lock.yaml" && utils.IsFile(path) {
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
				Managers:    []string{"pnpm"},
				PackageFile: packageJsonPath,
				LockFile:    path,
				NodeModules: filepath.Join(projectPath, "node_modules"),
			}

			projects = append(projects, project)
		}

		return nil
	})

	if err != nil {
		return nil, core.NewManagerError("pnpm", "scan projects", err)
	}

	return projects, nil
}

// getDefaultStorePath returns the default pnpm store path for the current OS
func (p *PNPMManager) getDefaultStorePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "pnpm", "store")
	case "darwin":
		home, _ := utils.GetHomeDir()
		return filepath.Join(home, "Library", "pnpm", "store")
	default: // linux and others
		home, _ := utils.GetHomeDir()
		return filepath.Join(home, ".local", "share", "pnpm", "store")
	}
}

// getPackagesFromPackageJson fallback method to read packages from package.json
func (p *PNPMManager) getPackagesFromPackageJson(packageJsonPath string) ([]core.Package, error) {
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return nil, core.NewManagerError("pnpm", "read package.json", err)
	}

	var packageJson struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &packageJson); err != nil {
		return nil, core.NewManagerError("pnpm", "parse package.json", err)
	}

	var packages []core.Package
	projectPath := filepath.Dir(packageJsonPath)

	// Add regular dependencies
	for name, version := range packageJson.Dependencies {
		pkg := core.Package{
			Name:     name,
			Version:  version,
			Manager:  "pnpm",
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
			Manager:  "pnpm",
			IsGlobal: false,
			Path:     filepath.Join(projectPath, "node_modules", name),
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}
