package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"npm-console/internal/web"
	"npm-console/pkg/config"
	"npm-console/pkg/logger"

	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web server",
	Long: `Start the web server to provide a web-based interface for managing
package managers, caches, packages, and configurations.
	
The web interface provides:
- Dashboard with overview of all package managers
- Cache management and cleanup
- Package browsing and search
- Registry and proxy configuration
- Project analysis and management`,
	RunE: runWebServer,
}

func init() {
	rootCmd.AddCommand(webCmd)

	// Add flags
	webCmd.Flags().StringP("host", "H", "", "Host to bind to (default from config)")
	webCmd.Flags().IntP("port", "p", 0, "Port to listen on (default from config)")
	webCmd.Flags().BoolP("open", "o", false, "Open browser automatically")
	webCmd.Flags().BoolP("dev", "d", false, "Enable development mode")
}

func runWebServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with command line flags
	if host, _ := cmd.Flags().GetString("host"); host != "" {
		cfg.Web.Host = host
	}
	if port, _ := cmd.Flags().GetInt("port"); port != 0 {
		cfg.Web.Port = port
	}

	// Set up logger
	loggerInstance, err := logger.New(&cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	logger.SetDefault(loggerInstance)

	log := logger.GetDefault()
	log.Info("Starting npm-console web server", "version", rootCmd.Version)

	// Check if web server is enabled
	if !cfg.Web.Enabled {
		return fmt.Errorf("web server is disabled in configuration")
	}

	// Create and start web server
	server := web.NewServer(cfg)

	serverAddr := fmt.Sprintf("%s:%d", cfg.Web.Host, cfg.Web.Port)

	fmt.Printf("🚀 npm-console web server starting...\n")
	fmt.Printf("📍 Address: http://%s\n", serverAddr)
	fmt.Printf("🌐 Web interface is now available!\n")
	fmt.Printf("\nPress Ctrl+C to stop the server\n\n")

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.WithError(err).Error("Failed to start web server")
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Info("Shutdown signal received, stopping server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Failed to shutdown server gracefully")
		return err
	}

	log.Info("Web server stopped gracefully")
	return nil
}
