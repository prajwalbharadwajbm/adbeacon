package service

import (
	"context"
	"errors"

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

// OptimizedCampaignRepository extends CampaignRepository with fast lookup capabilities
type OptimizedCampaignRepository interface {
	CampaignRepository
	GetCampaignsByRequest(ctx context.Context, req models.DeliveryRequest) ([]models.CampaignWithRules, error)
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

	// Try optimized lookup first if repository supports it
	var campaignsWithRules []models.CampaignWithRules
	var err error

	if optimizedRepo, ok := s.repository.(OptimizedCampaignRepository); ok {
		// Use fast index-based lookup
		campaignsWithRules, err = optimizedRepo.GetCampaignsByRequest(ctx, req)
		if err != nil {
			return nil, errors.New("failed to retrieve campaigns")
		}
	} else {
		// Fallback to loading all campaigns
		campaignsWithRules, err = s.repository.GetActiveCampaignsWithRules(ctx)
		if err != nil {
			return nil, errors.New("failed to retrieve campaigns")
		}
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
