package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/prajwalbharadwajbm/adbeacon/internal/database"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
)

// PostgresRepository implements service.CampaignRepository using PostgreSQL
type PostgresRepository struct {
	db *database.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *database.DB) service.CampaignRepository {
	return &PostgresRepository{
		db: db,
	}
}

// GetActiveCampaignsWithRules retrieves all active campaigns with their targeting rules
func (r *PostgresRepository) GetActiveCampaignsWithRules(ctx context.Context) ([]models.CampaignWithRules, error) {
	// First, get all active campaigns
	campaignsQuery := `
		SELECT id, name, image_url, cta, status, created_at, updated_at
		FROM campaigns
		WHERE status = 'ACTIVE'
		ORDER BY updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, campaignsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []models.CampaignWithRules
	campaignIDs := make([]string, 0)

	for rows.Next() {
		var campaignWithRules models.CampaignWithRules
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&campaignWithRules.ID,
			&campaignWithRules.Name,
			&campaignWithRules.ImageURL,
			&campaignWithRules.CTA,
			&campaignWithRules.Status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}

		campaignWithRules.CreatedAt = createdAt
		campaignWithRules.UpdatedAt = updatedAt
		campaignWithRules.Rules = []models.TargetingRule{} // Initialize empty rules slice

		campaigns = append(campaigns, campaignWithRules)
		campaignIDs = append(campaignIDs, campaignWithRules.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over campaign rows: %w", err)
	}

	// If no campaigns found, return empty slice
	if len(campaigns) == 0 {
		return campaigns, nil
	}

	// Get targeting rules for all campaigns
	rulesQuery := `
		SELECT campaign_id, dimension, rule_type, values
		FROM targeting_rules
		WHERE campaign_id = ANY($1)
		ORDER BY campaign_id, id
	`

	rulesRows, err := r.db.QueryContext(ctx, rulesQuery, pq.Array(campaignIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query targeting rules: %w", err)
	}
	defer rulesRows.Close()

	// Group rules by campaign ID
	rulesByCampaign := make(map[string][]models.TargetingRule)
	for rulesRows.Next() {
		var rule models.TargetingRule
		var campaignID string

		err := rulesRows.Scan(
			&campaignID,
			&rule.Dimension,
			&rule.RuleType,
			pq.Array(&rule.Values),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan targeting rule: %w", err)
		}

		rule.CampaignID = campaignID
		rulesByCampaign[campaignID] = append(rulesByCampaign[campaignID], rule)
	}

	if err := rulesRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over targeting rules: %w", err)
	}

	// Assign rules to campaigns
	for i := range campaigns {
		if rules, exists := rulesByCampaign[campaigns[i].ID]; exists {
			campaigns[i].Rules = rules
		}
	}

	return campaigns, nil
}
