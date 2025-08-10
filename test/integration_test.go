package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"npm-console/internal/managers"
	"npm-console/internal/services"
)

func TestIntegration_ManagerFactory(t *testing.T) {
	factory := managers.GetGlobalFactory()
	
	// Test getting all managers
	allManagers := factory.GetAllManagers()
	if len(allManagers) == 0 {
		t.Error("Expected at least one manager to be registered")
	}
	
	expectedManagers := []string{"npm", "pnpm", "yarn", "bun"}
	for _, expected := range expectedManagers {
		if _, exists := allManagers[expected]; !exists {
			t.Errorf("Expected manager %s to be registered", expected)
		}
	}
	
	// Test getting available managers (this will depend on what's installed)
	ctx := context.Background()
	availableManagers := factory.GetAvailableManagers(ctx)
	t.Logf("Available managers: %d", len(availableManagers))
	
	// Test getting specific manager
	npmManager, err := factory.GetManager("npm")
	if err != nil {
		t.Errorf("Failed to get npm manager: %v", err)
	}
	
	if npmManager.Name() != "npm" {
		t.Errorf("Expected manager name 'npm', got '%s'", npmManager.Name())
	}
}

func TestIntegration_CacheService(t *testing.T) {
	cacheService := services.NewCacheService()
	ctx := context.Background()
	
	// Test getting all cache info
	cacheInfos, err := cacheService.GetAllCacheInfo(ctx)
	if err != nil {
		t.Errorf("Failed to get cache info: %v", err)
		return
	}
	
	t.Logf("Found cache info for %d managers", len(cacheInfos))
	
	// Test getting cache summary
	summary, err := cacheService.GetCacheSummary(ctx)
	if err != nil {
		t.Errorf("Failed to get cache summary: %v", err)
		return
	}
	
	t.Logf("Total cache size: %d bytes", summary.TotalSize)
	t.Logf("Total files: %d", summary.TotalFiles)
	t.Logf("Manager count: %d", summary.ManagerCount)
	
	// Test getting total cache size
	totalSize, err := cacheService.GetTotalCacheSize(ctx)
	if err != nil {
		t.Errorf("Failed to get total cache size: %v", err)
		return
	}
	
	if totalSize != summary.TotalSize {
		t.Errorf("Total cache size mismatch: %d vs %d", totalSize, summary.TotalSize)
	}
}

func TestIntegration_PackageService(t *testing.T) {
	packageService := services.NewPackageService()
	ctx := context.Background()
	
	// Test getting global packages
	globalPackages, err := packageService.GetGlobalPackages(ctx)
	if err != nil {
		t.Errorf("Failed to get global packages: %v", err)
		return
	}
	
	t.Logf("Found %d global packages", len(globalPackages))
	
	// Test getting global package stats
	stats, err := packageService.GetGlobalPackageStats(ctx)
	if err != nil {
		t.Errorf("Failed to get global package stats: %v", err)
		return
	}
	
	if stats.TotalPackages != len(globalPackages) {
		t.Errorf("Package count mismatch: %d vs %d", stats.TotalPackages, len(globalPackages))
	}
	
	t.Logf("Package stats: %+v", stats)
}

func TestIntegration_ConfigService(t *testing.T) {
	configService := services.NewConfigService()
	ctx := context.Background()
	
	// Test getting all configs
	configs, err := configService.GetAllConfigs(ctx)
	if err != nil {
		t.Errorf("Failed to get configs: %v", err)
		return
	}
	
	t.Logf("Found configs for %d managers", len(configs))
	
	// Test getting config summary
	summary, err := configService.GetConfigSummary(ctx)
	if err != nil {
		t.Errorf("Failed to get config summary: %v", err)
		return
	}
	
	if summary.ManagerCount != len(configs) {
		t.Errorf("Manager count mismatch: %d vs %d", summary.ManagerCount, len(configs))
	}
	
	t.Logf("Config summary: %+v", summary)
	
	// Test registry URL validation
	validURLs := []string{
		"https://registry.npmjs.org/",
		"http://localhost:4873/",
		"https://registry.npmmirror.com/",
	}
	
	for _, url := range validURLs {
		err := configService.ValidateRegistryURL(url)
		if err != nil {
			t.Errorf("Valid URL %s should not return error: %v", url, err)
		}
	}
	
	// Test invalid registry URLs
	invalidURLs := []string{
		"",
		"not-a-url",
		"ftp://invalid.com/",
		"registry.npmjs.org", // missing scheme
	}
	
	for _, url := range invalidURLs {
		err := configService.ValidateRegistryURL(url)
		if err == nil {
			t.Errorf("Invalid URL %s should return error", url)
		}
	}
}

