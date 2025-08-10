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

var packagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "Manage and view packages",
	Long: `Manage and view packages for npm, pnpm, yarn, and bun package managers.
	
This command provides functionality to:
- List installed packages in a project or globally
- Search for packages by name
- Show package statistics and information`,
	Aliases: []string{"pkg", "p"},
}

var packagesListCmd = &cobra.Command{
	Use:   "list [project-path]",
	Short: "List installed packages",
	Long: `List installed packages in a project or globally.
	
Examples:
  npm-console packages list                    # List packages in current directory
  npm-console packages list /path/to/project   # List packages in specific project
  npm-console packages list --global           # List global packages`,
	RunE: runPackagesList,
}

var packagesSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for packages",
	Long: `Search for packages by name or description.
	
Examples:
  npm-console packages search react           # Search for packages containing "react"
  npm-console packages search "web framework" # Search with multiple words`,
	Args: cobra.MinimumNArgs(1),
	RunE: runPackagesSearch,
}

var packagesInfoCmd = &cobra.Command{
	Use:   "info <package-name>",
	Short: "Show detailed package information",
	Long: `Show detailed information about a specific package.
	
Examples:
  npm-console packages info react             # Show info for react package
  npm-console packages info @types/node       # Show info for scoped package`,
	Args: cobra.ExactArgs(1),
	RunE: runPackagesInfo,
}

var packagesStatsCmd = &cobra.Command{
	Use:   "stats [project-path]",
	Short: "Show package statistics",
	Long: `Show statistics about packages in a project or globally.
	
Examples:
  npm-console packages stats                  # Show stats for current directory
  npm-console packages stats /path/to/project # Show stats for specific project
  npm-console packages stats --global         # Show global package stats`,
	RunE: runPackagesStats,
}

func init() {
	rootCmd.AddCommand(packagesCmd)
	packagesCmd.AddCommand(packagesListCmd)
	packagesCmd.AddCommand(packagesSearchCmd)
	packagesCmd.AddCommand(packagesInfoCmd)
	packagesCmd.AddCommand(packagesStatsCmd)

	// Add flags
	packagesListCmd.Flags().BoolP("global", "g", false, "List global packages")
	packagesListCmd.Flags().StringP("manager", "m", "", "Filter by specific package manager")
	packagesListCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	
	packagesSearchCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	packagesInfoCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	
	packagesStatsCmd.Flags().BoolP("global", "g", false, "Show global package stats")
	packagesStatsCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
}

func runPackagesList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	packageService := services.NewPackageService()
	
	logger := logger.GetDefault()
	
	global, _ := cmd.Flags().GetBool("global")
	manager, _ := cmd.Flags().GetString("manager")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	var packages []core.Package
	var err error

	if global {
		logger.Debug("Listing global packages")
		if manager != "" {
			packages, err = packageService.GetGlobalPackagesByManager(ctx, manager)
		} else {
			packages, err = packageService.GetGlobalPackages(ctx)
		}
	} else {
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
		
		logger.Debug("Listing project packages", "path", absPath)
		
		if manager != "" {
			packages, err = packageService.GetPackagesByManager(ctx, manager, absPath)
		} else {
			packages, err = packageService.GetAllPackages(ctx, absPath)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to get packages: %w", err)
	}

	if jsonOutput {
		return outputJSON(packages)
	}

	if len(packages) == 0 {
		if global {
			fmt.Println("No global packages found.")
		} else {
			fmt.Println("No packages found in this project.")
		}
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tMANAGER\tTYPE\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-------\t-------\t----\t-----------")

	for _, pkg := range packages {
		pkgType := "local"
		if pkg.IsGlobal {
			pkgType = "global"
		}
		
		description := pkg.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			pkg.Name,
			pkg.Version,
			pkg.Manager,
			pkgType,
			description,
		)
	}

	w.Flush()
	
	fmt.Printf("\nTotal packages: %d\n", len(packages))
	return nil
}

