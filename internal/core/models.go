package core

import "time"

// CacheInfo 表示包管理器的缓存信息
type CacheInfo struct {
	Manager     string    `json:"manager"`     // 包管理器名称 (npm, pnpm, yarn, bun)
	Path        string    `json:"path"`        // 缓存目录路径
	Size        int64     `json:"size"`        // 缓存大小（字节）
	FileCount   int       `json:"file_count"`  // 缓存中的文件数量
	LastUpdated time.Time `json:"last_updated"` // 最后缓存更新时间
}

// Package 表示一个包
type Package struct {
	Name        string            `json:"name"`        // 包名称
	Version     string            `json:"version"`     // 版本号
	Description string            `json:"description"` // 包描述
	Manager     string            `json:"manager"`     // 包管理器
	IsGlobal    bool              `json:"is_global"`   // 是否为全局包
	Path        string            `json:"path"`        // 包路径
	Size        int64             `json:"size"`        // 包大小
	Dependencies map[string]string `json:"dependencies,omitempty"` // 依赖包
	DevDependencies map[string]string `json:"dev_dependencies,omitempty"` // 开发依赖包
}

// PackageDetail represents detailed package information
type PackageDetail struct {
	Package
	Author      string            `json:"author"`
	License     string            `json:"license"`
	Homepage    string            `json:"homepage"`
	Repository  string            `json:"repository"`
	Keywords    []string          `json:"keywords"`
	Scripts     map[string]string `json:"scripts,omitempty"`
	Engines     map[string]string `json:"engines,omitempty"`
	PeerDependencies map[string]string `json:"peer_dependencies,omitempty"`
}

// Config represents package manager configuration
type Config struct {
	Manager  string            `json:"manager"`
	Registry string            `json:"registry"`
	Proxy    string            `json:"proxy"`
	Settings map[string]string `json:"settings"`
}

// Project represents a project using package managers
type Project struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Managers    []string `json:"managers"`
	PackageFile string   `json:"package_file"`
	LockFile    string   `json:"lock_file"`
	NodeModules string   `json:"node_modules"`
}

// ProjectAnalysis represents detailed project analysis
type ProjectAnalysis struct {
	Project
	PackageCount     int               `json:"package_count"`
	DevPackageCount  int               `json:"dev_package_count"`
	TotalSize        int64             `json:"total_size"`
	OutdatedPackages []Package         `json:"outdated_packages"`
	Vulnerabilities  []Vulnerability   `json:"vulnerabilities"`
	Scripts          map[string]string `json:"scripts"`
}

// DependencyTree represents the dependency tree of a project
type DependencyTree struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies []*DependencyTree `json:"dependencies,omitempty"`
	DevDependency bool             `json:"dev_dependency"`
	Depth        int               `json:"depth"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	Package     string `json:"package"`
	Version     string `json:"version"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	FixedIn     string `json:"fixed_in"`
}

// ManagerType represents the type of package manager
type ManagerType string

const (
	NPM  ManagerType = "npm"
	PNPM ManagerType = "pnpm"
	YARN ManagerType = "yarn"
	BUN  ManagerType = "bun"
)

// String returns the string representation of ManagerType
func (m ManagerType) String() string {
	return string(m)
}

// IsValid checks if the manager type is valid
func (m ManagerType) IsValid() bool {
	switch m {
	case NPM, PNPM, YARN, BUN:
		return true
	default:
		return false
	}
}
