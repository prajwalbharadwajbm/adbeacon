package repository

import (
	"context"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
)

// mockRepository implements service.CampaignRepository for testing
type mockRepository struct {
	campaigns []models.CampaignWithRules
}

// NewMockRepository creates a new mock repository with sample data
func NewMockRepository() service.CampaignRepository {
	now := time.Now()

	// Sample campaigns with targeting rules
	campaigns := []models.CampaignWithRules{
		{
			Campaign: models.Campaign{
				ID:        "spotify",
				Name:      "Spotify - Music for everyone",
				ImageURL:  "https://somelink",
				CTA:       "Download",
				Status:    models.StatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Rules: []models.TargetingRule{
				{
					ID:         1,
					CampaignID: "spotify",
					Dimension:  models.DimensionCountry,
					RuleType:   models.RuleTypeInclude,
					Values:     []string{"US", "Canada"},
					CreatedAt:  now,
				},
			},
		},
		{
			Campaign: models.Campaign{
				ID:        "duolingo",
				Name:      "Duolingo: Best way to learn",
				ImageURL:  "https://somelink2",
				CTA:       "Install",
				Status:    models.StatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Rules: []models.TargetingRule{
				{
					ID:         2,
					CampaignID: "duolingo",
					Dimension:  models.DimensionOS,
					RuleType:   models.RuleTypeInclude,
					Values:     []string{"Android", "iOS"},
					CreatedAt:  now,
				},
				{
					ID:         3,
					CampaignID: "duolingo",
					Dimension:  models.DimensionCountry,
					RuleType:   models.RuleTypeExclude,
					Values:     []string{"US"},
					CreatedAt:  now,
				},
			},
		},
		{
			Campaign: models.Campaign{
				ID:        "subwaysurfer",
				Name:      "Subway Surfer",
				ImageURL:  "https://somelink3",
				CTA:       "Play",
				Status:    models.StatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Rules: []models.TargetingRule{
				{
					ID:         4,
					CampaignID: "subwaysurfer",
					Dimension:  models.DimensionOS,
					RuleType:   models.RuleTypeInclude,
					Values:     []string{"Android"},
					CreatedAt:  now,
				},
				{
					ID:         5,
					CampaignID: "subwaysurfer",
					Dimension:  models.DimensionApp,
					RuleType:   models.RuleTypeInclude,
					Values:     []string{"com.gametion.ludokinggame"},
					CreatedAt:  now,
				},
			},
		},
	}

	return &mockRepository{
		campaigns: campaigns,
	}
}

// GetActiveCampaignsWithRules returns all active campaigns with their targeting rules
func (r *mockRepository) GetActiveCampaignsWithRules(ctx context.Context) ([]models.CampaignWithRules, error) {
	var activeCampaigns []models.CampaignWithRules

	for _, campaign := range r.campaigns {
		if campaign.IsActive() {
			activeCampaigns = append(activeCampaigns, campaign)
		}
	}

	return activeCampaigns, nil
}
