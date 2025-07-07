package middleware

import (
	"context"

	"github.com/prajwalbharadwajbm/adbeacon/internal/metrics"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
)

// serviceMetricsMiddleware implements metrics collection for DeliveryService
type serviceMetricsMiddleware struct {
	metrics *metrics.CachedMetrics
	next    service.CampaignDeliveryService
}

// NewServiceMetricsMiddleware creates a new service metrics middleware
func NewServiceMetricsMiddleware(metrics *metrics.CachedMetrics) func(service.CampaignDeliveryService) service.CampaignDeliveryService {
	return func(next service.CampaignDeliveryService) service.CampaignDeliveryService {
		return &serviceMetricsMiddleware{
			metrics: metrics,
			next:    next,
		}
	}
}

// GetCampaigns implements service.DeliveryService with business metrics
func (mw *serviceMetricsMiddleware) GetCampaigns(ctx context.Context, req models.DeliveryRequest) (campaigns []models.CampaignResponse, err error) {
	// Call the next service
	campaigns, err = mw.next.GetCampaigns(ctx, req)

	// Record business metrics
	if err == nil {
		// Record successful campaign delivery
		mw.metrics.RecordCampaignDelivery(req.App, req.Country, req.OS, len(campaigns))
	}

	return campaigns, err
}
