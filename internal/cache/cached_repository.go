package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
)

// CachedRepository wraps a repository with caching capabilities
type CachedRepository struct {
	repo  service.CampaignRepository
	cache Cache
	ttl   time.Duration
}

// NewCachedRepository creates a new cached repository
func NewCachedRepository(repo service.CampaignRepository, cache Cache, ttl time.Duration) service.CampaignRepository {
	return &CachedRepository{
		repo:  repo,
		cache: cache,
		ttl:   ttl,
	}
}

// GetActiveCampaignsWithRules retrieves campaigns from cache first, then database
func (cr *CachedRepository) GetActiveCampaignsWithRules(ctx context.Context) ([]models.CampaignWithRules, error) {
	// Try cache first
	campaigns, err := cr.cache.GetActiveCampaigns(ctx)
	if err == nil {
		return campaigns, nil
	}

	// If cache miss, get from database
	campaigns, err = cr.repo.GetActiveCampaignsWithRules(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache for next time (async to not block the response)
	go func() {
		// Use a new context to avoid timeout issues
		cacheCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := cr.cache.SetActiveCampaigns(cacheCtx, campaigns, cr.ttl); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to cache campaigns: %v\n", err)
		}

		// Also build and cache indexes for faster lookups
		cr.buildAndCacheIndexes(cacheCtx, campaigns)
	}()

	return campaigns, nil
}

// buildAndCacheIndexes creates pre-computed indexes for fast campaign lookups
func (cr *CachedRepository) buildAndCacheIndexes(ctx context.Context, campaigns []models.CampaignWithRules) {
	// Build indexes by targeting dimensions
	countryIndex := make(map[string][]string)
	osIndex := make(map[string][]string)
	appIndex := make(map[string][]string)

	for _, campaign := range campaigns {
		if !campaign.IsActive() {
			continue
		}

		for _, rule := range campaign.Rules {
			if rule.RuleType != models.RuleTypeInclude {
				continue // Only index include rules for now
			}

			switch rule.Dimension {
			case models.DimensionCountry:
				for _, value := range rule.NormalizeValues() {
					countryIndex[value] = append(countryIndex[value], campaign.ID)
				}
			case models.DimensionOS:
				for _, value := range rule.NormalizeValues() {
					osIndex[value] = append(osIndex[value], campaign.ID)
				}
			case models.DimensionApp:
				for _, value := range rule.Values { // Don't normalize app IDs
					appIndex[value] = append(appIndex[value], campaign.ID)
				}
			}
		}
	}

	// Cache the indexes
	indexTTL := cr.ttl + time.Minute // Index TTL slightly longer than campaign TTL

	// Cache country indexes
	for country, campaignIDs := range countryIndex {
		cr.cache.SetCampaignIndex(ctx, models.DimensionCountry, country, campaignIDs, indexTTL)
	}

	// Cache OS indexes
	for os, campaignIDs := range osIndex {
		cr.cache.SetCampaignIndex(ctx, models.DimensionOS, os, campaignIDs, indexTTL)
	}

	// Cache app indexes
	for app, campaignIDs := range appIndex {
		cr.cache.SetCampaignIndex(ctx, models.DimensionApp, app, campaignIDs, indexTTL)
	}
}

// InvalidateCache clears all cached data
func (cr *CachedRepository) InvalidateCache(ctx context.Context) error {
	return cr.cache.InvalidateAll(ctx)
}

// GetCacheStats returns cache performance statistics
func (cr *CachedRepository) GetCacheStats() CacheStats {
	return cr.cache.GetStats()
}
