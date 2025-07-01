package cache

import (
	"context"
	"testing"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHybridCache_MemoryOnly(t *testing.T) {
	// Test with memory cache only
	config := CacheConfig{
		DefaultTTL:      time.Minute,
		MemoryCacheSize: 100,
		EnableMemory:    true,
		EnableRedis:     false,
	}

	cache, err := NewHybridCache(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test campaign caching
	campaigns := []models.CampaignWithRules{
		{
			Campaign: models.Campaign{
				ID:     "test1",
				Name:   "Test Campaign 1",
				Status: models.StatusActive,
			},
			Rules: []models.TargetingRule{
				{
					Dimension: models.DimensionCountry,
					RuleType:  models.RuleTypeInclude,
					Values:    []string{"US"},
				},
			},
		},
	}

	// Store in cache
	err = cache.SetActiveCampaigns(ctx, campaigns, time.Minute)
	assert.NoError(t, err)

	// Retrieve from cache
	cachedCampaigns, err := cache.GetActiveCampaigns(ctx)
	assert.NoError(t, err)
	assert.Equal(t, campaigns, cachedCampaigns)

	// Test cache stats
	stats := cache.GetStats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
}

func TestHybridCache_CacheIndexes(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      time.Minute,
		MemoryCacheSize: 100,
		EnableMemory:    true,
		EnableRedis:     false,
	}

	cache, err := NewHybridCache(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test campaign index caching
	campaignIDs := []string{"campaign1", "campaign2", "campaign3"}

	err = cache.SetCampaignIndex(ctx, models.DimensionCountry, "us", campaignIDs, time.Minute)
	assert.NoError(t, err)

	// Retrieve from cache
	cachedIDs, err := cache.GetCampaignIndex(ctx, models.DimensionCountry, "us")
	assert.NoError(t, err)
	assert.Equal(t, campaignIDs, cachedIDs)
}

func TestHybridCache_CacheMiss(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      time.Minute,
		MemoryCacheSize: 100,
		EnableMemory:    true,
		EnableRedis:     false,
	}

	cache, err := NewHybridCache(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Try to get non-existent data
	_, err = cache.GetActiveCampaigns(ctx)
	assert.Equal(t, ErrCacheMiss, err)

	// Check cache stats
	stats := cache.GetStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
}

func TestHybridCache_TTLExpiration(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      50 * time.Millisecond, // Very short TTL for testing
		MemoryCacheSize: 100,
		EnableMemory:    true,
		EnableRedis:     false,
	}

	cache, err := NewHybridCache(config)
	require.NoError(t, err)

	ctx := context.Background()

	campaigns := []models.CampaignWithRules{
		{
			Campaign: models.Campaign{
				ID:     "test1",
				Status: models.StatusActive,
			},
		},
	}

	// Store in cache
	err = cache.SetActiveCampaigns(ctx, campaigns, 50*time.Millisecond)
	assert.NoError(t, err)

	// Should be available immediately
	_, err = cache.GetActiveCampaigns(ctx)
	assert.NoError(t, err)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	_, err = cache.GetActiveCampaigns(ctx)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestHybridCache_InvalidateAll(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      time.Minute,
		MemoryCacheSize: 100,
		EnableMemory:    true,
		EnableRedis:     false,
	}

	cache, err := NewHybridCache(config)
	require.NoError(t, err)

	ctx := context.Background()

	campaigns := []models.CampaignWithRules{
		{
			Campaign: models.Campaign{
				ID:     "test1",
				Status: models.StatusActive,
			},
		},
	}

	// Store in cache
	err = cache.SetActiveCampaigns(ctx, campaigns, time.Minute)
	assert.NoError(t, err)

	// Verify it's there
	_, err = cache.GetActiveCampaigns(ctx)
	assert.NoError(t, err)

	// Invalidate all
	err = cache.InvalidateAll(ctx)
	assert.NoError(t, err)

	// Should be gone now
	_, err = cache.GetActiveCampaigns(ctx)
	assert.Equal(t, ErrCacheMiss, err)
}

// Benchmark tests to demonstrate performance improvements
func BenchmarkCacheHit_Memory(b *testing.B) {
	config := CacheConfig{
		DefaultTTL:      time.Hour,
		MemoryCacheSize: 1000,
		EnableMemory:    true,
		EnableRedis:     false,
	}

	cache, err := NewHybridCache(config)
	require.NoError(b, err)

	ctx := context.Background()

	// Pre-populate cache
	campaigns := make([]models.CampaignWithRules, 100)
	for i := range campaigns {
		campaigns[i] = models.CampaignWithRules{
			Campaign: models.Campaign{
				ID:     "campaign" + string(rune(i)),
				Status: models.StatusActive,
			},
		}
	}

	cache.SetActiveCampaigns(ctx, campaigns, time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cache.GetActiveCampaigns(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCacheMiss_Memory(b *testing.B) {
	config := CacheConfig{
		DefaultTTL:      time.Hour,
		MemoryCacheSize: 1000,
		EnableMemory:    true,
		EnableRedis:     false,
	}

	cache, err := NewHybridCache(config)
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cache.GetActiveCampaigns(ctx)
			if err != ErrCacheMiss {
				b.Fatal("Expected cache miss")
			}
		}
	})
}
