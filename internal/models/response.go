package models

// ErrorResponse represents error response format
type ErrorResponse struct {
	Error string `json:"error"`
}

// DeliveryResponse represents the delivery API response
type DeliveryResponse []CampaignResponse

// NewErrorResponse creates a new error response
func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{Error: message}
}

// IsEmpty checks if delivery response is empty
func (dr DeliveryResponse) IsEmpty() bool {
	return len(dr) == 0
}

// FromCampaigns converts campaigns to delivery response
func FromCampaigns(campaigns []Campaign) DeliveryResponse {
	response := make(DeliveryResponse, len(campaigns))
	for i, campaign := range campaigns {
		response[i] = campaign.ToResponse()
	}
	return response
}
