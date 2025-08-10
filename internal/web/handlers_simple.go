package web

import (
	"context"
	"fmt"
	"path/filepath"

	"npm-console/internal/core"
	"npm-console/internal/managers"
	"npm-console/internal/services"

	"github.com/gofiber/fiber/v2"
)

// Cache handlers

func (s *Server) handleGetAllCacheInfo(c *fiber.Ctx) error {
	ctx := context.Background()
	
	cacheInfos, err := s.cacheService.GetAllCacheInfo(ctx)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, cacheInfos)
}

func (s *Server) handleGetCacheSummary(c *fiber.Ctx) error {
	ctx := context.Background()
	
	summary, err := s.cacheService.GetCacheSummary(ctx)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, summary)
}

func (s *Server) handleGetTotalCacheSize(c *fiber.Ctx) error {
	ctx := context.Background()
	
	totalSize, err := s.cacheService.GetTotalCacheSize(ctx)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, fiber.Map{
		"total_size": totalSize,
	})
}

func (s *Server) handleGetCacheInfo(c *fiber.Ctx) error {
	ctx := context.Background()
	manager := c.Params("manager")
	
	cacheInfo, err := s.cacheService.GetCacheInfo(ctx, manager)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, cacheInfo)
}

func (s *Server) handleClearAllCaches(c *fiber.Ctx) error {
	ctx := context.Background()
	
	err := s.cacheService.ClearAllCaches(ctx)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, fiber.Map{
		"message": "All caches cleared successfully",
	})
}

func (s *Server) handleClearCache(c *fiber.Ctx) error {
	ctx := context.Background()
	manager := c.Params("manager")
	
	err := s.cacheService.ClearCache(ctx, manager)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, fiber.Map{
		"message": "Cache cleared successfully for " + manager,
	})
}

// Package handlers

func (s *Server) handleGetPackages(c *fiber.Ctx) error {
	ctx := context.Background()
	projectPath := c.Query("path", ".")
	manager := c.Query("manager", "")
	
	// Convert to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return s.sendError(c, fiber.StatusBadRequest, "Invalid project path")
	}
	
	var packages []core.Package
	if manager != "" {
		packages, err = s.packageService.GetPackagesByManager(ctx, manager, absPath)
	} else {
		packages, err = s.packageService.GetAllPackages(ctx, absPath)
	}
	
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, packages)
}

func (s *Server) handleGetGlobalPackages(c *fiber.Ctx) error {
	ctx := context.Background()
	manager := c.Query("manager", "")
	
	var packages []core.Package
	var err error
	
	if manager != "" {
		packages, err = s.packageService.GetGlobalPackagesByManager(ctx, manager)
	} else {
		packages, err = s.packageService.GetGlobalPackages(ctx)
	}
	
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, packages)
}

func (s *Server) handleSearchPackages(c *fiber.Ctx) error {
	ctx := context.Background()
	query := c.Query("q", "")
	
	if query == "" {
		return s.sendError(c, fiber.StatusBadRequest, "Search query is required")
	}
	
	packages, err := s.packageService.SearchPackages(ctx, query)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, packages)
}

func (s *Server) handleGetPackageStats(c *fiber.Ctx) error {
	ctx := context.Background()
	projectPath := c.Query("path", ".")
	global := c.Query("global", "false") == "true"
	
	var stats *services.PackageStats
	var err error
	
	if global {
		stats, err = s.packageService.GetGlobalPackageStats(ctx)
	} else {
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			return s.sendError(c, fiber.StatusBadRequest, "Invalid project path")
		}
		stats, err = s.packageService.GetPackageStats(ctx, absPath)
	}
	
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, stats)
}

func (s *Server) handleGetPackageInfo(c *fiber.Ctx) error {
	ctx := context.Background()
	packageName := c.Params("name")
	
	packageInfo, err := s.packageService.GetPackageInfo(ctx, packageName)
	if err != nil {
		return s.sendError(c, fiber.StatusNotFound, err.Error())
	}
	
	return s.sendSuccess(c, packageInfo)
}

