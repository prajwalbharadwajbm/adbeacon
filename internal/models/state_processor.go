package models

import (
	"errors"
	"fmt"
	"strings"
)

// StateProcessor handles state-based targeting that depends on country
type StateProcessor struct {
	countryStates map[string][]string // Maps country codes to valid states
}

// NewStateProcessor creates a new state processor
func NewStateProcessor() DimensionProcessor {
	return &StateProcessor{
		countryStates: getCountryStatesMapping(),
	}
}

// GetName returns the dimension name
func (sp *StateProcessor) GetName() string {
	return "state"
}

// GetValue extracts the state value from the request
func (sp *StateProcessor) GetValue(req DeliveryRequest) string {
	return req.State
}

// NormalizeValue normalizes state codes (lowercase, trimmed)
func (sp *StateProcessor) NormalizeValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// GetDependencies returns the dimensions this processor depends on
func (sp *StateProcessor) GetDependencies() []string {
	return []string{"country"}
}

// ValidateRule validates a state targeting rule
func (sp *StateProcessor) ValidateRule(rule TargetingRule) error {
	if len(rule.Values) == 0 {
		return errors.New("state rule must have at least one value")
	}

	// Basic validation - check if state codes are not empty
	for _, value := range rule.Values {
		if len(strings.TrimSpace(value)) < 2 {
			return errors.New("state code must be at least 2 characters")
		}
	}

	return nil
}

// ValidateWithDependencies validates the state rule considering the country context
func (sp *StateProcessor) ValidateWithDependencies(rule TargetingRule, request DeliveryRequest) error {
	// First do basic validation
	if err := sp.ValidateRule(rule); err != nil {
		return err
	}

	// Get country from request
	country := strings.ToLower(strings.TrimSpace(request.Country))
	if country == "" {
		return errors.New("country is required for state targeting")
	}

	// Get valid states for this country
	validStates, exists := sp.countryStates[country]
	if !exists {
		return fmt.Errorf("country %s does not support state-level targeting", country)
	}

	// Validate each state value
	for _, stateValue := range rule.Values {
		normalizedState := sp.NormalizeValue(stateValue)

		// Check if state is valid for this country
		found := false
		for _, validState := range validStates {
			if normalizedState == strings.ToLower(validState) {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("state %s is not valid for country %s", stateValue, country)
		}
	}

	return nil
}

// MatchesRule checks if a request value matches a targeting rule
func (sp *StateProcessor) MatchesRule(requestValue string, rule TargetingRule) bool {
	normalizedRequest := sp.NormalizeValue(requestValue)

	for _, ruleValue := range rule.Values {
		normalizedRule := sp.NormalizeValue(ruleValue)
		if normalizedRequest == normalizedRule {
			return true
		}
	}

	return false
}

// MatchesRuleWithDependencies checks if a request matches the rule considering dependencies
func (sp *StateProcessor) MatchesRuleWithDependencies(rule TargetingRule, request DeliveryRequest) bool {
	country := strings.ToLower(strings.TrimSpace(request.Country))

	// Does the country belongs to state check
	validStates, exists := sp.countryStates[country]
	if !exists {
		return false
	}

	requestState := sp.NormalizeValue(request.State)
	if requestState == "" {
		return false // No state provided
	}

	// Check if the request state is valid for this country
	stateValidForCountry := false
	for _, validState := range validStates {
		if requestState == strings.ToLower(validState) {
			stateValidForCountry = true
			break
		}
	}

	if !stateValidForCountry {
		return false
	}

	// Now check if the state matches the rule
	return sp.MatchesRule(requestState, rule)
}

// getCountryStatesMapping returns a mapping of country codes to their states/provinces
func getCountryStatesMapping() map[string][]string {
	return map[string][]string{
		"in": {
			"gj", "ma", "ka",
		},
	}
}
