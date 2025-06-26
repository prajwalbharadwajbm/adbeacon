package models

import (
	"errors"
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
	return td == DimensionCountry || td == DimensionOS || td == DimensionApp
}

func (rt RuleType) IsValid() bool {
	return rt == RuleTypeInclude || rt == RuleTypeExclude
}

// Validate checks if targeting rule is valid
func (tr *TargetingRule) Validate() error {
	if tr.CampaignID == "" {
		return errors.New("campaign_id is required")
	}

	if !tr.Dimension.IsValid() {
		return errors.New("invalid dimension")
	}

	if !tr.RuleType.IsValid() {
		return errors.New("invalid rule_type")
	}

	if len(tr.Values) == 0 {
		return errors.New("values cannot be empty")
	}

	return nil
}

// NormalizeValues cleans/normalizes rule values for standard comparison
func (tr *TargetingRule) NormalizeValues() []string {
	normalized := make([]string, len(tr.Values))
	for i, v := range tr.Values {
		normalized[i] = strings.ToLower(strings.TrimSpace(v))
	}
	return normalized
}
