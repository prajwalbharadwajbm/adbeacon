package middleware

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	reqcontext "github.com/prajwalbharadwajbm/adbeacon/internal/context"
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

// GetCampaigns implements service.DeliveryService with enhanced logging
func (mw *loggingMiddleware) GetCampaigns(ctx context.Context, req models.DeliveryRequest) (campaigns []models.CampaignResponse, err error) {
	defer func(begin time.Time) {
		// Get request context information
		requestID := reqcontext.GetRequestID(ctx)
		userAgent := reqcontext.GetUserAgent(ctx)
		remoteAddr := reqcontext.GetRemoteAddr(ctx)

		// Build log fields
		logFields := []interface{}{
			"method", "GetCampaigns",
			"request_id", requestID,
			"app", req.App,
			"country", req.Country,
			"os", req.OS,
			"campaigns_count", len(campaigns),
			"took", time.Since(begin),
		}

		// Add user agent and remote address if available
		if userAgent != "" {
			logFields = append(logFields, "user_agent", userAgent)
		}
		if remoteAddr != "" {
			logFields = append(logFields, "remote_addr", remoteAddr)
		}

		// Add error information if present
		if err != nil {
			logFields = append(logFields, "error", err.Error())
			logFields = append(logFields, "success", false)
		} else {
			logFields = append(logFields, "error", nil)
			logFields = append(logFields, "success", true)
		}

		// Log the request
		mw.logger.Log(logFields...)
	}(time.Now())

	return mw.next.GetCampaigns(ctx, req)
}
