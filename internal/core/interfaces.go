package core

import "context"

// PackageManager defines the interface for different package managers
type PackageManager interface {
	// Name returns the name of the package manager
	Name() string
	
	// IsAvailable checks if the package manager is installed and available
	IsAvailable(ctx context.Context) bool
	
	// GetCacheInfo returns information about the cache
	GetCacheInfo(ctx context.Context) (*CacheInfo, error)
	
	// ClearCache clears the package manager's cache
	ClearCache(ctx context.Context) error
	
	// GetInstalledPackages returns packages installed in a specific project
	GetInstalledPackages(ctx context.Context, projectPath string) ([]Package, error)
	
	// GetGlobalPackages returns globally installed packages
	GetGlobalPackages(ctx context.Context) ([]Package, error)
	
	// GetConfig returns the current configuration
	GetConfig(ctx context.Context) (*Config, error)
	
	// SetRegistry sets the registry URL
	SetRegistry(ctx context.Context, url string) error
	
	// SetProxy sets the proxy configuration
	SetProxy(ctx context.Context, proxy string) error
	
	// GetProjects scans for projects using this package manager
	GetProjects(ctx context.Context, rootPath string) ([]Project, error)
}

// CacheService defines the interface for cache management
type CacheService interface {
	GetAllCacheInfo(ctx context.Context) ([]CacheInfo, error)
	ClearAllCaches(ctx context.Context) error
	GetTotalCacheSize(ctx context.Context) (int64, error)
}

// PackageService defines the interface for package management
type PackageService interface {
	GetAllPackages(ctx context.Context, projectPath string) ([]Package, error)
	SearchPackages(ctx context.Context, query string) ([]Package, error)
	GetPackageInfo(ctx context.Context, packageName string) (*PackageDetail, error)
}

// ConfigService defines the interface for configuration management
type ConfigService interface {
	GetAllConfigs(ctx context.Context) ([]Config, error)
	SetRegistry(ctx context.Context, manager string, url string) error
	SetProxy(ctx context.Context, manager string, proxy string) error
	TestRegistry(ctx context.Context, manager string, url string) error
}

// ProjectService defines the interface for project management
type ProjectService interface {
	ScanProjects(ctx context.Context, rootPath string) ([]Project, error)
	AnalyzeProject(ctx context.Context, projectPath string) (*ProjectAnalysis, error)
	GetProjectDependencies(ctx context.Context, projectPath string) (*DependencyTree, error)
}
