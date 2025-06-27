package middleware

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
)

// loggingMiddleware implements logging middleware for DeliveryService
type loggingMiddleware struct {
	logger log.Logger
	next   service.DeliveryService
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger log.Logger) func(service.DeliveryService) service.DeliveryService {
	return func(next service.DeliveryService) service.DeliveryService {
		return &loggingMiddleware{
			logger: logger,
			next:   next,
		}
	}
}

// GetCampaigns implements service.DeliveryService with logging
func (mw *loggingMiddleware) GetCampaigns(ctx context.Context, req models.DeliveryRequest) (campaigns []models.CampaignResponse, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "GetCampaigns",
			"app", req.App,
			"country", req.Country,
			"os", req.OS,
			"campaigns_count", len(campaigns),
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return mw.next.GetCampaigns(ctx, req)
}
