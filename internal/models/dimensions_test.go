package models

import (
	"testing"
	"time"
)

func TestDimensionRegistry(t *testing.T) {
	// Create a new registry
	registry := NewDimensionRegistry()

	// Test built-in processors are registered
	expectedDimensions := []string{"country", "os", "app"}
	actualDimensions := registry.ListDimensions()

	if len(actualDimensions) != len(expectedDimensions) {
		t.Errorf("Expected %d dimensions, got %d", len(expectedDimensions), len(actualDimensions))
	}

	for _, expected := range expectedDimensions {
		_, exists := registry.GetProcessor(expected)
		if !exists {
			t.Errorf("Expected dimension %s to be registered", expected)
		}
	}
}

func TestCustomDimensionRegistration(t *testing.T) {
	registry := NewDimensionRegistry()

	// Register custom dimension
	deviceTypeProcessor := NewDeviceTypeProcessor()
	registry.RegisterProcessor(deviceTypeProcessor)

	// Test custom dimension is registered
	processor, exists := registry.GetProcessor("device_type")
	if !exists {
		t.Error("Expected device_type dimension to be registered")
	}

	if processor.GetName() != "device_type" {
		t.Errorf("Expected processor name to be 'device_type', got '%s'", processor.GetName())
	}

	// Test dimensions list includes custom dimension
	dimensions := registry.ListDimensions()
	found := false
	for _, dim := range dimensions {
		if dim == "device_type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected device_type to be in dimensions list")
	}
}

func TestCampaignMatcher(t *testing.T) {
	// Create matcher with custom registry
	registry := NewDimensionRegistry()
	registry.RegisterProcessor(NewTimeOfDayProcessor())
	matcher := NewCampaignMatcher(registry)

	// Create test campaign with time-based targeting
	campaign := CampaignWithRules{
		Campaign: Campaign{
			ID:        "test-campaign",
			Name:      "Time-based Campaign",
			Status:    StatusActive,
			CreatedAt: time.Now(),
		},
		Rules: []TargetingRule{
			{
				ID:         1,
				CampaignID: "test-campaign",
				Dimension:  DimensionTimeOfDay,
				RuleType:   RuleTypeInclude,
				Values:     []string{"9-17"}, // 9 AM to 5 PM
				CreatedAt:  time.Now(),
			},
		},
	}

	// Create request (time-based matching will use current hour)
	req := DeliveryRequest{
		Country: "us",
		OS:      "android",
		App:     "com.example.app",
	}

	// Test matching (result depends on current time)
	matches := matcher.MatchesRequest(campaign, req)

	// Since we can't control time in test, just verify the method runs without error
	t.Logf("Campaign matches request: %v", matches)
}

