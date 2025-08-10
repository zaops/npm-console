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

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage package registry settings",
	Long: `Manage package registry settings for npm, pnpm, yarn, and bun package managers.
	
This command provides functionality to:
- List current registry configurations
- Set registry URLs for specific or all package managers
- Test registry connectivity`,
	Aliases: []string{"reg", "r"},
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current registry configurations",
	Long:  `Display current registry configurations for all available package managers.`,
	RunE:  runRegistryList,
}

var registrySetCmd = &cobra.Command{
	Use:   "set <registry-url> [manager]",
	Short: "Set registry URL",
	Long: `Set registry URL for a specific package manager or all managers.
	
Examples:
  npm-console registry set https://registry.npmjs.org/           # Set for all managers
  npm-console registry set https://registry.npmjs.org/ npm      # Set for npm only
  npm-console registry set https://registry.npmmirror.com/      # Use npm mirror`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRegistrySet,
}

var registryTestCmd = &cobra.Command{
	Use:   "test [registry-url] [manager]",
	Short: "Test registry connectivity",
	Long: `Test connectivity to a registry URL.
	
Examples:
  npm-console registry test                                     # Test current registries
  npm-console registry test https://registry.npmjs.org/        # Test specific URL
  npm-console registry test https://registry.npmjs.org/ npm    # Test URL for npm`,
	RunE: runRegistryTest,
}

var registryResetCmd = &cobra.Command{
	Use:   "reset [manager]",
	Short: "Reset registry to default",
	Long: `Reset registry to default npm registry for specific or all managers.
	
Examples:
  npm-console registry reset          # Reset all managers to default
  npm-console registry reset npm      # Reset npm to default`,
	RunE: runRegistryReset,
}

func init() {
	rootCmd.AddCommand(registryCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registrySetCmd)
	registryCmd.AddCommand(registryTestCmd)
	registryCmd.AddCommand(registryResetCmd)

	// Add flags
	registryListCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	registrySetCmd.Flags().BoolP("all", "a", false, "Set for all available managers")
	registryTestCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	registryResetCmd.Flags().BoolP("all", "a", false, "Reset all managers")
	registryResetCmd.Flags().BoolP("force", "f", false, "Force reset without confirmation")
}

func runRegistryList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	configService := services.NewConfigService()
	
	logger := logger.GetDefault()
	logger.Debug("Listing registry configurations")

	configs, err := configService.GetAllConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get registry configurations: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		return outputJSON(configs)
	}

	if len(configs) == 0 {
		fmt.Println("No package managers found or available.")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MANAGER\tREGISTRY\tPROXY")
	fmt.Fprintln(w, "-------\t--------\t-----")

	for _, config := range configs {
		registry := config.Registry
		if registry == "" {
			registry = "(not set)"
		}
		
		proxy := config.Proxy
		if proxy == "" {
			proxy = "(none)"
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			config.Manager,
			registry,
			proxy,
		)
	}

	w.Flush()
	return nil
}

func runRegistrySet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	configService := services.NewConfigService()
	
	registryURL := args[0]
	setAll, _ := cmd.Flags().GetBool("all")
	
	logger := logger.GetDefault()

	if len(args) > 1 && !setAll {
		// Set for specific manager
		managerName := args[1]
		logger.Debug("Setting registry for specific manager", "manager", managerName, "registry", registryURL)
		
		err := configService.SetRegistry(ctx, managerName, registryURL)
		if err != nil {
			return fmt.Errorf("failed to set registry for %s: %w", managerName, err)
		}
		
		fmt.Printf("✅ Registry set for %s: %s\n", managerName, registryURL)
		return nil
	}

	// Set for all managers
	logger.Debug("Setting registry for all managers", "registry", registryURL)
	
	err := configService.SetRegistryForAll(ctx, registryURL)
	if err != nil {
		return fmt.Errorf("failed to set registry for all managers: %w", err)
	}
	
	fmt.Printf("✅ Registry set for all managers: %s\n", registryURL)
	return nil
}

