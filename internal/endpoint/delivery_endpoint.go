package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
)

// DeliveryEndpoints holds all endpoints for the delivery service
type DeliveryEndpoints struct {
	GetCampaignsEndpoint endpoint.Endpoint
}

// MakeDeliveryEndpoints creates endpoints for delivery service
func MakeDeliveryEndpoints(s service.DeliveryService) DeliveryEndpoints {
	return DeliveryEndpoints{
		GetCampaignsEndpoint: makeGetCampaignsEndpoint(s),
	}
}

// GetCampaignsRequest represents the request for getting campaigns
type GetCampaignsRequest struct {
	DeliveryRequest models.DeliveryRequest
}

// GetCampaignsResponse represents the response for getting campaigns
type GetCampaignsResponse struct {
	Campaigns []models.CampaignResponse `json:"campaigns,omitempty"`
	Err       error                     `json:"error,omitempty"`
}

// Failed implements the endpoint.Failer interface
func (r GetCampaignsResponse) Failed() error {
	return r.Err
}

// makeGetCampaignsEndpoint creates the endpoint for getting campaigns
func makeGetCampaignsEndpoint(s service.DeliveryService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(GetCampaignsRequest)
		campaigns, err := s.GetCampaigns(ctx, req.DeliveryRequest)
		return GetCampaignsResponse{
			Campaigns: campaigns,
			Err:       err,
		}, nil
	}
}

// GetCampaigns is a helper method to call the endpoint
func (e DeliveryEndpoints) GetCampaigns(ctx context.Context, req models.DeliveryRequest) ([]models.CampaignResponse, error) {
	response, err := e.GetCampaignsEndpoint(ctx, GetCampaignsRequest{DeliveryRequest: req})
	if err != nil {
		return nil, err
	}
	resp := response.(GetCampaignsResponse)
	return resp.Campaigns, resp.Err
}
