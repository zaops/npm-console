package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"npm-console/internal/core"
	"npm-console/internal/managers"
	"npm-console/pkg/logger"
	"npm-console/pkg/utils"
)

// ProjectService implements project management functionality
type ProjectService struct {
	factory *managers.ManagerFactory
	logger  *logger.Logger
}

// NewProjectService creates a new project service
func NewProjectService() *ProjectService {
	return &ProjectService{
		factory: managers.GetGlobalFactory(),
		logger:  logger.GetDefault().WithField("service", "projects"),
	}
}

// ScanProjects scans for projects using any package manager in the given root path
func (s *ProjectService) ScanProjects(ctx context.Context, rootPath string) ([]core.Project, error) {
	if rootPath == "" {
		return nil, core.NewValidationError("rootPath", rootPath, "root path cannot be empty")
	}
	
	// Expand and validate path
	expandedPath, err := utils.ExpandPath(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}
	
	if !utils.PathExists(expandedPath) {
		return nil, core.NewValidationError("rootPath", rootPath, "path does not exist")
	}
	
	if !utils.IsDir(expandedPath) {
		return nil, core.NewValidationError("rootPath", rootPath, "path is not a directory")
	}
	
	availableManagers := s.factory.GetAvailableManagers(ctx)
	
	var allProjects []core.Project
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errors []error

	// Scan projects concurrently with all managers
	for name, manager := range availableManagers {
		wg.Add(1)
		go func(name string, mgr core.PackageManager) {
			defer wg.Done()
			
			projects, err := mgr.GetProjects(ctx, expandedPath)
			if err != nil {
				s.logger.WithError(err).WithField("manager", name).Warn("Failed to scan projects")
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to scan projects with %s: %w", name, err))
				mu.Unlock()
				return
			}
			
			mu.Lock()
			allProjects = append(allProjects, projects...)
			mu.Unlock()
		}(name, manager)
	}
	
	wg.Wait()
	
	// Merge projects that use multiple managers
	mergedProjects := s.mergeProjects(allProjects)
	
	// Sort by project path for consistent output
	sort.Slice(mergedProjects, func(i, j int) bool {
		return mergedProjects[i].Path < mergedProjects[j].Path
	})
	
	// Log any errors but don't fail the entire operation
	if len(errors) > 0 {
		for _, err := range errors {
			s.logger.WithError(err).Warn("Project scanning error")
		}
	}
	
	s.logger.WithField("project_count", len(mergedProjects)).WithField("scan_path", expandedPath).Info("Project scan completed")
	
	return mergedProjects, nil
}

// AnalyzeProject analyzes a specific project and returns detailed information
func (s *ProjectService) AnalyzeProject(ctx context.Context, projectPath string) (*core.ProjectAnalysis, error) {
	if projectPath == "" {
		return nil, core.NewValidationError("projectPath", projectPath, "project path cannot be empty")
	}
	
	// Expand and validate path
	expandedPath, err := utils.ExpandPath(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}
	
	if !utils.PathExists(expandedPath) {
		return nil, core.NewValidationError("projectPath", projectPath, "path does not exist")
	}
	
	// Check if package.json exists
	packageJsonPath := filepath.Join(expandedPath, "package.json")
	if !utils.IsFile(packageJsonPath) {
		return nil, core.ErrProjectNotFound
	}
	
	// Read package.json
	packageJson, err := s.readPackageJson(packageJsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}
	
	// Detect which package managers are used
	managers := s.detectProjectManagers(expandedPath)
	
	// Get package information
	packageService := NewPackageService()
	packages, err := packageService.GetAllPackages(ctx, expandedPath)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get packages for project analysis")
		packages = []core.Package{} // Continue with empty packages
	}
	
	// Calculate statistics
	var devPackageCount int
	var totalSize int64
	
	for _, pkg := range packages {
		if pkg.Size > 0 {
			totalSize += pkg.Size
		}
		// Note: We don't have a direct way to determine if a package is dev dependency
		// This would require parsing package.json dependencies vs devDependencies
	}
	
	// Get node_modules size if it exists
	nodeModulesPath := filepath.Join(expandedPath, "node_modules")
	if utils.IsDir(nodeModulesPath) {
		nodeModulesSize, err := utils.GetDirSize(nodeModulesPath)
		if err == nil {
			totalSize = nodeModulesSize // Use actual node_modules size
		}
	}
	
	analysis := &core.ProjectAnalysis{
		Project: core.Project{
			Name:        packageJson.Name,
			Path:        expandedPath,
			Managers:    managers,
			PackageFile: packageJsonPath,
			NodeModules: nodeModulesPath,
		},
		PackageCount:     len(packages),
		DevPackageCount:  devPackageCount,
		TotalSize:        totalSize,
		OutdatedPackages: []core.Package{}, // TODO: Implement outdated package detection
		Vulnerabilities:  []core.Vulnerability{}, // TODO: Implement vulnerability scanning
		Scripts:          packageJson.Scripts,
	}
	
	// Set lock file based on detected managers
	for _, manager := range managers {
		switch manager {
		case "npm":
			lockFile := filepath.Join(expandedPath, "package-lock.json")
			if utils.IsFile(lockFile) {
				analysis.LockFile = lockFile
				break
			}
		case "pnpm":
			lockFile := filepath.Join(expandedPath, "pnpm-lock.yaml")
			if utils.IsFile(lockFile) {
				analysis.LockFile = lockFile
				break
			}
		case "yarn":
			lockFile := filepath.Join(expandedPath, "yarn.lock")
			if utils.IsFile(lockFile) {
				analysis.LockFile = lockFile
				break
			}
		case "bun":
			lockFile := filepath.Join(expandedPath, "bun.lockb")
			if utils.IsFile(lockFile) {
				analysis.LockFile = lockFile
				break
			}
		}
	}
	
	return analysis, nil
}

