package tracker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/channinghe/ems2sns/internal/config"
	"github.com/channinghe/ems2sns/internal/model"
	"github.com/channinghe/ems2sns/internal/provider"
	"github.com/channinghe/ems2sns/internal/store"
)

// NotifyFunc is called by the Tracker when a subscription has an update.
// Implementations (Telegram, Discord) send the formatted message.
type NotifyFunc func(sub *model.Subscription, info *model.TrackingInfo, delivered bool)

// PushResult contains statistics from a manual push operation
type PushResult struct {
	TotalChecked   int
	UpdatesFound   int
	DeliveredCount int
}

// Tracker orchestrates polling, change detection, and notification dispatch
type Tracker struct {
	merger        *provider.Merger
	seventeenTrk  *provider.SeventeenTrackProvider
	store         store.Store
	cfg           *config.Config
	subscriptions map[string]*model.Subscription
	mu            sync.RWMutex
	notifyFuncs   []NotifyFunc
	notifyMu      sync.RWMutex
}

func New(
	merger *provider.Merger,
	seventeenTrk *provider.SeventeenTrackProvider,
	st store.Store,
	cfg *config.Config,
) *Tracker {
	return &Tracker{
		merger:        merger,
		seventeenTrk:  seventeenTrk,
		store:         st,
		cfg:           cfg,
		subscriptions: make(map[string]*model.Subscription),
	}
}

// RegisterNotifyFunc adds a callback that is invoked on tracking updates
func (t *Tracker) RegisterNotifyFunc(fn NotifyFunc) {
	t.notifyMu.Lock()
	defer t.notifyMu.Unlock()
	t.notifyFuncs = append(t.notifyFuncs, fn)
}

func (t *Tracker) notify(sub *model.Subscription, info *model.TrackingInfo, delivered bool) {
	t.notifyMu.RLock()
	defer t.notifyMu.RUnlock()
	for _, fn := range t.notifyFuncs {
		fn(sub, info, delivered)
	}
}

// Start begins the polling loop. Blocks until ctx is cancelled.
func (t *Tracker) Start(ctx context.Context) {
	log.Printf("Tracker started (poll_interval=%v)", t.cfg.Tracking.PollInterval)

	if err := t.loadSubscriptions(); err != nil {
		log.Printf("WARNING: failed to load subscriptions: %v", err)
	}

	ticker := time.NewTicker(t.cfg.Tracking.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Tracker stopping")
			return
		case <-ticker.C:
			t.pollAll(ctx)
		}
	}
}

func (t *Tracker) pollAll(ctx context.Context) {
	log.Println("Starting periodic check...")

	subs := t.allActiveSubscriptions()
	for _, sub := range subs {
		select {
		case <-ctx.Done():
			return
		default:
		}
		t.checkSubscription(ctx, sub)
		time.Sleep(t.cfg.Tracking.RequestDelay)
	}

	log.Printf("Periodic check done. Active: %d", len(subs))
}

func (t *Tracker) checkSubscription(ctx context.Context, sub *model.Subscription) {
	if sub.IsDelivered {
		return
	}

	info, err := t.merger.FetchAll(ctx, sub.TrackingNumber)
	if err != nil {
		log.Printf("Fetch error for %s: %v", sub.TrackingNumber, err)
		return
	}

	currentHash := computeHash(info)
	delivered := isDelivered(info)

	if currentHash != sub.LastHash {
		log.Printf("Update detected: %s [%s:%s]", sub.TrackingNumber, sub.Platform, sub.ChannelID)

		t.mu.Lock()
		sub.LastHash = currentHash
		sub.LastCheck = time.Now()
		sub.IsDelivered = delivered
		t.mu.Unlock()

		t.saveSubscriptions()
		t.notify(sub, info, delivered)
	} else {
		t.mu.Lock()
		sub.LastCheck = time.Now()
		t.mu.Unlock()
	}
}

// --- CommandHandler methods (called by notifiers) ---

