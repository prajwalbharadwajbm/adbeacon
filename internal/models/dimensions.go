package models

import (
	"fmt"
)

// DimensionProcessor defines the interface for processing targeting dimensions
type DimensionProcessor interface {
	// GetName returns the dimension name (e.g., "country", "os", "app")
	GetName() string

	// GetValue extracts the dimension value from a delivery request
	GetValue(req DeliveryRequest) string

	// NormalizeValue normalizes a value for consistent comparison
	NormalizeValue(value string) string

	// ValidateRule checks if a targeting rule is valid for this dimension
	ValidateRule(rule TargetingRule) error

	// MatchesRule checks if a request value matches a targeting rule
	MatchesRule(requestValue string, rule TargetingRule) bool
}

// DimensionRegistry manages all available dimension processors
type DimensionRegistry struct {
	processors map[string]DimensionProcessor
}

// NewDimensionRegistry creates a new dimension registry with built-in processors
func NewDimensionRegistry() *DimensionRegistry {
	registry := &DimensionRegistry{
		processors: make(map[string]DimensionProcessor),
	}

	// Register built-in dimension processors
	registry.RegisterProcessor(NewCountryProcessor())
	registry.RegisterProcessor(NewOSProcessor())
	registry.RegisterProcessor(NewAppProcessor())

	return registry
}

// RegisterProcessor adds a new dimension processor to the registry
func (dr *DimensionRegistry) RegisterProcessor(processor DimensionProcessor) {
	dr.processors[processor.GetName()] = processor
}

// GetProcessor retrieves a dimension processor by name
func (dr *DimensionRegistry) GetProcessor(dimensionName string) (DimensionProcessor, bool) {
	processor, exists := dr.processors[dimensionName]
	return processor, exists
}

// GetAllProcessors returns all registered dimension processors
func (dr *DimensionRegistry) GetAllProcessors() map[string]DimensionProcessor {
	result := make(map[string]DimensionProcessor)
	for name, processor := range dr.processors {
		result[name] = processor
	}
	return result
}

// ListDimensions returns all available dimension names
func (dr *DimensionRegistry) ListDimensions() []string {
	dimensions := make([]string, 0, len(dr.processors))
	for name := range dr.processors {
		dimensions = append(dimensions, name)
	}
	return dimensions
}

// CampaignMatcher provides extensible campaign matching using dimension processors
type CampaignMatcher struct {
	Registry *DimensionRegistry
}

// NewCampaignMatcher creates a new campaign matcher with the given registry
func NewCampaignMatcher(registry *DimensionRegistry) *CampaignMatcher {
	return &CampaignMatcher{
		Registry: registry,
	}
}

// MatchesRequest checks if a campaign matches a delivery request using all registered processors
func (cm *CampaignMatcher) MatchesRequest(campaign CampaignWithRules, req DeliveryRequest) bool {
	// Only active campaigns can match
	if !campaign.IsActive() {
		return false
	}

	// If no rules exist, campaign matches everyone
	if len(campaign.Rules) == 0 {
		return true
	}

	// Group rules by dimension
	rulesByDimension := make(map[string][]TargetingRule)
	for _, rule := range campaign.Rules {
		dimensionName := string(rule.Dimension)
		rulesByDimension[dimensionName] = append(rulesByDimension[dimensionName], rule)
	}

	// Check each dimension using its processor
	for dimensionName, rules := range rulesByDimension {
		processor, exists := cm.Registry.GetProcessor(dimensionName)
		if !exists {
			// Skip unknown dimensions (backward compatibility)
			continue
		}

		if !cm.dimensionMatches(req, rules, processor) {
			return false
		}
	}

	return true
}

// dimensionMatches checks if request matches rules for a specific dimension using its processor
func (cm *CampaignMatcher) dimensionMatches(req DeliveryRequest, rules []TargetingRule, processor DimensionProcessor) bool {
	var includeRules, excludeRules []TargetingRule

	// Separate include and exclude rules
	for _, rule := range rules {
		switch rule.RuleType {
		case RuleTypeInclude:
			includeRules = append(includeRules, rule)
		case RuleTypeExclude:
			excludeRules = append(excludeRules, rule)
		}
	}

	// Get the request value for this dimension
	requestValue := processor.GetValue(req)
	if requestValue == "" {
		return len(includeRules) == 0 // No value means only match if no include rules
	}

	// If there are include rules, request must match at least one
	if len(includeRules) > 0 {
		matched := false
		for _, rule := range includeRules {
			if processor.MatchesRule(requestValue, rule) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// If there are exclude rules, request must not match any
	for _, rule := range excludeRules {
		if processor.MatchesRule(requestValue, rule) {
			return false
		}
	}

	return true
}

// ValidateTargetingRule validates a targeting rule using the appropriate processor
func (cm *CampaignMatcher) ValidateTargetingRule(rule TargetingRule) error {
	processor, exists := cm.Registry.GetProcessor(string(rule.Dimension))
	if !exists {
		return fmt.Errorf("unknown dimension: %s", rule.Dimension)
	}

	return processor.ValidateRule(rule)
}

// BuildIndexKey creates a cache index key for a dimension and value
func (cm *CampaignMatcher) BuildIndexKey(dimensionName, value string) string {
	processor, exists := cm.Registry.GetProcessor(dimensionName)
	if !exists {
		return fmt.Sprintf("index:%s:%s", dimensionName, value)
	}

	normalizedValue := processor.NormalizeValue(value)
	return fmt.Sprintf("index:%s:%s", dimensionName, normalizedValue)
}
