package repository

import (
	"context"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/metrics"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
)

// InstrumentedRepository wraps a repository with metrics collection
type InstrumentedRepository struct {
	next    service.CampaignRepository
	metrics *metrics.Metrics
}

// NewInstrumentedRepository creates a new instrumented repository
func NewInstrumentedRepository(repo service.CampaignRepository, metrics *metrics.Metrics) service.CampaignRepository {
	return &InstrumentedRepository{
		next:    repo,
		metrics: metrics,
	}
}

// GetActiveCampaignsWithRules implements service.CampaignRepository with metrics
func (r *InstrumentedRepository) GetActiveCampaignsWithRules(ctx context.Context) (campaigns []models.CampaignWithRules, err error) {
	defer func(begin time.Time) {
		// Record database query metrics
		r.metrics.RecordDatabaseQuery("select", "campaigns")
		r.metrics.RecordDatabaseQuery("select", "targeting_rules")

		// Record errors if any
		if err != nil {
			r.metrics.RecordDatabaseError("select", "query_error")
		}
	}(time.Now())

	campaigns, err = r.next.GetActiveCampaignsWithRules(ctx)
	return
}
