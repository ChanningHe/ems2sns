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

// NotifyFunc is called by the Tracker with the set of source segments that
// changed (or all segments for initial status / manual check).
type NotifyFunc func(sub *model.Subscription, update *model.TrackingUpdate)

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

func (t *Tracker) notify(sub *model.Subscription, update *model.TrackingUpdate) {
	t.notifyMu.RLock()
	defer t.notifyMu.RUnlock()
	for _, fn := range t.notifyFuncs {
		fn(sub, update)
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

// detectChanges fetches all applicable sources and returns the segments that
// changed hash vs the subscription's previously stored per-source hash. It
// mutates sub.LastHashes / DeliveredBySrc / IsDelivered / LastCheck and
// returns (changedSegments, anyDelivered).
func (t *Tracker) detectChanges(ctx context.Context, sub *model.Subscription) ([]model.SourceSegment, bool) {
	results, err := t.merger.FetchAll(ctx, sub.TrackingNumber)
	if err != nil {
		log.Printf("Fetch error for %s: %v", sub.TrackingNumber, err)
		return nil, sub.IsDelivered
	}

	var changed []model.SourceSegment

	t.mu.Lock()
	sub.EnsureMaps()
	for _, r := range results {
		if r.Err != nil || r.Info == nil {
			continue
		}
		h := computeHash(r.Info)
		delivered := isDelivered(r.Info)
		if h != sub.LastHashes[r.Source] {
			sub.LastHashes[r.Source] = h
			sub.DeliveredBySrc[r.Source] = delivered
			changed = append(changed, model.NewSegment(r.Info, delivered))
		} else if sub.DeliveredBySrc[r.Source] != delivered {
			sub.DeliveredBySrc[r.Source] = delivered
		}
	}
	sub.LastCheck = time.Now()
	// Any source reporting delivered marks the whole subscription delivered.
	anyDelivered := false
	for _, d := range sub.DeliveredBySrc {
		if d {
			anyDelivered = true
			break
		}
	}
	sub.IsDelivered = anyDelivered
	t.mu.Unlock()

	return changed, anyDelivered
}

func (t *Tracker) checkSubscription(ctx context.Context, sub *model.Subscription) {
	if sub.IsDelivered {
		return
	}

	changed, anyDelivered := t.detectChanges(ctx, sub)
	if len(changed) == 0 {
		return
	}

	log.Printf("Update detected: %s [%s:%s] sources=%d", sub.TrackingNumber, sub.Platform, sub.ChannelID, len(changed))
	t.saveSubscriptions()
	t.notify(sub, &model.TrackingUpdate{
		TrackingNumber: sub.TrackingNumber,
		Segments:       changed,
		Delivered:      anyDelivered,
	})
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
	sub.EnsureMaps()
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

// Check performs a one-shot query against all applicable providers and returns
// a TrackingUpdate containing one segment per successful source. Intended for
// user-triggered /check commands — no hash comparison is performed.
func (t *Tracker) Check(ctx context.Context, trackingNumber string) (*model.TrackingUpdate, error) {
	results, err := t.merger.FetchAll(ctx, trackingNumber)
	if err != nil {
		return nil, err
	}

	var segments []model.SourceSegment
	anyDelivered := false
	var errs int
	for _, r := range results {
		if r.Err != nil || r.Info == nil {
			errs++
			continue
		}
		delivered := isDelivered(r.Info)
		if delivered {
			anyDelivered = true
		}
		segments = append(segments, model.NewSegment(r.Info, delivered))
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("all providers failed for %s", trackingNumber)
	}

	return &model.TrackingUpdate{
		TrackingNumber: trackingNumber,
		Segments:       segments,
		Delivered:      anyDelivered,
	}, nil
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

		changed, anyDelivered := t.detectChanges(ctx, sub)
		if len(changed) > 0 {
			result.UpdatesFound++
			if anyDelivered {
				result.DeliveredCount++
			}
			t.saveSubscriptions()
			t.notify(sub, &model.TrackingUpdate{
				TrackingNumber: sub.TrackingNumber,
				Segments:       changed,
				Delivered:      anyDelivered,
			})
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

	results, err := t.merger.FetchAll(ctx, sub.TrackingNumber)
	if err != nil {
		log.Printf("Initial status error for %s: %v", sub.TrackingNumber, err)
		return
	}

	var segments []model.SourceSegment
	anyDelivered := false

	t.mu.Lock()
	key := sub.Key()
	stored, ok := t.subscriptions[key]
	if ok {
		stored.EnsureMaps()
	}
	for _, r := range results {
		if r.Err != nil || r.Info == nil {
			continue
		}
		delivered := isDelivered(r.Info)
		if delivered {
			anyDelivered = true
		}
		segments = append(segments, model.NewSegment(r.Info, delivered))
		if ok {
			stored.LastHashes[r.Source] = computeHash(r.Info)
			stored.DeliveredBySrc[r.Source] = delivered
		}
	}
	if ok {
		stored.LastCheck = time.Now()
		stored.IsDelivered = anyDelivered
	}
	t.mu.Unlock()

	if len(segments) == 0 {
		log.Printf("Initial status for %s: all providers failed", sub.TrackingNumber)
		return
	}

	t.saveSubscriptions()
	t.notify(sub, &model.TrackingUpdate{
		TrackingNumber: sub.TrackingNumber,
		Segments:       segments,
		Delivered:      anyDelivered,
		Initial:        true,
	})
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
	for _, s := range subs {
		s.EnsureMaps()
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
