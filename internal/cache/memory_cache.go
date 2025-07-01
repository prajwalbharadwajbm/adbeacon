package cache

import (
	"sync"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
)

// cacheItem represents a cached item with expiration
type cacheItem struct {
	data      any
	expiresAt time.Time
}

// isExpired checks if the cache item has expired
func (ci *cacheItem) isExpired() bool {
	return time.Now().After(ci.expiresAt)
}

// memoryCache implements in-memory caching with TTL
type memoryCache struct {
	items    map[string]*cacheItem
	mu       sync.RWMutex
	maxSize  int
	stopChan chan struct{}
}

// newMemoryCache creates a new in-memory cache
func newMemoryCache(maxSize int) *memoryCache {
	mc := &memoryCache{
		items:    make(map[string]*cacheItem),
		maxSize:  maxSize,
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go mc.cleanup()

	return mc
}

// getActiveCampaigns retrieves active campaigns from memory cache
func (mc *memoryCache) getActiveCampaigns() ([]models.CampaignWithRules, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.items["active_campaigns"]
	if !exists || item.isExpired() {
		return nil, false
	}

	campaigns, ok := item.data.([]models.CampaignWithRules)
	return campaigns, ok
}

// setActiveCampaigns stores active campaigns in memory cache
func (mc *memoryCache) setActiveCampaigns(campaigns []models.CampaignWithRules, ttl time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items["active_campaigns"] = &cacheItem{
		data:      campaigns,
		expiresAt: time.Now().Add(ttl),
	}

	// Check if we need to evict items
	mc.evictIfNeeded()
}

// getCampaignIndex retrieves campaign index from memory cache
func (mc *memoryCache) getCampaignIndex(key string) ([]string, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.items[key]
	if !exists || item.isExpired() {
		return nil, false
	}

	campaignIDs, ok := item.data.([]string)
	return campaignIDs, ok
}

// setCampaignIndex stores campaign index in memory cache
func (mc *memoryCache) setCampaignIndex(key string, campaignIDs []string, ttl time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items[key] = &cacheItem{
		data:      campaignIDs,
		expiresAt: time.Now().Add(ttl),
	}

	// Check if we need to evict items
	mc.evictIfNeeded()
}

// clear removes all items from memory cache
func (mc *memoryCache) clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items = make(map[string]*cacheItem)
}

// evictIfNeeded removes expired items and enforces max size
func (mc *memoryCache) evictIfNeeded() {
	// Remove expired items first
	for key, item := range mc.items {
		if item.isExpired() {
			delete(mc.items, key)
		}
	}

	// If still over max size, remove oldest items (simple FIFO for now)
	if len(mc.items) > mc.maxSize {
		count := len(mc.items) - mc.maxSize
		for key := range mc.items {
			if count <= 0 {
				break
			}
			delete(mc.items, key)
			count--
		}
	}
}

// cleanup periodically removes expired items
func (mc *memoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.mu.Lock()
			for key, item := range mc.items {
				if item.isExpired() {
					delete(mc.items, key)
				}
			}
			mc.mu.Unlock()
		case <-mc.stopChan:
			return
		}
	}
}

// close stops the cleanup goroutine
func (mc *memoryCache) close() {
	close(mc.stopChan)
}

// size returns the current number of items in cache
func (mc *memoryCache) size() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return len(mc.items)
}
