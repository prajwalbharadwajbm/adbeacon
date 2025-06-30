package repository

import (
	"context"
	"testing"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewMockRepository(t *testing.T) {
	repo := NewMockRepository()

	assert.NotNil(t, repo)
	assert.IsType(t, &mockRepository{}, repo)
}

func TestMockRepository_GetActiveCampaignsWithRules(t *testing.T) {
	repo := NewMockRepository()

	campaigns, err := repo.GetActiveCampaignsWithRules(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, campaigns)

	// Verify all returned campaigns are active
	for _, campaign := range campaigns {
		assert.Equal(t, models.StatusActive, campaign.Campaign.Status)
		assert.True(t, campaign.IsActive())
	}

	// Verify we have the expected test campaigns
	expectedCampaigns := map[string]bool{
		"spotify":      false,
		"duolingo":     false,
		"subwaysurfer": false,
	}

	for _, campaign := range campaigns {
		if _, exists := expectedCampaigns[campaign.Campaign.ID]; exists {
			expectedCampaigns[campaign.Campaign.ID] = true
		}
	}

	// All expected campaigns should be found
	for id, found := range expectedCampaigns {
		assert.True(t, found, "Campaign %s not found", id)
	}
}

func TestMockRepository_CampaignDataIntegrity(t *testing.T) {
	repo := NewMockRepository()

	campaigns, err := repo.GetActiveCampaignsWithRules(context.Background())
	assert.NoError(t, err)

	for _, campaign := range campaigns {
		// Verify campaign has required fields
		assert.NotEmpty(t, campaign.Campaign.ID)
		assert.NotEmpty(t, campaign.Campaign.Name)
		assert.NotEmpty(t, campaign.Campaign.ImageURL)
		assert.NotEmpty(t, campaign.Campaign.CTA)
		assert.Equal(t, models.StatusActive, campaign.Campaign.Status)
		assert.False(t, campaign.Campaign.CreatedAt.IsZero())
		assert.False(t, campaign.Campaign.UpdatedAt.IsZero())

		// Verify targeting rules
		assert.NotEmpty(t, campaign.Rules, "Campaign %s should have targeting rules", campaign.Campaign.ID)

		for _, rule := range campaign.Rules {
			assert.Equal(t, campaign.Campaign.ID, rule.CampaignID)
			assert.NotEmpty(t, rule.Values)

			// Verify dimension is valid
			validDimensions := []models.TargetDimension{
				models.DimensionCountry,
				models.DimensionOS,
				models.DimensionApp,
			}
			assert.Contains(t, validDimensions, rule.Dimension)

			// Verify rule type is valid
			validRuleTypes := []models.RuleType{
				models.RuleTypeInclude,
				models.RuleTypeExclude,
			}
			assert.Contains(t, validRuleTypes, rule.RuleType)
		}
	}
}

func TestMockRepository_SpecificCampaignRules(t *testing.T) {
	repo := NewMockRepository()

	campaigns, err := repo.GetActiveCampaignsWithRules(context.Background())
	assert.NoError(t, err)

	campaignMap := make(map[string]models.CampaignWithRules)
	for _, campaign := range campaigns {
		campaignMap[campaign.Campaign.ID] = campaign
	}

	// Test Spotify campaign
	spotify, exists := campaignMap["spotify"]
	assert.True(t, exists)
	assert.Len(t, spotify.Rules, 1)
	assert.Equal(t, models.DimensionCountry, spotify.Rules[0].Dimension)
	assert.Equal(t, models.RuleTypeInclude, spotify.Rules[0].RuleType)
	assert.Contains(t, spotify.Rules[0].Values, "US")
	assert.Contains(t, spotify.Rules[0].Values, "Canada")

	// Test Duolingo campaign
	duolingo, exists := campaignMap["duolingo"]
	assert.True(t, exists)
	assert.Len(t, duolingo.Rules, 2)

	// Find OS rule
	var osRule, countryRule *models.TargetingRule
	for i := range duolingo.Rules {
		if duolingo.Rules[i].Dimension == models.DimensionOS {
			osRule = &duolingo.Rules[i]
		} else if duolingo.Rules[i].Dimension == models.DimensionCountry {
			countryRule = &duolingo.Rules[i]
		}
	}

	assert.NotNil(t, osRule)
	assert.Equal(t, models.RuleTypeInclude, osRule.RuleType)
	assert.Contains(t, osRule.Values, "Android")
	assert.Contains(t, osRule.Values, "iOS")

	assert.NotNil(t, countryRule)
	assert.Equal(t, models.RuleTypeExclude, countryRule.RuleType)
	assert.Contains(t, countryRule.Values, "US")

	// Test Subway Surfer campaign
	subwaysurfer, exists := campaignMap["subwaysurfer"]
	assert.True(t, exists)
	assert.Len(t, subwaysurfer.Rules, 2)

	// Find OS and App rules
	var appRule *models.TargetingRule
	osRule = nil
	for i := range subwaysurfer.Rules {
		if subwaysurfer.Rules[i].Dimension == models.DimensionOS {
			osRule = &subwaysurfer.Rules[i]
		} else if subwaysurfer.Rules[i].Dimension == models.DimensionApp {
			appRule = &subwaysurfer.Rules[i]
		}
	}

	assert.NotNil(t, osRule)
	assert.Equal(t, models.RuleTypeInclude, osRule.RuleType)
	assert.Contains(t, osRule.Values, "Android")

	assert.NotNil(t, appRule)
	assert.Equal(t, models.RuleTypeInclude, appRule.RuleType)
	assert.Contains(t, appRule.Values, "com.gametion.ludokinggame")
}

func TestMockRepository_ContextHandling(t *testing.T) {
	repo := NewMockRepository()

	// Test with different context values
	ctx := context.Background()
	campaigns1, err1 := repo.GetActiveCampaignsWithRules(ctx)

	ctxWithValue := context.WithValue(context.Background(), "test", "value")
	campaigns2, err2 := repo.GetActiveCampaignsWithRules(ctxWithValue)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, campaigns1, campaigns2) // Should return same data regardless of context
}
