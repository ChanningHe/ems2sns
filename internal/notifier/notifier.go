package notifier

import (
	"context"

	"github.com/channinghe/ems2sns/internal/model"
)

// Notifier represents a chat platform that can receive tracking updates
type Notifier interface {
	Platform() string
	Start(ctx context.Context) error
	Stop() error
	SendUpdate(sub *model.Subscription, info *model.TrackingInfo, delivered bool) error
}
