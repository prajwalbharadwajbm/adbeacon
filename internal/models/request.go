package models

import (
	"errors"
	"slices"
	"strings"
)

// DeliveryRequest represents the incoming targeting request
type DeliveryRequest struct {
	App     string `json:"app" validate:"required"`
	Country string `json:"country" validate:"required"`
	OS      string `json:"os" validate:"required"`
}

// Validate checks if the request has all required parameters
func (r *DeliveryRequest) Validate() error {
	if strings.TrimSpace(r.App) == "" {
		return errors.New("missing app param")
	}
	if strings.TrimSpace(r.Country) == "" {
		return errors.New("missing country param")
	}
	if strings.TrimSpace(r.OS) == "" {
		return errors.New("missing os param")
	}
	return nil
}

// Normalize converts request values to lowercase for consistent comparison
func (r *DeliveryRequest) Normalize() {
	r.App = strings.TrimSpace(r.App)
	r.Country = strings.ToLower(strings.TrimSpace(r.Country))
	r.OS = strings.ToLower(strings.TrimSpace(r.OS))
}

// GetDimensionValue returns the value for a specific targeting dimension
func (r *DeliveryRequest) GetDimensionValue(dimension TargetDimension) string {
	switch dimension {
	case DimensionApp:
		return r.App
	case DimensionCountry:
		return strings.ToLower(r.Country)
	case DimensionOS:
		return strings.ToLower(r.OS)
	default:
		return ""
	}
}

// MatchesRule checks if a targeting rule applies to this request.
// For include rules: returns true if the request value is in the rule's allowed values
// For exclude rules: returns true if the request value is in the rule's excluded values
// This method determines if the rule "triggers" for this request, not if the request "passes" the rule
func (r *DeliveryRequest) MatchesRule(rule TargetingRule) bool {
	requestValue := r.GetDimensionValue(rule.Dimension)
	if requestValue == "" {
		return false
	}

	// Normalize rule values for comparison
	normalizedValues := rule.NormalizeValues()

	// Check if request value exists in rule values
	valueInRuleList := slices.Contains(normalizedValues, requestValue)

	// Return whether this rule applies to the request
	switch rule.RuleType {
	case RuleTypeInclude:
		// Include rule applies if the value is in the allowed list
		return valueInRuleList
	case RuleTypeExclude:
		// Exclude rule applies if the value is in the excluded list
		return valueInRuleList
	default:
		return false
	}
}
