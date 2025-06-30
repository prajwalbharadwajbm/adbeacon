package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCampaignWithRules_IsActive(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		campaign Campaign
		expected bool
	}{
		{
			name: "active campaign",
			campaign: Campaign{
				ID:        "test",
				Status:    StatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
			expected: true,
		},
		{
			name: "inactive campaign",
			campaign: Campaign{
				ID:        "test",
				Status:    StatusInactive,
				CreatedAt: now,
				UpdatedAt: now,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cwr := CampaignWithRules{
				Campaign: tt.campaign,
				Rules:    []TargetingRule{},
			}
			assert.Equal(t, tt.expected, cwr.IsActive())
		})
	}
}

func TestCampaignWithRules_MatchesRequest(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		campaign CampaignWithRules
		request  DeliveryRequest
		expected bool
	}{
		{
			name: "matches include country rule",
			campaign: CampaignWithRules{
				Campaign: Campaign{
					ID:        "spotify",
					Status:    StatusActive,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Rules: []TargetingRule{
					{
						Dimension: DimensionCountry,
						RuleType:  RuleTypeInclude,
						Values:    []string{"us", "canada"},
					},
				},
			},
			request: DeliveryRequest{
				App:     "com.test.app",
				Country: "us",
				OS:      "android",
			},
			expected: true,
		},
		{
			name: "excluded by country rule",
			campaign: CampaignWithRules{
				Campaign: Campaign{
					ID:        "duolingo",
					Status:    StatusActive,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Rules: []TargetingRule{
					{
						Dimension: DimensionOS,
						RuleType:  RuleTypeInclude,
						Values:    []string{"android", "ios"},
					},
					{
						Dimension: DimensionCountry,
						RuleType:  RuleTypeExclude,
						Values:    []string{"us"},
					},
				},
			},
			request: DeliveryRequest{
				App:     "com.test.app",
				Country: "us",
				OS:      "android",
			},
			expected: false,
		},
		{
			name: "matches multiple rules",
			campaign: CampaignWithRules{
				Campaign: Campaign{
					ID:        "subwaysurfer",
					Status:    StatusActive,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Rules: []TargetingRule{
					{
						Dimension: DimensionOS,
						RuleType:  RuleTypeInclude,
						Values:    []string{"android"},
					},
					{
						Dimension: DimensionApp,
						RuleType:  RuleTypeInclude,
						Values:    []string{"com.gametion.ludokinggame"},
					},
				},
			},
			request: DeliveryRequest{
				App:     "com.gametion.ludokinggame",
				Country: "in",
				OS:      "android",
			},
			expected: true,
		},
		{
			name: "fails OS include rule",
			campaign: CampaignWithRules{
				Campaign: Campaign{
					ID:        "test",
					Status:    StatusActive,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Rules: []TargetingRule{
					{
						Dimension: DimensionOS,
						RuleType:  RuleTypeInclude,
						Values:    []string{"ios"},
					},
				},
			},
			request: DeliveryRequest{
				App:     "com.test.app",
				Country: "us",
				OS:      "android",
			},
			expected: false,
		},
		{
			name: "no rules always matches",
			campaign: CampaignWithRules{
				Campaign: Campaign{
					ID:        "test",
					Status:    StatusActive,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Rules: []TargetingRule{},
			},
			request: DeliveryRequest{
				App:     "com.test.app",
				Country: "us",
				OS:      "android",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.campaign.MatchesRequest(tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCampaignWithRules_ToResponse(t *testing.T) {
	now := time.Now()

	campaign := CampaignWithRules{
		Campaign: Campaign{
			ID:        "spotify",
			Name:      "Spotify - Music for everyone",
			ImageURL:  "https://example.com/spotify.jpg",
			CTA:       "Download",
			Status:    StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Rules: []TargetingRule{},
	}

	response := campaign.ToResponse()

	assert.Equal(t, "spotify", response.CID)
	assert.Equal(t, "https://example.com/spotify.jpg", response.Img)
	assert.Equal(t, "Download", response.CTA)
}
