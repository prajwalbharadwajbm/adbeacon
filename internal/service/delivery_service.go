package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
)

// CampaignDeliveryService defines the interface for campaign delivery service
type CampaignDeliveryService interface {
	GetCampaigns(ctx context.Context, req models.DeliveryRequest) ([]models.CampaignResponse, error)
}

// CampaignRepository interface for data access
type CampaignRepository interface {
	GetActiveCampaignsWithRules(ctx context.Context) ([]models.CampaignWithRules, error)
}

// DeliveryService handles ad delivery requests
type DeliveryService struct {
	repository CampaignRepository
	matcher    *models.CampaignMatcher
}

// NewDeliveryService creates a new delivery service
func NewDeliveryService(repo CampaignRepository) *DeliveryService {
	// Use the default campaign matcher with extensible dimension system
	registry := models.GetDimensionRegistry()
	matcher := models.NewCampaignMatcher(registry)

	return &DeliveryService{
		repository: repo,
		matcher:    matcher,
	}
}

// NewDeliveryServiceWithMatcher creates a delivery service with custom matcher
func NewDeliveryServiceWithMatcher(repo CampaignRepository, matcher *models.CampaignMatcher) *DeliveryService {
	return &DeliveryService{
		repository: repo,
		matcher:    matcher,
	}
}

// GetCampaigns finds all campaigns that match the delivery request
func (s *DeliveryService) GetCampaigns(ctx context.Context, req models.DeliveryRequest) ([]models.CampaignResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Normalize values for consistent comparison
	req.NormalizeValues()

	// Get all active campaigns with their targeting rules
	campaignsWithRules, err := s.repository.GetActiveCampaignsWithRules(ctx)
	if err != nil {
		return nil, errors.New("failed to retrieve campaigns")
	}

	// Filter campaigns that match the request using extensible matcher
	var matchingCampaigns []models.CampaignResponse
	for _, campaign := range campaignsWithRules {
		if s.matcher.MatchesRequest(campaign, req) {
			matchingCampaigns = append(matchingCampaigns, campaign.ToResponse())
		}
	}

	return matchingCampaigns, nil
}

// GetMatchingCampaigns finds campaigns that match the delivery request using extensible dimensions
func (ds *DeliveryService) GetMatchingCampaigns(ctx context.Context, req models.DeliveryRequest) ([]models.CampaignResponse, error) {
	// Build cache keys for all dimensions in the request
	cacheKeys := ds.buildCacheKeys(req)

	// Try to get campaigns from cache using multi-key lookup
	campaigns, err := ds.getCampaignsFromCache(ctx, cacheKeys)
	if err == nil && len(campaigns) > 0 {
		return ds.filterAndConvert(campaigns, req), nil
	}

	// Fallback to database
	campaignsWithRules, err := ds.repository.GetActiveCampaignsWithRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaigns: %w", err)
	}

	// Filter campaigns using extensible matcher
	var matchingCampaigns []models.CampaignWithRules
	for _, campaign := range campaignsWithRules {
		if ds.matcher.MatchesRequest(campaign, req) {
			matchingCampaigns = append(matchingCampaigns, campaign)
		}
	}

	// Cache the results for each dimension
	ds.cacheResultsForDimensions(ctx, req, matchingCampaigns)

	// Convert to response format
	responses := make([]models.CampaignResponse, len(matchingCampaigns))
	for i, campaign := range matchingCampaigns {
		responses[i] = campaign.ToResponse()
	}

	return responses, nil
}

// buildCacheKeys creates cache keys for all dimensions in the request
func (ds *DeliveryService) buildCacheKeys(req models.DeliveryRequest) []string {
	var keys []string

	// Get all registered dimension processors
	registry := ds.matcher.Registry
	if registry == nil {
		registry = models.GetDimensionRegistry()
	}

	processors := registry.GetAllProcessors()

	// Build cache key for each dimension that has a value in the request
	for dimensionName, processor := range processors {
		value := processor.GetValue(req)
		if value != "" {
			key := ds.matcher.BuildIndexKey(dimensionName, value)
			keys = append(keys, key)
		}
	}

	return keys
}

