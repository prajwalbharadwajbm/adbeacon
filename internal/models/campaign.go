package models

import (
	"time"
)

// Campaign is advertising package/inventory which advertisers want to run on their platform.
// For example, someone from the Spotify team could
// create a campaign which will consist CTA, image, Status of the campaign
type Campaign struct {
	ID        string         `json:"cid" db:"id"`
	Name      string         `json:"name" db:"name"`
	ImageURL  string         `json:"img" db:"image_url"`
	CTA       string         `json:"cta" db:"cta"`
	Status    CampaignStatus `json:"status" db:"status"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// CampaignStatus represents the status of a campaign
type CampaignStatus string

// enum values for CampaignStatus
const (
	StatusActive   CampaignStatus = "ACTIVE"
	StatusInactive CampaignStatus = "INACTIVE"
)

// IsActive returns true if campaign is active
func (c *Campaign) IsActive() bool {
	return c.Status == StatusActive
}

// CampaignResponse represents the API response format
type CampaignResponse struct {
	CID string `json:"cid"`
	Img string `json:"img"`
	CTA string `json:"cta"`
}

// ToResponse converts Campaign to CampaignResponse
func (c *Campaign) ToResponse() CampaignResponse {
	return CampaignResponse{
		CID: c.ID,
		Img: c.ImageURL,
		CTA: c.CTA,
	}
}
