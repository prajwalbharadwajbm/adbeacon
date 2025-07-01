package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
)

// Cache defines the interface for campaign caching
type Cache interface {
	// Campaign operations
	GetActiveCampaigns(ctx context.Context) ([]models.CampaignWithRules, error)
	SetActiveCampaigns(ctx context.Context, campaigns []models.CampaignWithRules, ttl time.Duration) error

	// Campaign index operations (for fast lookups)
	GetCampaignIndex(ctx context.Context, dimension models.TargetDimension, value string) ([]string, error)
	SetCampaignIndex(ctx context.Context, dimension models.TargetDimension, value string, campaignIDs []string, ttl time.Duration) error

	// Cache management
	InvalidateAll(ctx context.Context) error
	GetStats() CacheStats
}

// CacheStats holds cache performance statistics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Errors      int64
	HitRatio    float64
	TotalOps    int64
	LastUpdated time.Time
}

// HybridCache implements both in-memory and Redis caching
// Minimize database hits while maintaining consistency
type HybridCache struct {
	// In-memory cache for ultra-fast access
	memoryCache *memoryCache
	// Redis cache for shared state
	redisCache *redisCache
	// Configuration
	config CacheConfig
	// Metrics
	stats CacheStats
	mu    sync.RWMutex
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	DefaultTTL      time.Duration
	MemoryCacheSize int
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	EnableMemory    bool
	EnableRedis     bool
	RefreshInterval time.Duration
}

// NewHybridCache creates a new hybrid cache
func NewHybridCache(config CacheConfig) (*HybridCache, error) {
	hc := &HybridCache{
		config: config,
		stats: CacheStats{
			LastUpdated: time.Now(),
		},
	}

	// Initialize in-memory cache if enabled
	if config.EnableMemory {
		hc.memoryCache = newMemoryCache(config.MemoryCacheSize)
	}

	// Initialize Redis cache if enabled
	if config.EnableRedis {
		var err error
		hc.redisCache, err = newRedisCache(config)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Redis cache: %w", err)
		}
	}

	return hc, nil
}

// GetActiveCampaigns retrieves campaigns from cache (memory first, then Redis, then miss)
func (hc *HybridCache) GetActiveCampaigns(ctx context.Context) ([]models.CampaignWithRules, error) {
	// Try memory cache first
	if hc.memoryCache != nil {
		if campaigns, found := hc.memoryCache.getActiveCampaigns(); found {
			hc.recordHit()
			return campaigns, nil
		}
	}

	// Try Redis cache
	if hc.redisCache != nil {
		campaigns, err := hc.redisCache.getActiveCampaigns(ctx)
		if err == nil {
			hc.recordHit()
			// Warm memory cache
			if hc.memoryCache != nil {
				hc.memoryCache.setActiveCampaigns(campaigns, hc.config.DefaultTTL)
			}
			return campaigns, nil
		}
	}

	hc.recordMiss()
	return nil, ErrCacheMiss
}

// SetActiveCampaigns stores campaigns in both caches
func (hc *HybridCache) SetActiveCampaigns(ctx context.Context, campaigns []models.CampaignWithRules, ttl time.Duration) error {
	var errs []error

	// Store in memory cache
	if hc.memoryCache != nil {
		hc.memoryCache.setActiveCampaigns(campaigns, ttl)
	}

	// Store in Redis cache
	if hc.redisCache != nil {
		if err := hc.redisCache.setActiveCampaigns(ctx, campaigns, ttl); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		hc.recordError()
		return fmt.Errorf("cache store errors: %v", errs)
	}

	return nil
}

// GetCampaignIndex gets campaign IDs for a specific targeting dimension/value
func (hc *HybridCache) GetCampaignIndex(ctx context.Context, dimension models.TargetDimension, value string) ([]string, error) {
	key := fmt.Sprintf("index:%s:%s", dimension, value)

	// Try memory cache first
	if hc.memoryCache != nil {
		if campaignIDs, found := hc.memoryCache.getCampaignIndex(key); found {
			hc.recordHit()
			return campaignIDs, nil
		}
	}

	// Try Redis cache
	if hc.redisCache != nil {
		campaignIDs, err := hc.redisCache.getCampaignIndex(ctx, key)
		if err == nil {
			hc.recordHit()
			// Warm memory cache
			if hc.memoryCache != nil {
				hc.memoryCache.setCampaignIndex(key, campaignIDs, hc.config.DefaultTTL)
			}
			return campaignIDs, nil
		}
	}

	hc.recordMiss()
	return nil, ErrCacheMiss
}

// SetCampaignIndex stores campaign index in both caches
func (hc *HybridCache) SetCampaignIndex(ctx context.Context, dimension models.TargetDimension, value string, campaignIDs []string, ttl time.Duration) error {
	key := fmt.Sprintf("index:%s:%s", dimension, value)
	var errs []error

	// Store in memory cache
	if hc.memoryCache != nil {
		hc.memoryCache.setCampaignIndex(key, campaignIDs, ttl)
	}

	// Store in Redis cache
	if hc.redisCache != nil {
		if err := hc.redisCache.setCampaignIndex(ctx, key, campaignIDs, ttl); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		hc.recordError()
		return fmt.Errorf("cache index store errors: %v", errs)
	}

	return nil
}

// InvalidateAll clears all caches
func (hc *HybridCache) InvalidateAll(ctx context.Context) error {
	var errs []error

	// Clear memory cache
	if hc.memoryCache != nil {
		hc.memoryCache.clear()
	}

	// Clear Redis cache
	if hc.redisCache != nil {
		if err := hc.redisCache.clear(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cache invalidation errors: %v", errs)
	}

	return nil
}

// GetStats returns cache statistics
func (hc *HybridCache) GetStats() CacheStats {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	stats := hc.stats
	if stats.TotalOps > 0 {
		stats.HitRatio = float64(stats.Hits) / float64(stats.TotalOps)
	}
	return stats
}

// Helper methods for statistics
func (hc *HybridCache) recordHit() {
	hc.mu.Lock()
	hc.stats.Hits++
	hc.stats.TotalOps++
	hc.mu.Unlock()
}

func (hc *HybridCache) recordMiss() {
	hc.mu.Lock()
	hc.stats.Misses++
	hc.stats.TotalOps++
	hc.mu.Unlock()
}

func (hc *HybridCache) recordError() {
	hc.mu.Lock()
	hc.stats.Errors++
	hc.mu.Unlock()
}

// Custom errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)
