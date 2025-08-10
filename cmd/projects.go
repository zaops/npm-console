package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"npm-console/internal/core"
	"npm-console/internal/services"
	"npm-console/pkg/logger"

	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage and analyze projects",
	Long: `Manage and analyze projects using npm, pnpm, yarn, and bun package managers.
	
This command provides functionality to:
- Scan for projects in a directory tree
- Analyze project dependencies and structure
- Show project statistics`,
	Aliases: []string{"proj", "project"},
}

var projectsScanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan for projects",
	Long: `Scan for projects using any package manager in the specified directory.
	
Examples:
  npm-console projects scan                    # Scan current directory
  npm-console projects scan /path/to/projects  # Scan specific directory
  npm-console projects scan --depth 2         # Limit scan depth`,
	RunE: runProjectsScan,
}

var projectsAnalyzeCmd = &cobra.Command{
	Use:   "analyze [project-path]",
	Short: "Analyze a specific project",
	Long: `Analyze a specific project and show detailed information about dependencies,
size, scripts, and potential issues.
	
Examples:
  npm-console projects analyze                    # Analyze current directory
  npm-console projects analyze /path/to/project   # Analyze specific project`,
	RunE: runProjectsAnalyze,
}

var projectsStatsCmd = &cobra.Command{
	Use:   "stats [path]",
	Short: "Show project statistics",
	Long: `Show statistics about projects found in the specified directory.
	
Examples:
  npm-console projects stats                   # Stats for current directory
  npm-console projects stats /path/to/projects # Stats for specific directory`,
	RunE: runProjectsStats,
}

var projectsDepsCmd = &cobra.Command{
	Use:   "deps [project-path]",
	Short: "Show project dependency tree",
	Long: `Show the dependency tree for a specific project.
	
Examples:
  npm-console projects deps                    # Show deps for current directory
  npm-console projects deps /path/to/project   # Show deps for specific project`,
	RunE: runProjectsDeps,
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(projectsScanCmd)
	projectsCmd.AddCommand(projectsAnalyzeCmd)
	projectsCmd.AddCommand(projectsStatsCmd)
	projectsCmd.AddCommand(projectsDepsCmd)

	// Add flags
	projectsScanCmd.Flags().IntP("depth", "d", 0, "Maximum scan depth (0 = unlimited)")
	projectsScanCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	
	projectsAnalyzeCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	projectsAnalyzeCmd.Flags().BoolP("detailed", "D", false, "Show detailed analysis")
	
	projectsStatsCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	
	projectsDepsCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	projectsDepsCmd.Flags().IntP("depth", "d", 1, "Dependency tree depth")
}

func runProjectsScan(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	projectService := services.NewProjectService()
	
	// Determine scan path
	scanPath := "."
	if len(args) > 0 {
		scanPath = args[0]
	}
	
	// Convert to absolute path
	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve scan path: %w", err)
	}
	
	jsonOutput, _ := cmd.Flags().GetBool("json")
	
	logger := logger.GetDefault()
	logger.Debug("Scanning for projects", "path", absPath)

	projects, err := projectService.ScanProjects(ctx, absPath)
	if err != nil {
		return fmt.Errorf("failed to scan projects: %w", err)
	}

	if jsonOutput {
		return outputJSON(projects)
	}

	if len(projects) == 0 {
		fmt.Printf("No projects found in %s\n", absPath)
		return nil
	}

	fmt.Printf("Found %d projects in %s:\n\n", len(projects), absPath)

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPATH\tMANAGERS\tLOCK FILE")
	fmt.Fprintln(w, "----\t----\t--------\t---------")

	for _, project := range projects {
		name := project.Name
		if name == "" {
			name = filepath.Base(project.Path)
		}
		
		managers := strings.Join(project.Managers, ", ")
		
		lockFile := "None"
		if project.LockFile != "" {
			lockFile = filepath.Base(project.LockFile)
		}
		
		// Shorten path for display
		displayPath := project.Path
		if len(displayPath) > 50 {
			displayPath = "..." + displayPath[len(displayPath)-47:]
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			name,
			displayPath,
			managers,
			lockFile,
		)
	}

	w.Flush()
	return nil
}

