package models

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Example custom dimension processors to demonstrate extensibility

// DeviceTypeProcessor handles device type targeting (mobile, tablet, desktop)
type DeviceTypeProcessor struct{}

func NewDeviceTypeProcessor() DimensionProcessor {
	return &DeviceTypeProcessor{}
}

func (dtp *DeviceTypeProcessor) GetName() string {
	return "device_type"
}

func (dtp *DeviceTypeProcessor) GetValue(req DeliveryRequest) string {
	// This would need to be added to DeliveryRequest struct
	// For now, return empty (can be extended later)
	return ""
}

func (dtp *DeviceTypeProcessor) NormalizeValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (dtp *DeviceTypeProcessor) ValidateRule(rule TargetingRule) error {
	if len(rule.Values) == 0 {
		return errors.New("device_type rule must have at least one value")
	}

	validTypes := []string{"mobile", "tablet", "desktop"}
	for _, value := range rule.Values {
		normalized := dtp.NormalizeValue(value)
		found := false
		for _, valid := range validTypes {
			if normalized == valid {
				found = true
				break
			}
		}
		if !found {
			return errors.New("device_type must be one of: mobile, tablet, desktop")
		}
	}

	return nil
}

func (dtp *DeviceTypeProcessor) MatchesRule(requestValue string, rule TargetingRule) bool {
	normalizedRequest := dtp.NormalizeValue(requestValue)

	for _, ruleValue := range rule.Values {
		normalizedRule := dtp.NormalizeValue(ruleValue)
		if normalizedRequest == normalizedRule {
			return true
		}
	}

	return false
}

// AgeGroupProcessor handles age-based targeting
type AgeGroupProcessor struct{}

func NewAgeGroupProcessor() DimensionProcessor {
	return &AgeGroupProcessor{}
}

func (agp *AgeGroupProcessor) GetName() string {
	return "age_group"
}

func (agp *AgeGroupProcessor) GetValue(req DeliveryRequest) string {
	// This would need to be added to DeliveryRequest struct
	return ""
}

func (agp *AgeGroupProcessor) NormalizeValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (agp *AgeGroupProcessor) ValidateRule(rule TargetingRule) error {
	if len(rule.Values) == 0 {
		return errors.New("age_group rule must have at least one value")
	}

	validGroups := []string{"13-17", "18-24", "25-34", "35-44", "45-54", "55-64", "65+"}
	for _, value := range rule.Values {
		normalized := agp.NormalizeValue(value)
		found := false
		for _, valid := range validGroups {
			if normalized == valid {
				found = true
				break
			}
		}
		if !found {
			return errors.New("age_group must be one of the predefined age ranges")
		}
	}

	return nil
}

func (agp *AgeGroupProcessor) MatchesRule(requestValue string, rule TargetingRule) bool {
	normalizedRequest := agp.NormalizeValue(requestValue)

	for _, ruleValue := range rule.Values {
		normalizedRule := agp.NormalizeValue(ruleValue)
		if normalizedRequest == normalizedRule {
			return true
		}
	}

	return false
}

// TimeOfDayProcessor handles time-based targeting
type TimeOfDayProcessor struct{}

func NewTimeOfDayProcessor() DimensionProcessor {
	return &TimeOfDayProcessor{}
}

func (todp *TimeOfDayProcessor) GetName() string {
	return "time_of_day"
}

func (todp *TimeOfDayProcessor) GetValue(req DeliveryRequest) string {
	// Get current hour (0-23)
	return strconv.Itoa(time.Now().Hour())
}

func (todp *TimeOfDayProcessor) NormalizeValue(value string) string {
	return strings.TrimSpace(value)
}

func (todp *TimeOfDayProcessor) ValidateRule(rule TargetingRule) error {
	if len(rule.Values) == 0 {
		return errors.New("time_of_day rule must have at least one value")
	}

	for _, value := range rule.Values {
		// Support hour ranges like "9-17" or individual hours like "14"
		value = strings.TrimSpace(value)
		if strings.Contains(value, "-") {
			// Range validation
			parts := strings.Split(value, "-")
			if len(parts) != 2 {
				return errors.New("time range must be in format 'start-end'")
			}

			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])

			if err1 != nil || err2 != nil {
				return errors.New("time range values must be integers")
			}

			if start < 0 || start > 23 || end < 0 || end > 23 {
				return errors.New("hour values must be between 0 and 23")
			}
		} else {
			// Single hour validation
			hour, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("hour must be an integer")
			}
			if hour < 0 || hour > 23 {
				return errors.New("hour must be between 0 and 23")
			}
		}
	}

	return nil
}

func (todp *TimeOfDayProcessor) MatchesRule(requestValue string, rule TargetingRule) bool {
	currentHour, err := strconv.Atoi(requestValue)
	if err != nil {
		return false
	}

	for _, ruleValue := range rule.Values {
		ruleValue = strings.TrimSpace(ruleValue)

		if strings.Contains(ruleValue, "-") {
			// Range matching
			parts := strings.Split(ruleValue, "-")
			if len(parts) == 2 {
				start, err1 := strconv.Atoi(parts[0])
				end, err2 := strconv.Atoi(parts[1])

				if err1 == nil && err2 == nil {
					if currentHour >= start && currentHour <= end {
						return true
					}
				}
			}
		} else {
			// Exact hour matching
			ruleHour, err := strconv.Atoi(ruleValue)
			if err == nil && currentHour == ruleHour {
				return true
			}
		}
	}

	return false
}
