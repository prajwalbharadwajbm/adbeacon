package models

import "fmt"

// DependentDimensionProcessor extends DimensionProcessor for dimensions that depend on others
type DependentDimensionProcessor interface {
	DimensionProcessor

	// We need this for getting the country value
	GetDependencies() []string

	// validates the rule considering dependencies(country)
	ValidateWithDependencies(rule TargetingRule, request DeliveryRequest) error

	// checks if a request matches the rule considering dependencies
	MatchesRuleWithDependencies(rule TargetingRule, request DeliveryRequest) bool
}

type DependencyValidator struct {
	registry *DimensionRegistry
}

// NewDependencyValidator creates a new dependency validator
func NewDependencyValidator(registry *DimensionRegistry) *DependencyValidator {
	return &DependencyValidator{registry: registry}
}

// validates a rule considering its dependencies
func (dv *DependencyValidator) ValidateRuleWithDependencies(rule TargetingRule, allRules []TargetingRule) error {
	processor, exists := dv.registry.GetProcessor(string(rule.Dimension))
	if !exists {
		return fmt.Errorf("unknown dimension: %s", rule.Dimension)
	}

	// Check if this is a dependent dimension
	if depProcessor, ok := processor.(DependentDimensionProcessor); ok {
		return dv.validateDependentRule(rule, allRules, depProcessor)
	}

	// For regular dimensions, use standard validation
	return processor.ValidateRule(rule)
}

// validateDependentRule validates a rule that depends on other dimensions
func (dv *DependencyValidator) validateDependentRule(rule TargetingRule, allRules []TargetingRule, processor DependentDimensionProcessor) error {
	// First, validate the rule itself
	if err := processor.ValidateRule(rule); err != nil {
		return err
	}

	// Get the dependencies
	dependencies := processor.GetDependencies()

	// Check if all dependencies are present in the rule set
	presentDependencies := make(map[string][]TargetingRule)
	for _, r := range allRules {
		presentDependencies[string(r.Dimension)] = append(presentDependencies[string(r.Dimension)], r)
	}

	// Validate each dependency
	for _, dep := range dependencies {
		if _, exists := presentDependencies[dep]; !exists {
			return fmt.Errorf("dimension %s depends on %s but %s rules are not present", rule.Dimension, dep, dep)
		}
	}

	return nil
}
