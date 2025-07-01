package models

import (
	"fmt"
)

// CampaignWithRules represents a campaign with its targeting rules
type CampaignWithRules struct {
	Campaign
	Rules []TargetingRule `json:"rules,omitempty"`
}

// Global campaign matcher instance (can be configured)
var defaultCampaignMatcher *CampaignMatcher

func init() {
	// Initialize default campaign matcher with built-in processors
	registry := NewDimensionRegistry()
	defaultCampaignMatcher = NewCampaignMatcher(registry)
}

// MatchesRequest checks if this campaign matches the delivery request using the extensible system
func (cwr *CampaignWithRules) MatchesRequest(req DeliveryRequest) bool {
	return defaultCampaignMatcher.MatchesRequest(*cwr, req)
}

// MatchesRequestWithMatcher checks if this campaign matches using a custom matcher
func (cwr *CampaignWithRules) MatchesRequestWithMatcher(req DeliveryRequest, matcher *CampaignMatcher) bool {
	return matcher.MatchesRequest(*cwr, req)
}

// ValidateRules validates all targeting rules for this campaign
func (cwr *CampaignWithRules) ValidateRules() []error {
	var errors []error

	for i, rule := range cwr.Rules {
		if err := defaultCampaignMatcher.ValidateTargetingRule(rule); err != nil {
			errors = append(errors, fmt.Errorf("rule %d: %w", i, err))
		}
	}

	return errors
}

// GetDimensionRegistry returns the default dimension registry
func GetDimensionRegistry() *DimensionRegistry {
	return defaultCampaignMatcher.Registry
}

// SetDefaultCampaignMatcher sets a custom campaign matcher as default
func SetDefaultCampaignMatcher(matcher *CampaignMatcher) {
	defaultCampaignMatcher = matcher
}

// ToResponse converts Campaign to CampaignResponse
func (cwr *CampaignWithRules) ToResponse() CampaignResponse {
	return cwr.Campaign.ToResponse()
}
