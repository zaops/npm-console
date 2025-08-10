package services

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"npm-console/internal/core"
	"npm-console/internal/managers"
	"npm-console/pkg/logger"
)

// CacheService implements cache management functionality
type CacheService struct {
	factory *managers.ManagerFactory
	logger  *logger.Logger
}

// NewCacheService creates a new cache service
func NewCacheService() *CacheService {
	return &CacheService{
		factory: managers.GetGlobalFactory(),
		logger:  logger.GetDefault().WithField("service", "cache"),
	}
}

// GetAllCacheInfo returns cache information for all available package managers
func (s *CacheService) GetAllCacheInfo(ctx context.Context) ([]core.CacheInfo, error) {
	availableManagers := s.factory.GetAvailableManagers(ctx)
	
	var cacheInfos []core.CacheInfo
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errors []error

	// Get cache info concurrently for better performance
	for name, manager := range availableManagers {
		wg.Add(1)
		go func(name string, mgr core.PackageManager) {
			defer wg.Done()
			
			cacheInfo, err := mgr.GetCacheInfo(ctx)
			if err != nil {
				s.logger.WithError(err).WithField("manager", name).Warn("Failed to get cache info")
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to get cache info for %s: %w", name, err))
				mu.Unlock()
				return
			}
			
			mu.Lock()
			cacheInfos = append(cacheInfos, *cacheInfo)
			mu.Unlock()
		}(name, manager)
	}
	
	wg.Wait()
	
	// Sort by manager name for consistent output
	sort.Slice(cacheInfos, func(i, j int) bool {
		return cacheInfos[i].Manager < cacheInfos[j].Manager
	})
	
	// Log any errors but don't fail the entire operation
	if len(errors) > 0 {
		for _, err := range errors {
			s.logger.WithError(err).Warn("Cache info retrieval error")
		}
	}
	
	return cacheInfos, nil
}

// GetCacheInfo returns cache information for a specific package manager
func (s *CacheService) GetCacheInfo(ctx context.Context, managerName string) (*core.CacheInfo, error) {
	manager, err := s.factory.GetManager(managerName)
	if err != nil {
		return nil, err
	}
	
	if !manager.IsAvailable(ctx) {
		return nil, core.NewManagerError(managerName, "get cache info", core.ErrManagerNotAvailable)
	}
	
	return manager.GetCacheInfo(ctx)
}

// ClearAllCaches clears caches for all available package managers
func (s *CacheService) ClearAllCaches(ctx context.Context) error {
	availableManagers := s.factory.GetAvailableManagers(ctx)
	
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errors []error
	var clearedCount int

	// Clear caches concurrently
	for name, manager := range availableManagers {
		wg.Add(1)
		go func(name string, mgr core.PackageManager) {
			defer wg.Done()
			
			err := mgr.ClearCache(ctx)
			if err != nil {
				s.logger.WithError(err).WithField("manager", name).Error("Failed to clear cache")
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to clear cache for %s: %w", name, err))
				mu.Unlock()
				return
			}
			
			mu.Lock()
			clearedCount++
			mu.Unlock()
			
			s.logger.WithField("manager", name).Info("Cache cleared successfully")
		}(name, manager)
	}
	
	wg.Wait()
	
	s.logger.WithField("cleared_count", clearedCount).WithField("total_managers", len(availableManagers)).Info("Cache clearing completed")
	
	// Return error if any cache clearing failed
	if len(errors) > 0 {
		return fmt.Errorf("failed to clear some caches: %v", errors)
	}
	
	return nil
}

// ClearCache clears cache for a specific package manager
func (s *CacheService) ClearCache(ctx context.Context, managerName string) error {
	manager, err := s.factory.GetManager(managerName)
	if err != nil {
		return err
	}
	
	if !manager.IsAvailable(ctx) {
		return core.NewManagerError(managerName, "clear cache", core.ErrManagerNotAvailable)
	}
	
	return manager.ClearCache(ctx)
}

// GetTotalCacheSize calculates the total cache size across all package managers
func (s *CacheService) GetTotalCacheSize(ctx context.Context) (int64, error) {
	cacheInfos, err := s.GetAllCacheInfo(ctx)
	if err != nil {
		return 0, err
	}
	
	var totalSize int64
	for _, info := range cacheInfos {
		totalSize += info.Size
	}
	
	return totalSize, nil
}

// GetCacheStats returns detailed cache statistics
func (s *CacheService) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	cacheInfos, err := s.GetAllCacheInfo(ctx)
	if err != nil {
		return nil, err
	}
	
	stats := &CacheStats{
		Managers: make(map[string]CacheManagerStats),
	}
	
	for _, info := range cacheInfos {
		stats.TotalSize += info.Size
		stats.TotalFiles += info.FileCount
		stats.ManagerCount++
		
		stats.Managers[info.Manager] = CacheManagerStats{
			Size:        info.Size,
			FileCount:   info.FileCount,
			Path:        info.Path,
			LastUpdated: info.LastUpdated,
		}
		
		if info.Size > stats.LargestCache.Size {
			stats.LargestCache = CacheManagerStats{
				Size:        info.Size,
				FileCount:   info.FileCount,
				Path:        info.Path,
				LastUpdated: info.LastUpdated,
			}
			stats.LargestCacheManager = info.Manager
		}
	}
	
	return stats, nil
}

// GetCacheSummary returns a summary of cache information
func (s *CacheService) GetCacheSummary(ctx context.Context) (*CacheSummary, error) {
	stats, err := s.GetCacheStats(ctx)
	if err != nil {
		return nil, err
	}
	
	summary := &CacheSummary{
		TotalSize:            stats.TotalSize,
		TotalFiles:           stats.TotalFiles,
		ManagerCount:         stats.ManagerCount,
		LargestCacheManager:  stats.LargestCacheManager,
		LargestCacheSize:     stats.LargestCache.Size,
		AvailableManagers:    s.factory.GetAvailableManagerNames(ctx),
	}
	
	return summary, nil
}

// ValidateManagerName validates if a manager name is valid and available
func (s *CacheService) ValidateManagerName(ctx context.Context, managerName string) error {
	if err := s.factory.ValidateManager(managerName); err != nil {
		return err
	}
	
	if !s.factory.IsManagerAvailable(ctx, managerName) {
		return core.NewManagerError(managerName, "validate", core.ErrManagerNotAvailable)
	}
	
	return nil
}

// CacheStats represents detailed cache statistics
type CacheStats struct {
	TotalSize            int64                        `json:"total_size"`
	TotalFiles           int                          `json:"total_files"`
	ManagerCount         int                          `json:"manager_count"`
	LargestCache         CacheManagerStats            `json:"largest_cache"`
	LargestCacheManager  string                       `json:"largest_cache_manager"`
	Managers             map[string]CacheManagerStats `json:"managers"`
}

// CacheManagerStats represents cache statistics for a specific manager
type CacheManagerStats struct {
	Size        int64     `json:"size"`
	FileCount   int       `json:"file_count"`
	Path        string    `json:"path"`
	LastUpdated time.Time `json:"last_updated"`
}

// CacheSummary represents a summary of cache information
type CacheSummary struct {
	TotalSize           int64    `json:"total_size"`
	TotalFiles          int      `json:"total_files"`
	ManagerCount        int      `json:"manager_count"`
	LargestCacheManager string   `json:"largest_cache_manager"`
	LargestCacheSize    int64    `json:"largest_cache_size"`
	AvailableManagers   []string `json:"available_managers"`
}
