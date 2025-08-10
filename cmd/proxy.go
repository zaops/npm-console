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

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage proxy settings",
	Long: `Manage proxy settings for npm, pnpm, yarn, and bun package managers.
	
This command provides functionality to:
- List current proxy configurations
- Set proxy URLs for specific or all package managers
- Remove proxy settings
- Test proxy connectivity`,
}

var proxyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current proxy configurations",
	Long:  `Display current proxy configurations for all available package managers.`,
	RunE:  runProxyList,
}

var proxySetCmd = &cobra.Command{
	Use:   "set <proxy-url> [manager]",
	Short: "Set proxy URL",
	Long: `Set proxy URL for a specific package manager or all managers.
	
Examples:
  npm-console proxy set http://proxy.company.com:8080           # Set for all managers
  npm-console proxy set http://proxy.company.com:8080 npm      # Set for npm only
  npm-console proxy set http://user:pass@proxy.com:8080        # Set with authentication`,
	Args: cobra.MinimumNArgs(1),
	RunE: runProxySet,
}

var proxyUnsetCmd = &cobra.Command{
	Use:   "unset [manager]",
	Short: "Remove proxy settings",
	Long: `Remove proxy settings for a specific package manager or all managers.
	
Examples:
  npm-console proxy unset          # Remove proxy for all managers
  npm-console proxy unset npm      # Remove proxy for npm only`,
	RunE: runProxyUnset,
}

var proxyTestCmd = &cobra.Command{
	Use:   "test [proxy-url] [manager]",
	Short: "Test proxy connectivity",
	Long: `Test connectivity through a proxy.
	
Examples:
  npm-console proxy test                                        # Test current proxies
  npm-console proxy test http://proxy.company.com:8080         # Test specific proxy
  npm-console proxy test http://proxy.company.com:8080 npm     # Test proxy for npm`,
	RunE: runProxyTest,
}

func init() {
	rootCmd.AddCommand(proxyCmd)
	proxyCmd.AddCommand(proxyListCmd)
	proxyCmd.AddCommand(proxySetCmd)
	proxyCmd.AddCommand(proxyUnsetCmd)
	proxyCmd.AddCommand(proxyTestCmd)

	// Add flags
	proxyListCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	proxySetCmd.Flags().BoolP("all", "a", false, "Set for all available managers")
	proxyUnsetCmd.Flags().BoolP("all", "a", false, "Unset for all managers")
	proxyUnsetCmd.Flags().BoolP("force", "f", false, "Force unset without confirmation")
	proxyTestCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
}

func runProxyList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	configService := services.NewConfigService()
	
	logger := logger.GetDefault()
	logger.Debug("Listing proxy configurations")

	configs, err := configService.GetAllConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get proxy configurations: %w", err)
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
	fmt.Fprintln(w, "MANAGER\tPROXY\tSTATUS")
	fmt.Fprintln(w, "-------\t-----\t------")

	for _, config := range configs {
		proxy := config.Proxy
		status := "Not set"
		
		if proxy != "" {
			status = "Configured"
		} else {
			proxy = "(none)"
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			config.Manager,
			proxy,
			status,
		)
	}

	w.Flush()
	return nil
}

func runProxySet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	configService := services.NewConfigService()
	
	proxyURL := args[0]
	setAll, _ := cmd.Flags().GetBool("all")
	
	logger := logger.GetDefault()

	if len(args) > 1 && !setAll {
		// Set for specific manager
		managerName := args[1]
		logger.Debug("Setting proxy for specific manager", "manager", managerName, "proxy", proxyURL)
		
		err := configService.SetProxy(ctx, managerName, proxyURL)
		if err != nil {
			return fmt.Errorf("failed to set proxy for %s: %w", managerName, err)
		}
		
		fmt.Printf("✅ Proxy set for %s: %s\n", managerName, proxyURL)
		return nil
	}

	// Set for all managers
	logger.Debug("Setting proxy for all managers", "proxy", proxyURL)
	
	err := configService.SetProxyForAll(ctx, proxyURL)
	if err != nil {
		return fmt.Errorf("failed to set proxy for all managers: %w", err)
	}
	
	fmt.Printf("✅ Proxy set for all managers: %s\n", proxyURL)
	return nil
}

