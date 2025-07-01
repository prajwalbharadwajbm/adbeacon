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

	// Health check operations
	HealthCheck(ctx context.Context) CacheHealth
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

// CacheHealth represents comprehensive cache health information
type CacheHealth struct {
	Overall  string            `json:"overall"` // "healthy", "degraded", "unhealthy"
	Memory   MemoryCacheHealth `json:"memory"`
	Redis    RedisCacheHealth  `json:"redis"`
	Stats    CacheStats        `json:"stats"`
	Uptime   time.Duration     `json:"uptime"`
	LastTest time.Time         `json:"last_test"`
}

// MemoryCacheHealth represents in-memory cache health
type MemoryCacheHealth struct {
	Enabled     bool    `json:"enabled"`
	Status      string  `json:"status"`       // "healthy", "unhealthy"
	Size        int     `json:"size"`         // Current number of items
	MaxSize     int     `json:"max_size"`     // Maximum capacity
	UtilPct     float64 `json:"util_pct"`     // Utilization percentage
	EvictedKeys int64   `json:"evicted_keys"` // Number of evicted keys
}

// RedisCacheHealth represents Redis cache health
type RedisCacheHealth struct {
	Enabled   bool          `json:"enabled"`
	Status    string        `json:"status"` // "healthy", "unhealthy", "disconnected"
	Connected bool          `json:"connected"`
	Address   string        `json:"address"`
	Latency   time.Duration `json:"latency"` // Ping latency
	Error     string        `json:"error,omitempty"`
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
	// Add startTime tracking to HybridCache
	startTime time.Time
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
		startTime: time.Now(),
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

// HealthCheck performs comprehensive cache health check
func (hc *HybridCache) HealthCheck(ctx context.Context) CacheHealth {
	hc.mu.RLock()
	startTime := hc.startTime
	hc.mu.RUnlock()

	health := CacheHealth{
		Stats:    hc.GetStats(),
		Uptime:   time.Since(startTime),
		LastTest: time.Now(),
	}

	// Check memory cache health
	health.Memory = hc.checkMemoryHealth()

	// Check Redis cache health
	health.Redis = hc.checkRedisHealth(ctx)

	// Determine overall health
	health.Overall = hc.determineOverallHealth(health.Memory, health.Redis)

	return health
}

// checkMemoryHealth evaluates in-memory cache health
func (hc *HybridCache) checkMemoryHealth() MemoryCacheHealth {
	health := MemoryCacheHealth{
		Enabled: hc.config.EnableMemory,
		Status:  "unhealthy",
	}

	if !hc.config.EnableMemory || hc.memoryCache == nil {
		health.Status = "disabled"
		return health
	}

	// Get memory cache statistics
	currentSize := hc.memoryCache.size()
	maxSize := hc.memoryCache.maxSize

	health.Size = currentSize
	health.MaxSize = maxSize
	health.Status = "healthy"

	if maxSize > 0 {
		health.UtilPct = float64(currentSize) / float64(maxSize) * 100

		// Consider it degraded if utilization is very high
		if health.UtilPct > 90 {
			health.Status = "degraded"
		}
	}

	return health
}

// checkRedisHealth evaluates Redis cache health
func (hc *HybridCache) checkRedisHealth(ctx context.Context) RedisCacheHealth {
	health := RedisCacheHealth{
		Enabled:   hc.config.EnableRedis,
		Status:    "unhealthy",
		Connected: false,
		Address:   hc.config.RedisAddr,
	}

	if !hc.config.EnableRedis || hc.redisCache == nil {
		health.Status = "disabled"
		return health
	}

	// Test Redis connection with ping
	start := time.Now()
	err := hc.redisCache.healthCheck(ctx)
	health.Latency = time.Since(start)

	if err != nil {
		health.Status = "unhealthy"
		health.Connected = false
		health.Error = err.Error()
	} else {
		health.Status = "healthy"
		health.Connected = true

		// Consider it degraded if latency is high
		if health.Latency > 50*time.Millisecond {
			health.Status = "degraded"
		}
	}

	return health
}

// determineOverallHealth calculates overall cache system health
func (hc *HybridCache) determineOverallHealth(memory MemoryCacheHealth, redis RedisCacheHealth) string {
	// If both are disabled, overall is unhealthy
	if (!memory.Enabled || memory.Status == "disabled") &&
		(!redis.Enabled || redis.Status == "disabled") {
		return "unhealthy"
	}

	// Count healthy/degraded components
	healthyComponents := 0
	degradedComponents := 0
	totalEnabled := 0

	if memory.Enabled && memory.Status != "disabled" {
		totalEnabled++
		if memory.Status == "healthy" {
			healthyComponents++
		} else if memory.Status == "degraded" {
			degradedComponents++
		}
	}

	if redis.Enabled && redis.Status != "disabled" {
		totalEnabled++
		if redis.Status == "healthy" {
			healthyComponents++
		} else if redis.Status == "degraded" {
			degradedComponents++
		}
	}

	// Determine overall status
	if healthyComponents == totalEnabled {
		return "healthy"
	} else if healthyComponents+degradedComponents == totalEnabled {
		return "degraded"
	} else {
		return "unhealthy"
	}
}
