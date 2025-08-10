package web

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"npm-console/internal/services"
	"npm-console/pkg/config"
	"npm-console/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Server represents the web server
type Server struct {
	app           *fiber.App
	config        *config.Config
	logger        *logger.Logger
	cacheService  *services.CacheService
	packageService *services.PackageService
	configService *services.ConfigService
	projectService *services.ProjectService
}

// NewServer creates a new web server instance
func NewServer(cfg *config.Config) *Server {
	// Create Fiber app with custom config
	app := fiber.New(fiber.Config{
		AppName:      "npm-console",
		ServerHeader: "npm-console/1.0.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return errorHandler(c, err)
		},
	})

	server := &Server{
		app:            app,
		config:         cfg,
		logger:         logger.GetDefault().WithField("component", "web-server"),
		cacheService:   services.NewCacheService(),
		packageService: services.NewPackageService(),
		configService:  services.NewConfigService(),
		projectService: services.NewProjectService(),
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.app.Use(recover.New())

	// CORS middleware
	if s.config.Web.CORS.Enabled {
		corsConfig := cors.Config{
			AllowOrigins: strings.Join(s.config.Web.CORS.AllowedOrigins, ","),
			AllowMethods: strings.Join(s.config.Web.CORS.AllowedMethods, ","),
			AllowHeaders: strings.Join(s.config.Web.CORS.AllowedHeaders, ","),
		}

		// 只有在不是通配符时才允许凭据
		if !contains(s.config.Web.CORS.AllowedOrigins, "*") {
			corsConfig.AllowCredentials = true
		}

		s.app.Use(cors.New(corsConfig))
	}

	// Static files middleware
	staticPath := getStaticPath()
	s.app.Static("/", staticPath)

	// Logging middleware
	s.app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		
		s.logger.Info("HTTP request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", duration,
			"ip", c.IP(),
		)
		
		return err
	})
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// API routes
	api := s.app.Group("/api/v1")

	// Health check
	api.Get("/health", s.handleHealth)

	// Cache routes
	cache := api.Group("/cache")
	cache.Get("/", s.handleGetAllCacheInfo)
	cache.Get("/summary", s.handleGetCacheSummary)
	cache.Get("/size", s.handleGetTotalCacheSize)
	cache.Get("/:manager", s.handleGetCacheInfo)
	cache.Delete("/", s.handleClearAllCaches)
	cache.Delete("/:manager", s.handleClearCache)

	// Package routes
	packages := api.Group("/packages")
	packages.Get("/", s.handleGetPackages)
	packages.Get("/global", s.handleGetGlobalPackages)
	packages.Get("/search", s.handleSearchPackages)
	packages.Get("/stats", s.handleGetPackageStats)
	packages.Get("/:name", s.handleGetPackageInfo)
	packages.Post("/install", s.handleInstallPackage)
	packages.Post("/uninstall", s.handleUninstallPackage)

	// Config routes
	configs := api.Group("/config")
	configs.Get("/", s.handleGetAllConfigs)
	configs.Get("/summary", s.handleGetConfigSummary)
	configs.Get("/:manager", s.handleGetConfig)
	configs.Put("/:manager/registry", s.handleSetRegistry)
	configs.Put("/:manager/proxy", s.handleSetProxy)
	configs.Delete("/:manager/proxy", s.handleUnsetProxy)



	// Manager routes
	managers := api.Group("/managers")
	managers.Get("/", s.handleGetManagers)
	managers.Get("/available", s.handleGetAvailableManagers)

	// Catch-all route for SPA
	s.app.Get("/*", func(c *fiber.Ctx) error {
		staticPath := getStaticPath()
		indexPath := filepath.Join(staticPath, "index.html")
		return c.SendFile(indexPath)
	})
}

// Start starts the web server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Web.Host, s.config.Web.Port)
	
	s.logger.Info("Starting web server", "address", addr)
	
	if s.config.Web.TLS.Enabled {
		return s.app.ListenTLS(addr, s.config.Web.TLS.CertFile, s.config.Web.TLS.KeyFile)
	}
	
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the web server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down web server")
	return s.app.ShutdownWithContext(ctx)
}

// errorHandler handles Fiber errors
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Log error
	logger.GetDefault().WithError(err).Error("HTTP error",
		"method", c.Method(),
		"path", c.Path(),
		"status", code,
		"ip", c.IP(),
	)

	return c.Status(code).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
	})
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// sendSuccess sends a successful response
func (s *Server) sendSuccess(c *fiber.Ctx, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

// sendError sends an error response
func (s *Server) sendError(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(c *fiber.Ctx) error {
	return s.sendSuccess(c, fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

// contains 检查字符串切片中是否包含指定的字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getStaticPath 获取静态文件的正确路径
func getStaticPath() string {
	// 尝试多个可能的路径
	possiblePaths := []string{
		"./web/dist",           // 从项目根目录运行
		"../web/dist",          // 从bin目录运行
		"../../web/dist",       // 从更深的目录运行
		"web/dist",             // 相对路径
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(filepath.Join(path, "index.html")); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// 如果都找不到，返回默认路径
	return "./web/dist"
}