// getCampaignsFromCache attempts to retrieve campaigns from cache using multiple keys
func (ds *DeliveryService) getCampaignsFromCache(ctx context.Context, keys []string) ([]models.CampaignWithRules, error) {
	// This would use the cache system to do intersection of campaign IDs from multiple indexes
	// For now, return empty to indicate cache miss
	return nil, fmt.Errorf("cache miss")
}

// cacheResultsForDimensions caches the matching campaigns for each dimension in the request
func (ds *DeliveryService) cacheResultsForDimensions(ctx context.Context, req models.DeliveryRequest, campaigns []models.CampaignWithRules) {
	registry := ds.matcher.Registry
	if registry == nil {
		registry = models.GetDimensionRegistry()
	}

	processors := registry.GetAllProcessors()

	// Cache results for each dimension
	for dimensionName, processor := range processors {
		value := processor.GetValue(req)
		if value != "" {
			// Filter campaigns that match this specific dimension
			var dimensionCampaigns []models.CampaignWithRules
			for _, campaign := range campaigns {
				// Check if campaign has rules for this dimension
				hasRuleForDimension := false
				for _, rule := range campaign.Rules {
					if string(rule.Dimension) == dimensionName {
						hasRuleForDimension = true
						break
					}
				}

				// If no rules for this dimension, campaign matches everyone
				if !hasRuleForDimension || ds.campaignMatchesDimension(campaign, dimensionName, value, processor) {
					dimensionCampaigns = append(dimensionCampaigns, campaign)
				}
			}

			// Note: actual cache implementation would store dimensionCampaigns here
			// For now, this is just the indexing logic
		}
	}
}

// campaignMatchesDimension checks if a campaign matches a specific dimension value
func (ds *DeliveryService) campaignMatchesDimension(campaign models.CampaignWithRules, dimensionName, value string, processor models.DimensionProcessor) bool {
	// Find rules for this dimension
	var dimensionRules []models.TargetingRule
	for _, rule := range campaign.Rules {
		if string(rule.Dimension) == dimensionName {
			dimensionRules = append(dimensionRules, rule)
		}
	}

	// If no rules for this dimension, campaign matches
	if len(dimensionRules) == 0 {
		return true
	}

	// Use the processor to check if the value matches any rule
	for _, rule := range dimensionRules {
		if processor.MatchesRule(value, rule) {
			return rule.RuleType == models.RuleTypeInclude
		}
	}

	// If no include rules matched, check if it's not excluded
	for _, rule := range dimensionRules {
		if rule.RuleType == models.RuleTypeExclude && processor.MatchesRule(value, rule) {
			return false
		}
	}

	return true
}

// filterAndConvert filters cached campaigns and converts to response format
func (ds *DeliveryService) filterAndConvert(campaigns []models.CampaignWithRules, req models.DeliveryRequest) []models.CampaignResponse {
	var matching []models.CampaignResponse

	for _, campaign := range campaigns {
		if ds.matcher.MatchesRequest(campaign, req) {
			matching = append(matching, campaign.ToResponse())
		}
	}

	return matching
}

// RegisterCustomDimension allows registering new dimension processors at runtime
func (ds *DeliveryService) RegisterCustomDimension(processor models.DimensionProcessor) {
	if ds.matcher != nil && ds.matcher.Registry != nil {
		ds.matcher.Registry.RegisterProcessor(processor)
	} else {
		// Register in global registry
		models.RegisterCustomDimension(processor)
	}
}

// GetSupportedDimensions returns all supported targeting dimensions
func (ds *DeliveryService) GetSupportedDimensions() []string {
	if ds.matcher != nil && ds.matcher.Registry != nil {
		return ds.matcher.Registry.ListDimensions()
	}
	return models.GetSupportedDimensions()
}