func (t *Tracker) Subscribe(ctx context.Context, platform, channelID, userID, username, trackingNumber string) error {
	key := model.MakeSubKey(platform, channelID, trackingNumber)

	t.mu.Lock()
	if _, exists := t.subscriptions[key]; exists {
		t.mu.Unlock()
		return fmt.Errorf("already subscribed to %s", trackingNumber)
	}

	sub := &model.Subscription{
		TrackingNumber: trackingNumber,
		Platform:       platform,
		ChannelID:      channelID,
		UserID:         userID,
		Username:       username,
		CreatedAt:      time.Now(),
	}
	t.subscriptions[key] = sub
	t.mu.Unlock()

	t.saveSubscriptions()

	// Register with 17track if needed
	if model.IsChinaTrackingNumber(trackingNumber) && t.seventeenTrk != nil && t.seventeenTrk.IsConfigured() {
		if err := t.seventeenTrk.Register(ctx, trackingNumber); err != nil {
			log.Printf("17track register for %s: %v", trackingNumber, err)
		} else {
			log.Printf("17track registered: %s", trackingNumber)
		}
	}

	// Send initial status asynchronously
	go t.sendInitialStatus(sub)

	return nil
}

func (t *Tracker) Unsubscribe(platform, channelID, trackingNumber string) error {
	key := model.MakeSubKey(platform, channelID, trackingNumber)

	t.mu.Lock()
	if _, exists := t.subscriptions[key]; !exists {
		t.mu.Unlock()
		return fmt.Errorf("subscription not found")
	}
	delete(t.subscriptions, key)
	t.mu.Unlock()

	t.saveSubscriptions()
	return nil
}

func (t *Tracker) List(platform, channelID string) []*model.Subscription {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []*model.Subscription
	for _, sub := range t.subscriptions {
		if sub.Platform == platform && sub.ChannelID == channelID {
			result = append(result, sub)
		}
	}
	return result
}

func (t *Tracker) Check(ctx context.Context, trackingNumber string) (*model.TrackingInfo, error) {
	return t.merger.FetchAll(ctx, trackingNumber)
}

func (t *Tracker) ManualPush(ctx context.Context, platform, channelID string) *PushResult {
	log.Printf("Manual push: %s:%s", platform, channelID)

	result := &PushResult{}
	subs := t.allActiveSubscriptions()
	result.TotalChecked = len(subs)

	for _, sub := range subs {
		if sub.IsDelivered {
			continue
		}

		info, err := t.merger.FetchAll(ctx, sub.TrackingNumber)
		if err != nil {
			log.Printf("Manual push error for %s: %v", sub.TrackingNumber, err)
			continue
		}

		currentHash := computeHash(info)
		delivered := isDelivered(info)

		if currentHash != sub.LastHash {
			result.UpdatesFound++
			if delivered {
				result.DeliveredCount++
			}

			t.mu.Lock()
			sub.LastHash = currentHash
			sub.LastCheck = time.Now()
			sub.IsDelivered = delivered
			t.mu.Unlock()

			t.saveSubscriptions()
			t.notify(sub, info, delivered)
		} else {
			t.mu.Lock()
			sub.LastCheck = time.Now()
			t.mu.Unlock()
		}

		time.Sleep(1 * time.Second)
	}

	log.Printf("Manual push done: checked=%d, updates=%d, delivered=%d",
		result.TotalChecked, result.UpdatesFound, result.DeliveredCount)
	return result
}

func (t *Tracker) sendInitialStatus(sub *model.Subscription) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	info, err := t.merger.FetchAll(ctx, sub.TrackingNumber)
	if err != nil {
		log.Printf("Initial status error for %s: %v", sub.TrackingNumber, err)
		return
	}

	currentHash := computeHash(info)
	delivered := isDelivered(info)

	t.mu.Lock()
	key := sub.Key()
	if s, ok := t.subscriptions[key]; ok {
		s.LastHash = currentHash
		s.LastCheck = time.Now()
		s.IsDelivered = delivered
	}
	t.mu.Unlock()

	t.saveSubscriptions()
	t.notify(sub, info, delivered)
}

// --- internal helpers ---

func (t *Tracker) allActiveSubscriptions() []*model.Subscription {
	t.mu.RLock()
	defer t.mu.RUnlock()

	subs := make([]*model.Subscription, 0, len(t.subscriptions))
	for _, sub := range t.subscriptions {
		subs = append(subs, sub)
	}
	return subs
}

func (t *Tracker) loadSubscriptions() error {
	subs, err := t.store.Load()
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.subscriptions = subs
	t.mu.Unlock()
	log.Printf("Loaded %d subscriptions", len(subs))
	return nil
}

func (t *Tracker) saveSubscriptions() {
	t.mu.RLock()
	subs := make(map[string]*model.Subscription, len(t.subscriptions))
	for k, v := range t.subscriptions {
		subs[k] = v
	}
	t.mu.RUnlock()

	if err := t.store.Save(subs); err != nil {
		log.Printf("Error saving subscriptions: %v", err)
	}
}
