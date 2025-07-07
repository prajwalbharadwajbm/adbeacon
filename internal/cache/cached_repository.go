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

// GetCampaignsByRequest uses indexes for fast campaign lookup based on delivery request
func (cr *CachedRepository) GetCampaignsByRequest(ctx context.Context, req models.DeliveryRequest) ([]models.CampaignWithRules, error) {
	// Try to get candidate campaign IDs from indexes
	candidateIDs, err := cr.getCandidateIDs(ctx, req)
	if err != nil {
		// If index lookup fails, fallback to loading all campaigns
		return cr.GetActiveCampaignsWithRules(ctx)
	}

	// If we found candidate IDs, get only those campaigns
	if len(candidateIDs) > 0 {
		return cr.getCampaignsByIDs(ctx, candidateIDs)
	}

	// If no candidates found, return empty result
	return []models.CampaignWithRules{}, nil
}

// getCandidateIDs retrieves campaign IDs that match the request using indexes
func (cr *CachedRepository) getCandidateIDs(ctx context.Context, req models.DeliveryRequest) ([]string, error) {
	var candidateSets [][]string

	// Get campaigns that match country
	if req.Country != "" {
		countryIDs, err := cr.cache.GetCampaignIndex(ctx, models.DimensionCountry, req.Country)
		if err == nil && len(countryIDs) > 0 {
			candidateSets = append(candidateSets, countryIDs)
		}
	}

	// Get campaigns that match OS
	if req.OS != "" {
		osIDs, err := cr.cache.GetCampaignIndex(ctx, models.DimensionOS, req.OS)
		if err == nil && len(osIDs) > 0 {
			candidateSets = append(candidateSets, osIDs)
		}
	}

	// Get campaigns that match app
	if req.App != "" {
		appIDs, err := cr.cache.GetCampaignIndex(ctx, models.DimensionApp, req.App)
		if err == nil && len(appIDs) > 0 {
			candidateSets = append(candidateSets, appIDs)
		}
	}

	// If we have no index matches, return error to trigger fallback
	if len(candidateSets) == 0 {
		return nil, fmt.Errorf("no index matches found")
	}

	// Find union of all candidate sets (campaigns that match any dimension)
	// Note: We use union instead of intersection because:
	// 1. A campaign might not have rules for all dimensions (matches everything for that dimension)
	// 2. Final filtering will be done by the service layer
	candidateIDs := cr.unionSlices(candidateSets...)

	return candidateIDs, nil
}

// getCampaignsByIDs retrieves specific campaigns by their IDs
func (cr *CachedRepository) getCampaignsByIDs(ctx context.Context, campaignIDs []string) ([]models.CampaignWithRules, error) {
	// Get all campaigns from cache
	allCampaigns, err := cr.cache.GetActiveCampaigns(ctx)
	if err != nil {
		// If cache miss, get from database
		allCampaigns, err = cr.repo.GetActiveCampaignsWithRules(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Filter campaigns by IDs
	campaignMap := make(map[string]models.CampaignWithRules)
	for _, campaign := range allCampaigns {
		campaignMap[campaign.ID] = campaign
	}

	var filteredCampaigns []models.CampaignWithRules
	for _, id := range campaignIDs {
		if campaign, exists := campaignMap[id]; exists {
			filteredCampaigns = append(filteredCampaigns, campaign)
		}
	}

	return filteredCampaigns, nil
}

// unionSlices combines multiple slices and removes duplicates
func (cr *CachedRepository) unionSlices(slices ...[]string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, slice := range slices {
		for _, item := range slice {
			if !seen[item] {
				seen[item] = true
				result = append(result, item)
			}
		}
	}

	return result
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