// Config handlers

func (s *Server) handleGetAllConfigs(c *fiber.Ctx) error {
	ctx := context.Background()
	
	configs, err := s.configService.GetAllConfigs(ctx)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, configs)
}

func (s *Server) handleGetConfigSummary(c *fiber.Ctx) error {
	ctx := context.Background()
	
	summary, err := s.configService.GetConfigSummary(ctx)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, summary)
}

func (s *Server) handleGetConfig(c *fiber.Ctx) error {
	ctx := context.Background()
	manager := c.Params("manager")
	
	config, err := s.configService.GetConfig(ctx, manager)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, config)
}

func (s *Server) handleSetRegistry(c *fiber.Ctx) error {
	ctx := context.Background()
	manager := c.Params("manager")
	
	var req struct {
		Registry string `json:"registry"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return s.sendError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	
	if req.Registry == "" {
		return s.sendError(c, fiber.StatusBadRequest, "Registry URL is required")
	}
	
	err := s.configService.SetRegistry(ctx, manager, req.Registry)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, fiber.Map{
		"message": "Registry updated successfully for " + manager,
	})
}

func (s *Server) handleSetProxy(c *fiber.Ctx) error {
	ctx := context.Background()
	manager := c.Params("manager")
	
	var req struct {
		Proxy string `json:"proxy"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return s.sendError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	
	err := s.configService.SetProxy(ctx, manager, req.Proxy)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, fiber.Map{
		"message": "Proxy updated successfully for " + manager,
	})
}

func (s *Server) handleUnsetProxy(c *fiber.Ctx) error {
	ctx := context.Background()
	manager := c.Params("manager")
	
	err := s.configService.SetProxy(ctx, manager, "")
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	
	return s.sendSuccess(c, fiber.Map{
		"message": "Proxy removed successfully for " + manager,
	})
}



// Manager handlers

func (s *Server) handleGetManagers(c *fiber.Ctx) error {
	factory := managers.GetGlobalFactory()
	managerNames := factory.GetManagerNames()
	
	return s.sendSuccess(c, managerNames)
}

func (s *Server) handleGetAvailableManagers(c *fiber.Ctx) error {
	ctx := context.Background()
	factory := managers.GetGlobalFactory()

	availableManagers := factory.GetAvailableManagerNames(ctx)

	return s.sendSuccess(c, availableManagers)
}

// Package installation and uninstallation handlers

func (s *Server) handleInstallPackage(c *fiber.Ctx) error {
	ctx := context.Background()

	var req struct {
		Name    string `json:"name"`
		Manager string `json:"manager"`
		Global  bool   `json:"global"`
	}

	if err := c.BodyParser(&req); err != nil {
		return s.sendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" {
		return s.sendError(c, fiber.StatusBadRequest, "Package name is required")
	}

	if req.Manager == "" {
		req.Manager = "npm" // Default to npm
	}

	err := s.packageService.InstallPackage(ctx, req.Name, req.Manager, req.Global)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return s.sendSuccess(c, fiber.Map{
		"message": fmt.Sprintf("Package %s installed successfully with %s", req.Name, req.Manager),
	})
}

func (s *Server) handleUninstallPackage(c *fiber.Ctx) error {
	ctx := context.Background()

	var req struct {
		Name    string `json:"name"`
		Manager string `json:"manager"`
		Global  bool   `json:"global"`
	}

	if err := c.BodyParser(&req); err != nil {
		return s.sendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" {
		return s.sendError(c, fiber.StatusBadRequest, "Package name is required")
	}

	if req.Manager == "" {
		req.Manager = "npm" // Default to npm
	}

	err := s.packageService.UninstallPackage(ctx, req.Name, req.Manager, req.Global)
	if err != nil {
		return s.sendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return s.sendSuccess(c, fiber.Map{
		"message": fmt.Sprintf("Package %s uninstalled successfully with %s", req.Name, req.Manager),
	})
}
