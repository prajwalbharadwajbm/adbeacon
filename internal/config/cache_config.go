package config

import (
	"os"
	"strconv"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/cache"
)

// GetCacheConfig creates cache configuration from environment variables
func GetCacheConfig() cache.CacheConfig {
	return cache.CacheConfig{
		DefaultTTL:      getDurationEnv("CACHE_DEFAULT_TTL", 5*time.Minute),
		MemoryCacheSize: getIntEnv("CACHE_MEMORY_SIZE", 1000),
		RedisAddr:       getStringEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getStringEnv("REDIS_PASSWORD", ""),
		RedisDB:         getIntEnv("REDIS_DB", 0),
		EnableMemory:    getBoolEnv("CACHE_ENABLE_MEMORY", true),
		EnableRedis:     getBoolEnv("CACHE_ENABLE_REDIS", true),
		RefreshInterval: getDurationEnv("CACHE_REFRESH_INTERVAL", 1*time.Minute),
	}
}

// Helper functions for environment variable parsing
func getStringEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// CacheHealthCheck represents cache health status
type CacheHealthCheck struct {
	Memory struct {
		Enabled bool `json:"enabled"`
		Size    int  `json:"size"`
	} `json:"memory"`
	Redis struct {
		Enabled   bool   `json:"enabled"`
		Connected bool   `json:"connected"`
		Address   string `json:"address"`
	} `json:"redis"`
	Stats cache.CacheStats `json:"stats"`
}

// GetCacheHealth returns current cache health status
func GetCacheHealth(cache cache.Cache) CacheHealthCheck {
	config := GetCacheConfig()
	health := CacheHealthCheck{}

	// Memory cache info
	health.Memory.Enabled = config.EnableMemory
	health.Memory.Size = config.MemoryCacheSize

	// Redis cache info
	health.Redis.Enabled = config.EnableRedis
	health.Redis.Address = config.RedisAddr
	// Note: would need additional interface methods to check Redis connection

	// Cache statistics
	health.Stats = cache.GetStats()

	return health
}