func runProxyUnset(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	configService := services.NewConfigService()
	
	unsetAll, _ := cmd.Flags().GetBool("all")
	force, _ := cmd.Flags().GetBool("force")
	
	logger := logger.GetDefault()

	if len(args) > 0 && !unsetAll {
		// Unset for specific manager
		managerName := args[0]
		
		if !force {
			fmt.Printf("This will remove proxy settings for %s. Continue? (y/N): ", managerName)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Proxy unset cancelled.")
				return nil
			}
		}
		
		logger.Debug("Unsetting proxy for specific manager", "manager", managerName)
		
		err := configService.SetProxy(ctx, managerName, "")
		if err != nil {
			return fmt.Errorf("failed to unset proxy for %s: %w", managerName, err)
		}
		
		fmt.Printf("✅ Proxy removed for %s\n", managerName)
		return nil
	}

	// Unset for all managers
	if !force {
		fmt.Print("This will remove proxy settings for all managers. Continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Proxy unset cancelled.")
			return nil
		}
	}
	
	logger.Debug("Unsetting proxy for all managers")
	
	err := configService.SetProxyForAll(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to unset proxy for all managers: %w", err)
	}
	
	fmt.Printf("✅ Proxy removed for all managers\n")
	return nil
}

func runProxyTest(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	configService := services.NewConfigService()
	
	jsonOutput, _ := cmd.Flags().GetBool("json")
	logger := logger.GetDefault()

	if len(args) == 0 {
		// Test current proxies
		logger.Debug("Testing current proxies")
		
		configs, err := configService.GetAllConfigs(ctx)
		if err != nil {
			return fmt.Errorf("failed to get proxy configurations: %w", err)
		}

		var results []ProxyTestResult
		
		for _, config := range configs {
			if config.Proxy == "" {
				continue
			}
			
			err := configService.TestProxy(ctx, config.Manager, config.Proxy)
			result := ProxyTestResult{
				Manager: config.Manager,
				Proxy:   config.Proxy,
				Success: err == nil,
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
			fmt.Println("No proxies configured to test.")
			return nil
		}

		fmt.Println("Proxy Test Results:")
		fmt.Println("==================")
		
		for _, result := range results {
			status := "✅ PASS"
			if !result.Success {
				status = "❌ FAIL"
			}
			
			fmt.Printf("%s %s: %s\n", status, result.Manager, result.Proxy)
			if result.Error != "" {
				fmt.Printf("   Error: %s\n", result.Error)
			}
		}
		
		return nil
	}

	// Test specific proxy
	proxyURL := args[0]
	managerName := "npm" // default
	if len(args) > 1 {
		managerName = args[1]
	}
	
	logger.Debug("Testing specific proxy", "manager", managerName, "proxy", proxyURL)
	
	err := configService.TestProxy(ctx, managerName, proxyURL)
	
	result := ProxyTestResult{
		Manager: managerName,
		Proxy:   proxyURL,
		Success: err == nil,
	}
	if err != nil {
		result.Error = err.Error()
	}

	if jsonOutput {
		return outputJSON(result)
	}

	if result.Success {
		fmt.Printf("✅ Proxy test passed: %s\n", proxyURL)
	} else {
		fmt.Printf("❌ Proxy test failed: %s\n", proxyURL)
		fmt.Printf("Error: %s\n", result.Error)
	}
	
	return nil
}

// ProxyTestResult represents the result of a proxy test
type ProxyTestResult struct {
	Manager string `json:"manager"`
	Proxy   string `json:"proxy"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