func TestIntegration_ProjectService(t *testing.T) {
	projectService := services.NewProjectService()
	ctx := context.Background()
	
	// Create a temporary test project
	tempDir := t.TempDir()
	packageJsonPath := filepath.Join(tempDir, "package.json")
	
	packageJsonContent := `{
  "name": "test-project",
  "version": "1.0.0",
  "description": "Test project for integration testing",
  "dependencies": {
    "react": "^18.0.0",
    "lodash": "^4.17.21"
  },
  "devDependencies": {
    "typescript": "^4.9.0",
    "@types/react": "^18.0.0"
  },
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "start": "node index.js"
  }
}`
	
	err := os.WriteFile(packageJsonPath, []byte(packageJsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test package.json: %v", err)
	}
	
	// Test project scanning
	projects, err := projectService.ScanProjects(ctx, tempDir)
	if err != nil {
		t.Errorf("Failed to scan projects: %v", err)
		return
	}
	
	if len(projects) == 0 {
		t.Error("Expected to find at least one project")
		return
	}
	
	project := projects[0]
	if project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", project.Name)
	}
	
	// Test project analysis
	analysis, err := projectService.AnalyzeProject(ctx, tempDir)
	if err != nil {
		t.Errorf("Failed to analyze project: %v", err)
		return
	}
	
	if analysis.Name != "test-project" {
		t.Errorf("Expected analysis name 'test-project', got '%s'", analysis.Name)
	}
	
	if len(analysis.Scripts) != 3 {
		t.Errorf("Expected 3 scripts, got %d", len(analysis.Scripts))
	}
	
	// Test dependency tree
	depTree, err := projectService.GetProjectDependencies(ctx, tempDir)
	if err != nil {
		t.Errorf("Failed to get project dependencies: %v", err)
		return
	}
	
	if depTree.Name != "test-project" {
		t.Errorf("Expected dependency tree name 'test-project', got '%s'", depTree.Name)
	}
	
	if len(depTree.Dependencies) != 4 { // 2 deps + 2 devDeps
		t.Errorf("Expected 4 dependencies, got %d", len(depTree.Dependencies))
	}
	
	// Test project stats
	stats, err := projectService.GetProjectStats(ctx, tempDir)
	if err != nil {
		t.Errorf("Failed to get project stats: %v", err)
		return
	}
	
	if stats.TotalProjects != 1 {
		t.Errorf("Expected 1 project, got %d", stats.TotalProjects)
	}
	
	t.Logf("Project stats: %+v", stats)
}

func TestIntegration_EndToEnd(t *testing.T) {
	// This test simulates a complete workflow
	ctx := context.Background()
	
	// 1. Get available managers
	factory := managers.GetGlobalFactory()
	availableManagers := factory.GetAvailableManagers(ctx)
	
	if len(availableManagers) == 0 {
		t.Skip("No package managers available for end-to-end test")
	}
	
	t.Logf("Available managers: %v", factory.GetAvailableManagerNames(ctx))
	
	// 2. Get cache information
	cacheService := services.NewCacheService()
	cacheInfos, err := cacheService.GetAllCacheInfo(ctx)
	if err != nil {
		t.Errorf("Failed to get cache info: %v", err)
	}
	
	// 3. Get global packages
	packageService := services.NewPackageService()
	globalPackages, err := packageService.GetGlobalPackages(ctx)
	if err != nil {
		t.Errorf("Failed to get global packages: %v", err)
	}
	
	// 4. Get configurations
	configService := services.NewConfigService()
	configs, err := configService.GetAllConfigs(ctx)
	if err != nil {
		t.Errorf("Failed to get configs: %v", err)
	}
	
	// 5. Scan for projects in current directory
	projectService := services.NewProjectService()
	projects, err := projectService.ScanProjects(ctx, ".")
	if err != nil {
		t.Errorf("Failed to scan projects: %v", err)
	}
	
	// Log summary
	t.Logf("End-to-end test summary:")
	t.Logf("  Available managers: %d", len(availableManagers))
	t.Logf("  Cache entries: %d", len(cacheInfos))
	t.Logf("  Global packages: %d", len(globalPackages))
	t.Logf("  Configurations: %d", len(configs))
	t.Logf("  Projects found: %d", len(projects))
	
	// Verify basic consistency
	if len(configs) != len(availableManagers) {
		t.Logf("Note: Config count (%d) doesn't match available managers (%d) - this may be expected", 
			len(configs), len(availableManagers))
	}
}
