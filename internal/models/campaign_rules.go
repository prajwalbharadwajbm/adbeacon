package models

// CampaignWithRules represents a campaign with its targeting rules
type CampaignWithRules struct {
	Campaign
	Rules []TargetingRule `json:"rules,omitempty"`
}

// MatchesRequest checks if this campaign matches the delivery request
func (cwr *CampaignWithRules) MatchesRequest(req DeliveryRequest) bool {
	// Only active campaigns can match
	if !cwr.IsActive() {
		return false
	}

	// If no rules exist, campaign matches everyone
	if len(cwr.Rules) == 0 {
		return true
	}

	// Group rules by dimension
	rulesByDimension := make(map[TargetDimension][]TargetingRule)
	for _, rule := range cwr.Rules {
		rulesByDimension[rule.Dimension] = append(rulesByDimension[rule.Dimension], rule)
	}

	// Check each dimension
	for _, rules := range rulesByDimension {
		if !dimensionMatches(req, rules) {
			return false
		}
	}

	return true
}

// dimensionMatches checks if request matches rules for a specific dimension
func dimensionMatches(req DeliveryRequest, rules []TargetingRule) bool {
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

	// If there are include rules, request must match at least one
	if len(includeRules) > 0 {
		matched := false
		for _, rule := range includeRules {
			if req.MatchesRule(rule) {
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
		if req.MatchesRule(rule) {
			return false
		}
	}

	return true
}
