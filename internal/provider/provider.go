package provider

import (
	"context"

	"github.com/channinghe/ems2sns/internal/model"
)

// Provider fetches tracking information from an external source
type Provider interface {
	Name() string
	Source() model.TrackingSource
	FetchTrackingInfo(ctx context.Context, trackingNumber string) (*model.TrackingInfo, error)

	// NeedsRegistration returns true if tracking numbers must be registered
	// before querying (e.g. 17track API)
	NeedsRegistration() bool
	Register(ctx context.Context, trackingNumber string) error
}