func runRegistryTest(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	configService := services.NewConfigService()
	
	jsonOutput, _ := cmd.Flags().GetBool("json")
	logger := logger.GetDefault()

	if len(args) == 0 {
		// Test current registries
		logger.Debug("Testing current registries")
		
		configs, err := configService.GetAllConfigs(ctx)
		if err != nil {
			return fmt.Errorf("failed to get registry configurations: %w", err)
		}

		var results []RegistryTestResult
		
		for _, config := range configs {
			if config.Registry == "" {
				continue
			}
			
			err := configService.TestRegistry(ctx, config.Manager, config.Registry)
			result := RegistryTestResult{
				Manager:  config.Manager,
				Registry: config.Registry,
				Success:  err == nil,
			}
			if err != nil {
				result.Error = err.Error()
			}
			results = append(results, result)
		}

		if jsonOutput {
			return outputJSON(results)
		}

		if len(results) == 0 {
			fmt.Println("No registries configured to test.")
			return nil
		}

		fmt.Println("Registry Test Results:")
		fmt.Println("=====================")
		
		for _, result := range results {
			status := "✅ PASS"
			if !result.Success {
				status = "❌ FAIL"
			}
			
			fmt.Printf("%s %s: %s\n", status, result.Manager, result.Registry)
			if result.Error != "" {
				fmt.Printf("   Error: %s\n", result.Error)
			}
		}
		
		return nil
	}

	// Test specific registry
	registryURL := args[0]
	managerName := "npm" // default
	if len(args) > 1 {
		managerName = args[1]
	}
	
	logger.Debug("Testing specific registry", "manager", managerName, "registry", registryURL)
	
	err := configService.TestRegistry(ctx, managerName, registryURL)
	
	result := RegistryTestResult{
		Manager:  managerName,
		Registry: registryURL,
		Success:  err == nil,
	}
	if err != nil {
		result.Error = err.Error()
	}

	if jsonOutput {
		return outputJSON(result)
	}

	if result.Success {
		fmt.Printf("✅ Registry test passed: %s\n", registryURL)
	} else {
		fmt.Printf("❌ Registry test failed: %s\n", registryURL)
		fmt.Printf("Error: %s\n", result.Error)
	}
	
	return nil
}

func runRegistryReset(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	configService := services.NewConfigService()
	
	resetAll, _ := cmd.Flags().GetBool("all")
	force, _ := cmd.Flags().GetBool("force")
	defaultRegistry := "https://registry.npmjs.org/"
	
	logger := logger.GetDefault()

	if len(args) > 0 && !resetAll {
		// Reset specific manager
		managerName := args[0]
		
		if !force {
			fmt.Printf("This will reset %s registry to default (%s). Continue? (y/N): ", managerName, defaultRegistry)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Registry reset cancelled.")
				return nil
			}
		}
		
		logger.Debug("Resetting registry for specific manager", "manager", managerName)
		
		err := configService.SetRegistry(ctx, managerName, defaultRegistry)
		if err != nil {
			return fmt.Errorf("failed to reset registry for %s: %w", managerName, err)
		}
		
		fmt.Printf("✅ Registry reset for %s: %s\n", managerName, defaultRegistry)
		return nil
	}

	// Reset all managers
	if !force {
		fmt.Printf("This will reset all registries to default (%s). Continue? (y/N): ", defaultRegistry)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Registry reset cancelled.")
			return nil
		}
	}
	
	logger.Debug("Resetting registry for all managers")
	
	err := configService.SetRegistryForAll(ctx, defaultRegistry)
	if err != nil {
		return fmt.Errorf("failed to reset registry for all managers: %w", err)
	}
	
	fmt.Printf("✅ Registry reset for all managers: %s\n", defaultRegistry)
	return nil
}

// RegistryTestResult represents the result of a registry test
type RegistryTestResult struct {
	Manager  string `json:"manager"`
	Registry string `json:"registry"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}