func runPackagesSearch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	packageService := services.NewPackageService()
	
	query := strings.Join(args, " ")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	
	logger := logger.GetDefault()
	logger.Debug("Searching packages", "query", query)

	packages, err := packageService.SearchPackages(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to search packages: %w", err)
	}

	if jsonOutput {
		return outputJSON(packages)
	}

	if len(packages) == 0 {
		fmt.Printf("No packages found matching '%s'.\n", query)
		return nil
	}

	fmt.Printf("Found %d packages matching '%s':\n\n", len(packages), query)

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tMANAGER\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-------\t-------\t-----------")

	for _, pkg := range packages {
		description := pkg.Description
		if len(description) > 60 {
			description = description[:57] + "..."
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			pkg.Name,
			pkg.Version,
			pkg.Manager,
			description,
		)
	}

	w.Flush()
	return nil
}

func runPackagesInfo(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	packageService := services.NewPackageService()
	
	packageName := args[0]
	jsonOutput, _ := cmd.Flags().GetBool("json")
	
	logger := logger.GetDefault()
	logger.Debug("Getting package info", "package", packageName)

	packageInfo, err := packageService.GetPackageInfo(ctx, packageName)
	if err != nil {
		return fmt.Errorf("failed to get package info: %w", err)
	}

	if jsonOutput {
		return outputJSON(packageInfo)
	}

	fmt.Printf("📦 %s\n", packageInfo.Name)
	fmt.Printf("===================\n\n")
	fmt.Printf("Version: %s\n", packageInfo.Version)
	fmt.Printf("Manager: %s\n", packageInfo.Manager)
	
	if packageInfo.Description != "" {
		fmt.Printf("Description: %s\n", packageInfo.Description)
	}
	
	if packageInfo.Author != "" {
		fmt.Printf("Author: %s\n", packageInfo.Author)
	}
	
	if packageInfo.License != "" {
		fmt.Printf("License: %s\n", packageInfo.License)
	}
	
	if packageInfo.Homepage != "" {
		fmt.Printf("Homepage: %s\n", packageInfo.Homepage)
	}
	
	if packageInfo.Repository != "" {
		fmt.Printf("Repository: %s\n", packageInfo.Repository)
	}
	
	if len(packageInfo.Keywords) > 0 {
		fmt.Printf("Keywords: %s\n", strings.Join(packageInfo.Keywords, ", "))
	}
	
	if packageInfo.Path != "" {
		fmt.Printf("Path: %s\n", packageInfo.Path)
	}
	
	if packageInfo.Size > 0 {
		fmt.Printf("Size: %s\n", formatSize(packageInfo.Size))
	}

	return nil
}

func runPackagesStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	packageService := services.NewPackageService()
	
	global, _ := cmd.Flags().GetBool("global")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	
	logger := logger.GetDefault()

	var stats *services.PackageStats
	var err error

	if global {
		logger.Debug("Getting global package stats")
		stats, err = packageService.GetGlobalPackageStats(ctx)
	} else {
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
		
		logger.Debug("Getting project package stats", "path", absPath)
		stats, err = packageService.GetPackageStats(ctx, absPath)
	}

	if err != nil {
		return fmt.Errorf("failed to get package stats: %w", err)
	}

	if jsonOutput {
		return outputJSON(stats)
	}

	fmt.Printf("📊 Package Statistics\n")
	fmt.Printf("====================\n\n")
	fmt.Printf("Total Packages: %d\n", stats.TotalPackages)
	
	if !global {
		fmt.Printf("Local Packages: %d\n", stats.LocalPackages)
		fmt.Printf("Global Packages: %d\n", stats.GlobalPackages)
	}
	
	if len(stats.ByManager) > 0 {
		fmt.Printf("\nBy Package Manager:\n")
		for manager, count := range stats.ByManager {
			fmt.Printf("  %s: %d\n", manager, count)
		}
	}

	return nil
}
