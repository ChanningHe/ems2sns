package model

import (
	"fmt"
	"time"
)

type Subscription struct {
	TrackingNumber string                      `json:"tracking_number"`
	Platform       string                      `json:"platform"`
	ChannelID      string                      `json:"channel_id"`
	UserID         string                      `json:"user_id"`
	Username       string                      `json:"username"`
	CreatedAt      time.Time                   `json:"created_at"`
	LastHashes     map[TrackingSource]string   `json:"last_hashes"`
	DeliveredBySrc map[TrackingSource]bool     `json:"delivered_by_source"`
	LastCheck      time.Time                   `json:"last_check"`
	IsDelivered    bool                        `json:"is_delivered"`
}

// EnsureMaps initializes the per-source state maps if nil.
func (s *Subscription) EnsureMaps() {
	if s.LastHashes == nil {
		s.LastHashes = make(map[TrackingSource]string)
	}
	if s.DeliveredBySrc == nil {
		s.DeliveredBySrc = make(map[TrackingSource]bool)
	}
}

// Key returns a unique identifier for this subscription
func (s *Subscription) Key() string {
	return MakeSubKey(s.Platform, s.ChannelID, s.TrackingNumber)
}

func MakeSubKey(platform, channelID, trackingNumber string) string {
	return fmt.Sprintf("%s:%s:%s", platform, channelID, trackingNumber)
}