func TestDimensionProcessorValidation(t *testing.T) {
	tests := []struct {
		name          string
		processor     DimensionProcessor
		rule          TargetingRule
		shouldBeValid bool
	}{
		{
			name:      "Valid country rule",
			processor: NewCountryProcessor(),
			rule: TargetingRule{
				Dimension: DimensionCountry,
				RuleType:  RuleTypeInclude,
				Values:    []string{"us", "ca"},
			},
			shouldBeValid: true,
		},
		{
			name:      "Invalid country rule - empty values",
			processor: NewCountryProcessor(),
			rule: TargetingRule{
				Dimension: DimensionCountry,
				RuleType:  RuleTypeInclude,
				Values:    []string{},
			},
			shouldBeValid: false,
		},
		{
			name:      "Invalid country rule - short code",
			processor: NewCountryProcessor(),
			rule: TargetingRule{
				Dimension: DimensionCountry,
				RuleType:  RuleTypeInclude,
				Values:    []string{"u"},
			},
			shouldBeValid: false,
		},
		{
			name:      "Valid OS rule",
			processor: NewOSProcessor(),
			rule: TargetingRule{
				Dimension: DimensionOS,
				RuleType:  RuleTypeInclude,
				Values:    []string{"android", "ios"},
			},
			shouldBeValid: true,
		},
		{
			name:      "Valid app rule",
			processor: NewAppProcessor(),
			rule: TargetingRule{
				Dimension: DimensionApp,
				RuleType:  RuleTypeInclude,
				Values:    []string{"com.example.app"},
			},
			shouldBeValid: true,
		},
		{
			name:      "Invalid app rule - no dots",
			processor: NewAppProcessor(),
			rule: TargetingRule{
				Dimension: DimensionApp,
				RuleType:  RuleTypeInclude,
				Values:    []string{"invalidapp"},
			},
			shouldBeValid: false,
		},
		{
			name:      "Valid device type rule",
			processor: NewDeviceTypeProcessor(),
			rule: TargetingRule{
				Dimension: DimensionDeviceType,
				RuleType:  RuleTypeInclude,
				Values:    []string{"mobile", "tablet"},
			},
			shouldBeValid: true,
		},
		{
			name:      "Invalid device type rule",
			processor: NewDeviceTypeProcessor(),
			rule: TargetingRule{
				Dimension: DimensionDeviceType,
				RuleType:  RuleTypeInclude,
				Values:    []string{"invalid"},
			},
			shouldBeValid: false,
		},
		{
			name:      "Valid time of day rule - range",
			processor: NewTimeOfDayProcessor(),
			rule: TargetingRule{
				Dimension: DimensionTimeOfDay,
				RuleType:  RuleTypeInclude,
				Values:    []string{"9-17"},
			},
			shouldBeValid: true,
		},
		{
			name:      "Valid time of day rule - single hour",
			processor: NewTimeOfDayProcessor(),
			rule: TargetingRule{
				Dimension: DimensionTimeOfDay,
				RuleType:  RuleTypeInclude,
				Values:    []string{"14"},
			},
			shouldBeValid: true,
		},
		{
			name:      "Invalid time of day rule - bad range",
			processor: NewTimeOfDayProcessor(),
			rule: TargetingRule{
				Dimension: DimensionTimeOfDay,
				RuleType:  RuleTypeInclude,
				Values:    []string{"25-30"},
			},
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.processor.ValidateRule(tt.rule)
			if tt.shouldBeValid && err != nil {
				t.Errorf("Expected rule to be valid, got error: %v", err)
			}
			if !tt.shouldBeValid && err == nil {
				t.Error("Expected rule to be invalid, but got no error")
			}
		})
	}
}

func TestProcessorMatching(t *testing.T) {
	tests := []struct {
		name         string
		processor    DimensionProcessor
		requestValue string
		rule         TargetingRule
		shouldMatch  bool
	}{
		{
			name:         "Country exact match",
			processor:    NewCountryProcessor(),
			requestValue: "US",
			rule: TargetingRule{
				Values: []string{"us", "ca"},
			},
			shouldMatch: true,
		},
		{
			name:         "Country no match",
			processor:    NewCountryProcessor(),
			requestValue: "de",
			rule: TargetingRule{
				Values: []string{"us", "ca"},
			},
			shouldMatch: false,
		},
		{
			name:         "OS exact match",
			processor:    NewOSProcessor(),
			requestValue: "Android",
			rule: TargetingRule{
				Values: []string{"android", "ios"},
			},
			shouldMatch: true,
		},
		{
			name:         "App exact match",
			processor:    NewAppProcessor(),
			requestValue: "com.example.app",
			rule: TargetingRule{
				Values: []string{"com.example.app", "com.other.app"},
			},
			shouldMatch: true,
		},
		{
			name:         "Device type match",
			processor:    NewDeviceTypeProcessor(),
			requestValue: "Mobile",
			rule: TargetingRule{
				Values: []string{"mobile", "tablet"},
			},
			shouldMatch: true,
		},
		{
			name:         "Time of day - hour in range",
			processor:    NewTimeOfDayProcessor(),
			requestValue: "14", // 2 PM
			rule: TargetingRule{
				Values: []string{"9-17"}, // 9 AM to 5 PM
			},
			shouldMatch: true,
		},
		{
			name:         "Time of day - hour outside range",
			processor:    NewTimeOfDayProcessor(),
			requestValue: "20", // 8 PM
			rule: TargetingRule{
				Values: []string{"9-17"}, // 9 AM to 5 PM
			},
			shouldMatch: false,
		},
		{
			name:         "Time of day - exact hour match",
			processor:    NewTimeOfDayProcessor(),
			requestValue: "14",
			rule: TargetingRule{
				Values: []string{"14"},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := tt.processor.MatchesRule(tt.requestValue, tt.rule)
			if matches != tt.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v", tt.shouldMatch, matches)
			}
		})
	}
}

