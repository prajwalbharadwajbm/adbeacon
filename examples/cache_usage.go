package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/cache"
	"github.com/prajwalbharadwajbm/adbeacon/internal/config"
	"github.com/prajwalbharadwajbm/adbeacon/internal/database"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/prajwalbharadwajbm/adbeacon/internal/repository"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
)

// This example shows how to set up and use the hybrid cache system

func ExampleBasicCacheUsage() {
	fmt.Println("=== Basic Cache Usage Example ===")

	// 1. Create cache configuration
	cacheConfig := cache.CacheConfig{
		DefaultTTL:      5 * time.Minute,
		MemoryCacheSize: 1000,
		RedisAddr:       "localhost:6379",
		RedisPassword:   "",
		RedisDB:         0,
		EnableMemory:    true,
		EnableRedis:     true,
		RefreshInterval: 1 * time.Minute,
	}

	// 2. Initialize hybrid cache
	hybridCache, err := cache.NewHybridCache(cacheConfig)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	// 3. Use the cache directly
	// Example: Check cache stats
	stats := hybridCache.GetStats()
	fmt.Printf("Cache Stats: Hits: %d, Misses: %d, Hit Ratio: %.2f%%\n",
		stats.Hits, stats.Misses, stats.HitRatio*100)
}

func ExampleCachedRepositorySetup() {
	fmt.Println("=== Cached Repository Setup Example ===")

	// Initialize database (in real usage, you'd pass actual config)
	db, cleanup, err := database.Initialize(config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "adbeacon_dev_user",
		Password: "adbeacon1234",
		DBName:   "adbeacon",
		SSLMode:  "disable",
	}, "./migrations")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer cleanup()

	// Create cache
	cacheConfig := config.GetCacheConfig()
	hybridCache, err := cache.NewHybridCache(cacheConfig)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	// Layer 1: Base repository
	baseRepo := repository.NewPostgresRepository(db)

	// Layer 2: Add caching
	cachedRepo := cache.NewCachedRepository(
		baseRepo,
		hybridCache,
		5*time.Minute, // TTL
	)

	// Layer 3: Create service with cached repository
	deliveryService := service.NewDeliveryService(cachedRepo)

	// Now your service automatically uses caching!
	ctx := context.Background()
	campaigns, err := deliveryService.GetCampaigns(ctx, models.DeliveryRequest{
		App:     "com.example.app",
		Country: "US",
		OS:      "Android",
	})

	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d matching campaigns\n", len(campaigns))
	}

	// Check cache performance
	stats := hybridCache.GetStats()
	fmt.Printf("Cache Performance: %.2f%% hit ratio\n", stats.HitRatio*100)
}

func ExampleCacheInvalidation() {
	fmt.Println("=== Cache Invalidation Example ===")

	cacheConfig := config.GetCacheConfig()
	hybridCache, err := cache.NewHybridCache(cacheConfig)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	ctx := context.Background()

	// Example: Clear all cache when campaigns are updated
	if err := hybridCache.InvalidateAll(ctx); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	} else {
		fmt.Println("Cache invalidated successfully")
	}
}

// Performance comparison example
func ExamplePerformanceComparison() {
	fmt.Println("=== Performance Comparison Example ===")

	// Simulate performance improvements
	fmt.Println("Performance Improvements with Caching:")
	fmt.Println("")

	fmt.Println("Without Cache:")
	fmt.Println("  - Database query: ~50ms")
	fmt.Println("  - Campaign filtering: ~10ms")
	fmt.Println("  - Total per request: ~60ms")
	fmt.Println("  - 1 billion requests = 60 billion ms = 16,667 hours of DB time!")
	fmt.Println("")

	fmt.Println("With In-Memory Cache (99% hit rate):")
	fmt.Println("  - Cache hit: ~0.1ms")
	fmt.Println("  - Cache miss: ~60ms")
	fmt.Println("  - Average per request: ~0.69ms")
	fmt.Println("  - 1 billion requests = 690 million ms = 192 hours")
	fmt.Println("  - Improvement: 87x faster!")
	fmt.Println("")

	fmt.Println("With Hybrid Cache + Indexing (99.9% index hit rate):")
	fmt.Println("  - Index lookup: ~0.05ms")
	fmt.Println("  - Traditional lookup: ~0.69ms")
	fmt.Println("  - Average per request: ~0.051ms")
	fmt.Println("  - 1 billion requests = 51 million ms = 14 hours")
	fmt.Println("  - Improvement: 1,200x faster!")
}

// Environment configuration example
func ExampleEnvironmentConfig() {
	fmt.Println("=== Environment Configuration Example ===")

	fmt.Println("Set these environment variables:")
	fmt.Println("export CACHE_DEFAULT_TTL=5m")
	fmt.Println("export CACHE_MEMORY_SIZE=1000")
	fmt.Println("export REDIS_ADDR=localhost:6379")
	fmt.Println("export REDIS_PASSWORD=your_password")
	fmt.Println("export REDIS_DB=0")
	fmt.Println("export CACHE_ENABLE_MEMORY=true")
	fmt.Println("export CACHE_ENABLE_REDIS=true")
	fmt.Println("export CACHE_REFRESH_INTERVAL=1m")

	// Load configuration
	cacheConfig := config.GetCacheConfig()
	fmt.Printf("Loaded config: TTL=%v, Memory=%d, Redis=%s\n",
		cacheConfig.DefaultTTL, cacheConfig.MemoryCacheSize, cacheConfig.RedisAddr)
}
