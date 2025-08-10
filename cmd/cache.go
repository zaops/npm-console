package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"npm-console/internal/services"
	"npm-console/pkg/logger"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage package manager caches",
	Long: `Manage caches for npm, pnpm, yarn, and bun package managers.
	
This command provides functionality to:
- List cache information for all package managers
- Clean caches for specific or all package managers
- Show cache statistics and summaries`,
}

var cacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cache information for all package managers",
	Long: `Display cache information including size, file count, and location 
for all available package managers.`,
	RunE: runCacheList,
}

var cacheCleanCmd = &cobra.Command{
	Use:   "clean [manager]",
	Short: "Clean cache for specific manager or all managers",
	Long: `Clean cache for a specific package manager or all available managers.
	
Examples:
  npm-console cache clean          # Clean all caches
  npm-console cache clean npm      # Clean only npm cache
  npm-console cache clean pnpm     # Clean only pnpm cache`,
	RunE: runCacheClean,
}

var cacheInfoCmd = &cobra.Command{
	Use:   "info [manager]",
	Short: "Show detailed cache information",
	Long: `Show detailed cache information for a specific manager or summary for all managers.
	
Examples:
  npm-console cache info           # Show summary for all managers
  npm-console cache info npm       # Show detailed info for npm`,
	RunE: runCacheInfo,
}

var cacheSizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Show total cache size across all managers",
	Long:  `Display the total cache size across all package managers.`,
	RunE:  runCacheSize,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheCleanCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
	cacheCmd.AddCommand(cacheSizeCmd)

	// Add flags
	cacheCleanCmd.Flags().BoolP("force", "f", false, "Force clean without confirmation")
	cacheListCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cacheInfoCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
}

func runCacheList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cacheService := services.NewCacheService()
	
	logger := logger.GetDefault()
	logger.Debug("Listing cache information")

	cacheInfos, err := cacheService.GetAllCacheInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cache information: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		return outputJSON(cacheInfos)
	}

	if len(cacheInfos) == 0 {
		fmt.Println("No package managers found or available.")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MANAGER\tSIZE\tFILES\tPATH\tLAST UPDATED")
	fmt.Fprintln(w, "-------\t----\t-----\t----\t------------")

	for _, info := range cacheInfos {
		size := formatSize(info.Size)
		lastUpdated := "Never"
		if !info.LastUpdated.IsZero() {
			lastUpdated = info.LastUpdated.Format("2006-01-02 15:04")
		}
		
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			info.Manager,
			size,
			info.FileCount,
			info.Path,
			lastUpdated,
		)
	}

	w.Flush()
	return nil
}

func runCacheClean(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cacheService := services.NewCacheService()
	
	force, _ := cmd.Flags().GetBool("force")
	
	if len(args) == 0 {
		// Clean all caches
		if !force {
			fmt.Print("This will clean all package manager caches. Continue? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Cache cleaning cancelled.")
				return nil
			}
		}

		fmt.Println("Cleaning all caches...")
		err := cacheService.ClearAllCaches(ctx)
		if err != nil {
			return fmt.Errorf("failed to clean caches: %w", err)
		}
		
		fmt.Println("✅ All caches cleaned successfully!")
		return nil
	}

	// Clean specific manager cache
	managerName := args[0]
	
	if !force {
		fmt.Printf("This will clean the %s cache. Continue? (y/N): ", managerName)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cache cleaning cancelled.")
			return nil
		}
	}

	fmt.Printf("Cleaning %s cache...\n", managerName)
	err := cacheService.ClearCache(ctx, managerName)
	if err != nil {
		return fmt.Errorf("failed to clean %s cache: %w", managerName, err)
	}
	
	fmt.Printf("✅ %s cache cleaned successfully!\n", managerName)
	return nil
}

func runCacheInfo(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cacheService := services.NewCacheService()
	
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if len(args) == 0 {
		// Show summary for all managers
		summary, err := cacheService.GetCacheSummary(ctx)
		if err != nil {
			return fmt.Errorf("failed to get cache summary: %w", err)
		}

		if jsonOutput {
			return outputJSON(summary)
		}

		fmt.Printf("📊 Cache Summary\n")
		fmt.Printf("================\n\n")
		fmt.Printf("Total Size: %s\n", formatSize(summary.TotalSize))
		fmt.Printf("Total Files: %d\n", summary.TotalFiles)
		fmt.Printf("Managers: %d\n", summary.ManagerCount)
		
		if summary.LargestCacheManager != "" {
			fmt.Printf("Largest Cache: %s (%s)\n", 
				summary.LargestCacheManager, 
				formatSize(summary.LargestCacheSize))
		}
		
		fmt.Printf("\nAvailable Managers: %s\n", strings.Join(summary.AvailableManagers, ", "))
		
		return nil
	}

	// Show detailed info for specific manager
	managerName := args[0]
	cacheInfo, err := cacheService.GetCacheInfo(ctx, managerName)
	if err != nil {
		return fmt.Errorf("failed to get cache info for %s: %w", managerName, err)
	}

	if jsonOutput {
		return outputJSON(cacheInfo)
	}

	fmt.Printf("📦 %s Cache Information\n", strings.ToUpper(cacheInfo.Manager))
	fmt.Printf("========================\n\n")
	fmt.Printf("Path: %s\n", cacheInfo.Path)
	fmt.Printf("Size: %s\n", formatSize(cacheInfo.Size))
	fmt.Printf("Files: %d\n", cacheInfo.FileCount)
	
	if !cacheInfo.LastUpdated.IsZero() {
		fmt.Printf("Last Updated: %s\n", cacheInfo.LastUpdated.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("Last Updated: Never\n")
	}

	return nil
}

func runCacheSize(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cacheService := services.NewCacheService()

	totalSize, err := cacheService.GetTotalCacheSize(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total cache size: %w", err)
	}

	fmt.Printf("Total cache size: %s\n", formatSize(totalSize))
	return nil
}

// formatSize formats bytes into human readable format
func formatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
