package models

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// TargetingRule defines where a campaign can actually run.
type TargetingRule struct {
	ID         int64           `json:"id" db:"id"`
	CampaignID string          `json:"campaign_id" db:"campaign_id"`
	Dimension  TargetDimension `json:"dimension" db:"dimension"`
	RuleType   RuleType        `json:"rule_type" db:"rule_type"`
	Values     []string        `json:"values" db:"values"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// TargetDimension represents targeting dimensions
type TargetDimension string

// enum values for TargetDimension
const (
	DimensionCountry TargetDimension = "country"
	DimensionOS      TargetDimension = "os"
	DimensionApp     TargetDimension = "app"
	DimensionState   TargetDimension = "state"

	// Extended dimensions (examples)
	DimensionDeviceType TargetDimension = "device_type"
	DimensionAgeGroup   TargetDimension = "age_group"
	DimensionTimeOfDay  TargetDimension = "time_of_day"
)

// RuleType represents include/exclude rule types
type RuleType string

// enum values for RuleType
const (
	RuleTypeInclude RuleType = "include"
	RuleTypeExclude RuleType = "exclude"
)

// IsValid methods for validation
func (td TargetDimension) IsValid() bool {
	// Use the extensible system for validation
	registry := GetDimensionRegistry()
	_, exists := registry.GetProcessor(string(td))
	return exists
}

func (rt RuleType) IsValid() bool {
	return rt == RuleTypeInclude || rt == RuleTypeExclude
}

// Validate checks if targeting rule is valid using the extensible system
func (tr *TargetingRule) Validate() error {
	if tr.CampaignID == "" {
		return errors.New("campaign_id is required")
	}

	if !tr.RuleType.IsValid() {
		return errors.New("invalid rule_type")
	}

	if len(tr.Values) == 0 {
		return errors.New("values cannot be empty")
	}

	// Use extensible validation
	registry := GetDimensionRegistry()
	processor, exists := registry.GetProcessor(string(tr.Dimension))
	if !exists {
		return fmt.Errorf("unknown dimension: %s", tr.Dimension)
	}

	return processor.ValidateRule(*tr)
}

// NormalizeValues cleans/normalizes rule values using the appropriate processor
func (tr *TargetingRule) NormalizeValues() []string {
	registry := GetDimensionRegistry()
	processor, exists := registry.GetProcessor(string(tr.Dimension))
	if !exists {
		// Fallback to basic normalization
		normalized := make([]string, len(tr.Values))
		for i, v := range tr.Values {
			normalized[i] = strings.ToLower(strings.TrimSpace(v))
		}
		return normalized
	}

	// Use processor-specific normalization
	normalized := make([]string, len(tr.Values))
	for i, v := range tr.Values {
		normalized[i] = processor.NormalizeValue(v)
	}
	return normalized
}

// GetSupportedDimensions returns all supported dimensions
func GetSupportedDimensions() []string {
	registry := GetDimensionRegistry()
	return registry.ListDimensions()
}

// RegisterCustomDimension adds a new dimension processor to the global registry
func RegisterCustomDimension(processor DimensionProcessor) {
	registry := GetDimensionRegistry()
	registry.RegisterProcessor(processor)
}

// CreateCustomTargetDimension creates a new custom dimension
func CreateCustomTargetDimension(name string) TargetDimension {
	return TargetDimension(name)
}