// GetProjectDependencies returns the dependency tree for a project
func (s *ProjectService) GetProjectDependencies(ctx context.Context, projectPath string) (*core.DependencyTree, error) {
	if projectPath == "" {
		return nil, core.NewValidationError("projectPath", projectPath, "project path cannot be empty")
	}
	
	// For now, return a basic dependency tree based on package.json
	// In the future, this could be enhanced to build a full dependency tree
	packageJsonPath := filepath.Join(projectPath, "package.json")
	if !utils.IsFile(packageJsonPath) {
		return nil, core.ErrProjectNotFound
	}
	
	packageJson, err := s.readPackageJson(packageJsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}
	
	// Create root node
	root := &core.DependencyTree{
		Name:    packageJson.Name,
		Version: packageJson.Version,
		Depth:   0,
	}
	
	// Add direct dependencies
	for name, version := range packageJson.Dependencies {
		dep := &core.DependencyTree{
			Name:          name,
			Version:       version,
			DevDependency: false,
			Depth:         1,
		}
		root.Dependencies = append(root.Dependencies, dep)
	}
	
	// Add dev dependencies
	for name, version := range packageJson.DevDependencies {
		dep := &core.DependencyTree{
			Name:          name,
			Version:       version,
			DevDependency: true,
			Depth:         1,
		}
		root.Dependencies = append(root.Dependencies, dep)
	}
	
	return root, nil
}

// GetProjectStats returns statistics about scanned projects
func (s *ProjectService) GetProjectStats(ctx context.Context, rootPath string) (*ProjectStats, error) {
	projects, err := s.ScanProjects(ctx, rootPath)
	if err != nil {
		return nil, err
	}
	
	stats := &ProjectStats{
		ByManager: make(map[string]int),
	}
	
	for _, project := range projects {
		stats.TotalProjects++
		
		for _, manager := range project.Managers {
			stats.ByManager[manager]++
		}
		
		if len(project.Managers) > 1 {
			stats.MultiManagerProjects++
		}
	}
	
	return stats, nil
}

// mergeProjects merges projects that have the same path but different managers
func (s *ProjectService) mergeProjects(projects []core.Project) []core.Project {
	projectMap := make(map[string]*core.Project)
	
	for _, project := range projects {
		if existing, exists := projectMap[project.Path]; exists {
			// Merge managers
			for _, manager := range project.Managers {
				found := false
				for _, existingManager := range existing.Managers {
					if existingManager == manager {
						found = true
						break
					}
				}
				if !found {
					existing.Managers = append(existing.Managers, manager)
				}
			}
			
			// Update lock file if not set
			if existing.LockFile == "" && project.LockFile != "" {
				existing.LockFile = project.LockFile
			}
		} else {
			// Create a copy to avoid modifying the original
			projectCopy := project
			projectMap[project.Path] = &projectCopy
		}
	}
	
	// Convert map back to slice
	var merged []core.Project
	for _, project := range projectMap {
		merged = append(merged, *project)
	}
	
	return merged
}

// detectProjectManagers detects which package managers are used in a project
func (s *ProjectService) detectProjectManagers(projectPath string) []string {
	var managers []string
	
	// Check for npm (package-lock.json)
	if utils.IsFile(filepath.Join(projectPath, "package-lock.json")) {
		managers = append(managers, "npm")
	}
	
	// Check for pnpm (pnpm-lock.yaml)
	if utils.IsFile(filepath.Join(projectPath, "pnpm-lock.yaml")) {
		managers = append(managers, "pnpm")
	}
	
	// Check for yarn (yarn.lock)
	if utils.IsFile(filepath.Join(projectPath, "yarn.lock")) {
		managers = append(managers, "yarn")
	}
	
	// Check for bun (bun.lockb)
	if utils.IsFile(filepath.Join(projectPath, "bun.lockb")) {
		managers = append(managers, "bun")
	}
	
	// If no lock files found, assume npm as default if package.json exists
	if len(managers) == 0 && utils.IsFile(filepath.Join(projectPath, "package.json")) {
		managers = append(managers, "npm")
	}
	
	return managers
}

// readPackageJson reads and parses package.json file
func (s *ProjectService) readPackageJson(packageJsonPath string) (*PackageJsonInfo, error) {
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return nil, err
	}
	
	var packageJson PackageJsonInfo
	if err := json.Unmarshal(data, &packageJson); err != nil {
		return nil, err
	}
	
	return &packageJson, nil
}

// PackageJsonInfo represents relevant information from package.json
type PackageJsonInfo struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

// ProjectStats represents project statistics
type ProjectStats struct {
	TotalProjects        int            `json:"total_projects"`
	MultiManagerProjects int            `json:"multi_manager_projects"`
	ByManager            map[string]int `json:"by_manager"`
}
