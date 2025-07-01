package models

import (
	"errors"
	"slices"
	"strings"
)

// DeliveryRequest represents a request for ad delivery
type DeliveryRequest struct {
	Country string `json:"country" validate:"required,len=2"`
	OS      string `json:"os" validate:"required,oneof=android ios"`
	App     string `json:"app" validate:"required"`
}

// Validate validates the delivery request
func (dr *DeliveryRequest) Validate() error {
	if dr.Country == "" {
		return errors.New("country is required")
	}
	if len(dr.Country) != 2 {
		return errors.New("country must be a 2-letter code")
	}
	if dr.OS == "" {
		return errors.New("os is required")
	}
	if dr.App == "" {
		return errors.New("app is required")
	}
	return nil
}

// NormalizeValues normalizes request values for consistent comparison
func (dr *DeliveryRequest) NormalizeValues() {
	dr.Country = strings.ToLower(strings.TrimSpace(dr.Country))
	dr.OS = strings.ToLower(strings.TrimSpace(dr.OS))
	dr.App = strings.TrimSpace(dr.App) // App IDs are case-sensitive
}

// ToMap converts the request to a map for extensible dimension processing
func (dr *DeliveryRequest) ToMap() map[string]string {
	return map[string]string{
		"country": dr.Country,
		"os":      dr.OS,
		"app":     dr.App,
	}
}

// GetDimensionValue gets a value for a specific dimension using the extensible system
func (dr *DeliveryRequest) GetDimensionValue(dimension string) string {
	switch dimension {
	case "country":
		return dr.Country
	case "os":
		return dr.OS
	case "app":
		return dr.App
	default:
		// For extensible dimensions, return empty (can be extended later)
		return ""
	}
}

// MatchesRule checks if a targeting rule applies to this request.
// For include rules: returns true if the request value is in the rule's allowed values
// For exclude rules: returns true if the request value is in the rule's excluded values
// This method determines if the rule "triggers" for this request, not if the request "passes" the rule
func (r *DeliveryRequest) MatchesRule(rule TargetingRule) bool {
	requestValue := r.GetDimensionValue(string(rule.Dimension))
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