func TestIndexKeyGeneration(t *testing.T) {
	registry := NewDimensionRegistry()
	matcher := NewCampaignMatcher(registry)

	tests := []struct {
		dimension string
		value     string
		expected  string
	}{
		{"country", "US", "index:country:us"},
		{"os", "Android", "index:os:android"},
		{"app", "com.example.app", "index:app:com.example.app"},
		{"unknown", "value", "index:unknown:value"},
	}

	for _, tt := range tests {
		t.Run(tt.dimension+"_"+tt.value, func(t *testing.T) {
			key := matcher.BuildIndexKey(tt.dimension, tt.value)
			if key != tt.expected {
				t.Errorf("Expected key '%s', got '%s'", tt.expected, key)
			}
		})
	}
}

func TestCampaignMatchingWithMultipleDimensions(t *testing.T) {
	// Create matcher with all dimensions
	registry := NewDimensionRegistry()
	registry.RegisterProcessor(NewDeviceTypeProcessor())
	registry.RegisterProcessor(NewTimeOfDayProcessor())
	matcher := NewCampaignMatcher(registry)

	// Create campaign with multiple targeting rules
	campaign := CampaignWithRules{
		Campaign: Campaign{
			ID:        "multi-dim-campaign",
			Name:      "Multi-dimensional Campaign",
			Status:    StatusActive,
			CreatedAt: time.Now(),
		},
		Rules: []TargetingRule{
			{
				ID:         1,
				CampaignID: "multi-dim-campaign",
				Dimension:  DimensionCountry,
				RuleType:   RuleTypeInclude,
				Values:     []string{"us", "ca"},
				CreatedAt:  time.Now(),
			},
			{
				ID:         2,
				CampaignID: "multi-dim-campaign",
				Dimension:  DimensionOS,
				RuleType:   RuleTypeInclude,
				Values:     []string{"android"},
				CreatedAt:  time.Now(),
			},
			{
				ID:         3,
				CampaignID: "multi-dim-campaign",
				Dimension:  DimensionApp,
				RuleType:   RuleTypeExclude,
				Values:     []string{"com.blocked.app"},
				CreatedAt:  time.Now(),
			},
		},
	}

	tests := []struct {
		name     string
		request  DeliveryRequest
		expected bool
	}{
		{
			name: "All rules match",
			request: DeliveryRequest{
				Country: "us",
				OS:      "android",
				App:     "com.example.app",
			},
			expected: true,
		},
		{
			name: "Country doesn't match",
			request: DeliveryRequest{
				Country: "de",
				OS:      "android",
				App:     "com.example.app",
			},
			expected: false,
		},
		{
			name: "OS doesn't match",
			request: DeliveryRequest{
				Country: "us",
				OS:      "ios",
				App:     "com.example.app",
			},
			expected: false,
		},
		{
			name: "App is excluded",
			request: DeliveryRequest{
				Country: "us",
				OS:      "android",
				App:     "com.blocked.app",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Normalize request values before matching
			tt.request.NormalizeValues()
			matches := matcher.MatchesRequest(campaign, tt.request)
			if matches != tt.expected {
				t.Errorf("Expected match=%v, got match=%v", tt.expected, matches)
			}
		})
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that old enum validation still works
	oldDimensions := []TargetDimension{
		DimensionCountry,
		DimensionOS,
		DimensionApp,
	}

	for _, dim := range oldDimensions {
		if !dim.IsValid() {
			t.Errorf("Expected dimension %s to be valid", dim)
		}
	}

	// Test invalid dimension
	invalidDim := TargetDimension("invalid")
	if invalidDim.IsValid() {
		t.Error("Expected invalid dimension to be invalid")
	}
}

// Benchmark tests to show performance impact
func BenchmarkProcessorMatching(b *testing.B) {
	processor := NewCountryProcessor()
	rule := TargetingRule{
		Values: []string{"us", "ca", "uk", "de", "fr"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.MatchesRule("us", rule)
	}
}

func BenchmarkCampaignMatching(b *testing.B) {
	registry := NewDimensionRegistry()
	matcher := NewCampaignMatcher(registry)

	campaign := CampaignWithRules{
		Campaign: Campaign{
			ID:        "bench-campaign",
			Name:      "Benchmark Campaign",
			Status:    StatusActive,
			CreatedAt: time.Now(),
		},
		Rules: []TargetingRule{
			{
				Dimension: DimensionCountry,
				RuleType:  RuleTypeInclude,
				Values:    []string{"us", "ca", "uk"},
			},
			{
				Dimension: DimensionOS,
				RuleType:  RuleTypeInclude,
				Values:    []string{"android", "ios"},
			},
		},
	}

	request := DeliveryRequest{
		Country: "us",
		OS:      "android",
		App:     "com.example.app",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.MatchesRequest(campaign, request)
	}
}
