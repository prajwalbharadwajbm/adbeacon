# Extensible Dimension System

The ad delivery system supports an extensible dimension system that allows you to add new targeting dimensions without modifying core code. This addresses the scalability concern about increasing code complexity as new dimensions are added.

## Overview

The extensible dimension system uses a plugin-like architecture where:
- Each dimension is handled by a `DimensionProcessor` 
- Processors are registered in a `DimensionRegistry`
- Campaign matching uses the registered processors dynamically
- New dimensions can be added at runtime without code changes

## Key Components

### DimensionProcessor Interface

```go
type DimensionProcessor interface {
    GetName() string                                      // Dimension name
    GetValue(req DeliveryRequest) string                 // Extract value from request
    NormalizeValue(value string) string                  // Normalize for comparison
    ValidateRule(rule TargetingRule) error              // Validate targeting rule
    MatchesRule(requestValue string, rule TargetingRule) bool // Check if values match
}
```

### DimensionRegistry

Manages all available dimension processors:

```go
registry := models.NewDimensionRegistry()              // Creates with built-in processors
registry.RegisterProcessor(customProcessor)           // Add custom processor
processor, exists := registry.GetProcessor("country") // Get processor by name
dimensions := registry.ListDimensions()               // List all dimensions
```

### CampaignMatcher

Uses the extensible system for matching:

```go
matcher := models.NewCampaignMatcher(registry)
matches := matcher.MatchesRequest(campaign, request)
```

## Built-in Dimensions

The system comes with three built-in dimension processors:

### 1. Country Processor
- **Name**: `country`
- **Values**: 2-letter country codes (e.g., "us", "ca")
- **Normalization**: Lowercase
- **Validation**: Minimum 2 characters

### 2. OS Processor
- **Name**: `os`
- **Values**: Operating system names (e.g., "android", "ios")
- **Normalization**: Lowercase
- **Validation**: Known OS values (with fallback for unknown)

### 3. App Processor
- **Name**: `app`
- **Values**: Application bundle IDs (e.g., "com.example.app")
- **Normalization**: Trim whitespace only (case-sensitive)
- **Validation**: Must contain dots (package naming convention)

## Adding Custom Dimensions

### Step 1: Implement DimensionProcessor

```go
type BrowserProcessor struct{}

func (bp *BrowserProcessor) GetName() string {
    return "browser"
}

func (bp *BrowserProcessor) GetValue(req models.DeliveryRequest) string {
    // Extract browser from request (e.g., User-Agent header)
    return extractBrowserFromRequest(req)
}

func (bp *BrowserProcessor) NormalizeValue(value string) string {
    return strings.ToLower(strings.TrimSpace(value))
}

func (bp *BrowserProcessor) ValidateRule(rule models.TargetingRule) error {
    validBrowsers := []string{"chrome", "firefox", "safari", "edge"}
    for _, value := range rule.Values {
        if !contains(validBrowsers, bp.NormalizeValue(value)) {
            return fmt.Errorf("unsupported browser: %s", value)
        }
    }
    return nil
}

func (bp *BrowserProcessor) MatchesRule(requestValue string, rule models.TargetingRule) bool {
    normalized := bp.NormalizeValue(requestValue)
    for _, ruleValue := range rule.Values {
        if normalized == bp.NormalizeValue(ruleValue) {
            return true
        }
    }
    return false
}
```

### Step 2: Register the Processor

```go
// Register globally
browserProcessor := &BrowserProcessor{}
models.RegisterCustomDimension(browserProcessor)

// Or register with specific service
deliveryService.RegisterCustomDimension(browserProcessor)
```

### Step 3: Create Targeting Rules

```go
rule := models.TargetingRule{
    Dimension: models.CreateCustomTargetDimension("browser"),
    RuleType:  models.RuleTypeInclude,
    Values:    []string{"chrome", "firefox"},
}
```

## Example Custom Dimensions

### Device Type Targeting

```go
type DeviceTypeProcessor struct{}

func (dtp *DeviceTypeProcessor) GetName() string { return "device_type" }
func (dtp *DeviceTypeProcessor) GetValue(req models.DeliveryRequest) string {
    // Determine device type from User-Agent or request headers
    return detectDeviceType(req)
}
// ... implement other methods
```

### Time-based Targeting

```go
type TimeOfDayProcessor struct{}

func (todp *TimeOfDayProcessor) GetName() string { return "time_of_day" }
func (todp *TimeOfDayProcessor) GetValue(req models.DeliveryRequest) string {
    return strconv.Itoa(time.Now().Hour()) // Current hour (0-23)
}
// Supports ranges like "9-17" or single hours like "14"
```

### Age Group Targeting

```go
type AgeGroupProcessor struct{}

func (agp *AgeGroupProcessor) GetName() string { return "age_group" }
func (agp *AgeGroupProcessor) GetValue(req models.DeliveryRequest) string {
    // Extract age group from user profile or demographics API
    return getUserAgeGroup(req.UserID)
}
// Supports predefined age ranges: "18-24", "25-34", etc.
```

## Extending the DeliveryRequest

To support new dimensions, you may need to extend the `DeliveryRequest` struct:

```go
type DeliveryRequest struct {
    Country     string            `json:"country"`
    OS          string            `json:"os"`
    App         string            `json:"app"`
    
    // Extended fields for custom dimensions
    Browser     string            `json:"browser,omitempty"`
    DeviceType  string            `json:"device_type,omitempty"`
    UserAgent   string            `json:"user_agent,omitempty"`
    Headers     map[string]string `json:"headers,omitempty"`
}
```

## Cache Integration

The extensible system automatically integrates with the caching layer:

```go
// Cache keys are generated for each dimension
key := matcher.BuildIndexKey("browser", "chrome")
// Result: "index:browser:chrome"

// Multiple dimensions create multiple cache keys
keys := service.buildCacheKeys(request)
// Result: ["index:country:us", "index:os:android", "index:browser:chrome"]
```
