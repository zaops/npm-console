package core

import (
	"testing"
	"time"
)

func TestCacheInfo(t *testing.T) {
	tests := []struct {
		name     string
		cache    CacheInfo
		expected bool
	}{
		{
			name: "valid cache info",
			cache: CacheInfo{
				Manager:     "npm",
				Path:        "/path/to/cache",
				Size:        1024,
				FileCount:   10,
				LastUpdated: time.Now(),
			},
			expected: true,
		},
		{
			name: "empty manager",
			cache: CacheInfo{
				Manager:     "",
				Path:        "/path/to/cache",
				Size:        1024,
				FileCount:   10,
				LastUpdated: time.Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.cache.Manager != ""
			if valid != tt.expected {
				t.Errorf("CacheInfo validation = %v, want %v", valid, tt.expected)
			}
		})
	}
}

func TestPackage(t *testing.T) {
	tests := []struct {
		name     string
		pkg      Package
		expected bool
	}{
		{
			name: "valid package",
			pkg: Package{
				Name:     "react",
				Version:  "18.2.0",
				Manager:  "npm",
				IsGlobal: false,
			},
			expected: true,
		},
		{
			name: "empty name",
			pkg: Package{
				Name:     "",
				Version:  "18.2.0",
				Manager:  "npm",
				IsGlobal: false,
			},
			expected: false,
		},
		{
			name: "empty version",
			pkg: Package{
				Name:     "react",
				Version:  "",
				Manager:  "npm",
				IsGlobal: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.pkg.Name != "" && tt.pkg.Version != ""
			if valid != tt.expected {
				t.Errorf("Package validation = %v, want %v", valid, tt.expected)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name: "valid config",
			config: Config{
				Manager:  "npm",
				Registry: "https://registry.npmjs.org/",
				Proxy:    "",
				Settings: map[string]string{"cache": "/tmp/cache"},
			},
			expected: true,
		},
		{
			name: "empty manager",
			config: Config{
				Manager:  "",
				Registry: "https://registry.npmjs.org/",
				Proxy:    "",
				Settings: map[string]string{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.config.Manager != ""
			if valid != tt.expected {
				t.Errorf("Config validation = %v, want %v", valid, tt.expected)
			}
		})
	}
}

func TestProject(t *testing.T) {
	tests := []struct {
		name     string
		project  Project
		expected bool
	}{
		{
			name: "valid project",
			project: Project{
				Name:        "my-project",
				Path:        "/path/to/project",
				Managers:    []string{"npm", "yarn"},
				PackageFile: "/path/to/project/package.json",
			},
			expected: true,
		},
		{
			name: "empty path",
			project: Project{
				Name:        "my-project",
				Path:        "",
				Managers:    []string{"npm"},
				PackageFile: "/path/to/project/package.json",
			},
			expected: false,
		},
		{
			name: "no managers",
			project: Project{
				Name:        "my-project",
				Path:        "/path/to/project",
				Managers:    []string{},
				PackageFile: "/path/to/project/package.json",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.project.Path != "" && len(tt.project.Managers) > 0
			if valid != tt.expected {
				t.Errorf("Project validation = %v, want %v", valid, tt.expected)
			}
		})
	}
}

func TestDependencyTree(t *testing.T) {
	root := &DependencyTree{
		Name:    "my-app",
		Version: "1.0.0",
		Depth:   0,
		Dependencies: []*DependencyTree{
			{
				Name:          "react",
				Version:       "18.2.0",
				DevDependency: false,
				Depth:         1,
			},
			{
				Name:          "typescript",
				Version:       "4.9.0",
				DevDependency: true,
				Depth:         1,
			},
		},
	}

	if root.Name != "my-app" {
		t.Errorf("Root name = %v, want %v", root.Name, "my-app")
	}

	if len(root.Dependencies) != 2 {
		t.Errorf("Dependencies count = %v, want %v", len(root.Dependencies), 2)
	}

	// Test dev dependency flag
	var devDep *DependencyTree
	for _, dep := range root.Dependencies {
		if dep.DevDependency {
			devDep = dep
			break
		}
	}

	if devDep == nil {
		t.Error("Expected to find dev dependency")
	} else if devDep.Name != "typescript" {
		t.Errorf("Dev dependency name = %v, want %v", devDep.Name, "typescript")
	}
}
