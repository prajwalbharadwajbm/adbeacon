package models

import (
	"errors"
	"slices"
	"strings"
)

// CountryProcessor handles country-based targeting
type CountryProcessor struct{}

func NewCountryProcessor() DimensionProcessor {
	return &CountryProcessor{}
}

func (cp *CountryProcessor) GetName() string {
	return "country"
}

func (cp *CountryProcessor) GetValue(req DeliveryRequest) string {
	return req.Country
}

func (cp *CountryProcessor) NormalizeValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (cp *CountryProcessor) ValidateRule(rule TargetingRule) error {
	if len(rule.Values) == 0 {
		return errors.New("country rule must have at least one value")
	}

	// Validate country codes (basic validation)
	for _, value := range rule.Values {
		if len(strings.TrimSpace(value)) < 2 {
			return errors.New("country code must be at least 2 characters")
		}
	}

	return nil
}

func (cp *CountryProcessor) MatchesRule(requestValue string, rule TargetingRule) bool {
	normalizedRequest := cp.NormalizeValue(requestValue)

	for _, ruleValue := range rule.Values {
		normalizedRule := cp.NormalizeValue(ruleValue)
		if normalizedRequest == normalizedRule {
			return true
		}
	}

	return false
}

// OSProcessor handles operating system targeting
type OSProcessor struct{}

func NewOSProcessor() DimensionProcessor {
	return &OSProcessor{}
}

func (osp *OSProcessor) GetName() string {
	return "os"
}

func (osp *OSProcessor) GetValue(req DeliveryRequest) string {
	return req.OS
}

func (osp *OSProcessor) NormalizeValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (osp *OSProcessor) ValidateRule(rule TargetingRule) error {
	if len(rule.Values) == 0 {
		return errors.New("os rule must have at least one value")
	}

	// Validate known OS values
	validOS := []string{"android", "ios", "windows", "macos", "linux", "web"}
	for _, value := range rule.Values {
		normalized := osp.NormalizeValue(value)
		if !slices.Contains(validOS, normalized) {
			// Allow unknown OS values but warn
			continue
		}
	}

	return nil
}

func (osp *OSProcessor) MatchesRule(requestValue string, rule TargetingRule) bool {
	normalizedRequest := osp.NormalizeValue(requestValue)

	for _, ruleValue := range rule.Values {
		normalizedRule := osp.NormalizeValue(ruleValue)
		if normalizedRequest == normalizedRule {
			return true
		}
	}

	return false
}

// AppProcessor handles application ID targeting
type AppProcessor struct{}

func NewAppProcessor() DimensionProcessor {
	return &AppProcessor{}
}

func (ap *AppProcessor) GetName() string {
	return "app"
}

func (ap *AppProcessor) GetValue(req DeliveryRequest) string {
	return req.App
}

func (ap *AppProcessor) NormalizeValue(value string) string {
	// App IDs are case-sensitive, only trim whitespace
	return strings.TrimSpace(value)
}

func (ap *AppProcessor) ValidateRule(rule TargetingRule) error {
	if len(rule.Values) == 0 {
		return errors.New("app rule must have at least one value")
	}

	// Validate app ID format (basic validation)
	for _, value := range rule.Values {
		trimmed := strings.TrimSpace(value)
		if len(trimmed) == 0 {
			return errors.New("app ID cannot be empty")
		}

		// Basic app ID format validation (com.company.app or bundle.id)
		if !strings.Contains(trimmed, ".") {
			return errors.New("app ID should follow package naming convention (e.g., com.company.app)")
		}
	}

	return nil
}

func (ap *AppProcessor) MatchesRule(requestValue string, rule TargetingRule) bool {
	normalizedRequest := ap.NormalizeValue(requestValue)

	for _, ruleValue := range rule.Values {
		normalizedRule := ap.NormalizeValue(ruleValue)
		if normalizedRequest == normalizedRule {
			return true
		}
	}

	return false
}