func runProjectsAnalyze(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	projectService := services.NewProjectService()
	
	// Determine project path
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}
	
	// Convert to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}
	
	jsonOutput, _ := cmd.Flags().GetBool("json")
	detailed, _ := cmd.Flags().GetBool("detailed")
	
	logger := logger.GetDefault()
	logger.Debug("Analyzing project", "path", absPath)

	analysis, err := projectService.AnalyzeProject(ctx, absPath)
	if err != nil {
		return fmt.Errorf("failed to analyze project: %w", err)
	}

	if jsonOutput {
		return outputJSON(analysis)
	}

	fmt.Printf("📊 Project Analysis: %s\n", analysis.Name)
	fmt.Printf("================================\n\n")
	
	fmt.Printf("Path: %s\n", analysis.Path)
	fmt.Printf("Package Managers: %s\n", strings.Join(analysis.Managers, ", "))
	
	if analysis.LockFile != "" {
		fmt.Printf("Lock File: %s\n", filepath.Base(analysis.LockFile))
	}
	
	fmt.Printf("\n📦 Package Information:\n")
	fmt.Printf("Total Packages: %d\n", analysis.PackageCount)
	if analysis.DevPackageCount > 0 {
		fmt.Printf("Dev Packages: %d\n", analysis.DevPackageCount)
	}
	
	if analysis.TotalSize > 0 {
		fmt.Printf("Total Size: %s\n", formatSize(analysis.TotalSize))
	}
	
	if len(analysis.Scripts) > 0 {
		fmt.Printf("\n📜 Available Scripts:\n")
		for name, script := range analysis.Scripts {
			if detailed {
				fmt.Printf("  %s: %s\n", name, script)
			} else {
				fmt.Printf("  %s\n", name)
			}
		}
	}
	
	if len(analysis.OutdatedPackages) > 0 {
		fmt.Printf("\n⚠️  Outdated Packages: %d\n", len(analysis.OutdatedPackages))
		if detailed {
			for _, pkg := range analysis.OutdatedPackages {
				fmt.Printf("  %s@%s\n", pkg.Name, pkg.Version)
			}
		}
	}
	
	if len(analysis.Vulnerabilities) > 0 {
		fmt.Printf("\n🔒 Security Vulnerabilities: %d\n", len(analysis.Vulnerabilities))
		if detailed {
			for _, vuln := range analysis.Vulnerabilities {
				fmt.Printf("  %s: %s (%s)\n", vuln.Package, vuln.Title, vuln.Severity)
			}
		}
	}

	return nil
}

func runProjectsStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	projectService := services.NewProjectService()
	
	// Determine scan path
	scanPath := "."
	if len(args) > 0 {
		scanPath = args[0]
	}
	
	// Convert to absolute path
	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve scan path: %w", err)
	}
	
	jsonOutput, _ := cmd.Flags().GetBool("json")
	
	logger := logger.GetDefault()
	logger.Debug("Getting project statistics", "path", absPath)

	stats, err := projectService.GetProjectStats(ctx, absPath)
	if err != nil {
		return fmt.Errorf("failed to get project stats: %w", err)
	}

	if jsonOutput {
		return outputJSON(stats)
	}

	fmt.Printf("📊 Project Statistics for %s\n", absPath)
	fmt.Printf("=====================================\n\n")
	fmt.Printf("Total Projects: %d\n", stats.TotalProjects)
	fmt.Printf("Multi-Manager Projects: %d\n", stats.MultiManagerProjects)
	
	if len(stats.ByManager) > 0 {
		fmt.Printf("\nBy Package Manager:\n")
		for manager, count := range stats.ByManager {
			fmt.Printf("  %s: %d\n", manager, count)
		}
	}

	return nil
}

func runProjectsDeps(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	projectService := services.NewProjectService()
	
	// Determine project path
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}
	
	// Convert to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}
	
	jsonOutput, _ := cmd.Flags().GetBool("json")
	maxDepth, _ := cmd.Flags().GetInt("depth")
	
	logger := logger.GetDefault()
	logger.Debug("Getting project dependencies", "path", absPath)

	depTree, err := projectService.GetProjectDependencies(ctx, absPath)
	if err != nil {
		return fmt.Errorf("failed to get project dependencies: %w", err)
	}

	if jsonOutput {
		return outputJSON(depTree)
	}

	fmt.Printf("🌳 Dependency Tree: %s\n", depTree.Name)
	fmt.Printf("==========================\n\n")
	
	printDependencyTree(depTree, "", maxDepth, 0)
	
	return nil
}

// printDependencyTree prints the dependency tree recursively
func printDependencyTree(node *core.DependencyTree, prefix string, maxDepth, currentDepth int) {
	if maxDepth > 0 && currentDepth >= maxDepth {
		return
	}
	
	// Print current node
	marker := "├── "
	if currentDepth == 0 {
		marker = ""
	}
	
	devMarker := ""
	if node.DevDependency {
		devMarker = " (dev)"
	}
	
	fmt.Printf("%s%s%s@%s%s\n", prefix, marker, node.Name, node.Version, devMarker)
	
	// Print children
	if len(node.Dependencies) > 0 && (maxDepth == 0 || currentDepth < maxDepth-1) {
		childPrefix := prefix
		if currentDepth > 0 {
			childPrefix += "│   "
		}
		
		for i, child := range node.Dependencies {
			if i == len(node.Dependencies)-1 {
				// Last child
				fmt.Printf("%s└── %s@%s", childPrefix, child.Name, child.Version)
				if child.DevDependency {
					fmt.Printf(" (dev)")
				}
				fmt.Println()
			} else {
				printDependencyTree(child, childPrefix, maxDepth, currentDepth+1)
			}
		}
	}
}
