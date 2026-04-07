package store

import (
	"github.com/channinghe/ems2sns/internal/model"
)

// Store persists subscription data
type Store interface {
	Load() (map[string]*model.Subscription, error)
	Save(subs map[string]*model.Subscription) error
}
