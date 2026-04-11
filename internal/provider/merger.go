package provider

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/channinghe/ems2sns/internal/model"
)

// Merger queries multiple providers concurrently and merges results
type Merger struct {
	providers []Provider
}

func NewMerger(providers ...Provider) *Merger {
	return &Merger{providers: providers}
}

// FetchAll queries all applicable providers and returns a merged result.
// For CN tracking numbers, all providers are queried.
// For non-CN numbers, only JapanPost providers are used.
func (m *Merger) FetchAll(ctx context.Context, trackingNumber string) (*model.TrackingInfo, error) {
	isCN := model.IsChinaTrackingNumber(trackingNumber)

	var applicable []Provider
	for _, p := range m.providers {
		if p.Source() == model.SourceChinaEMS && !isCN {
			continue
		}
		if p.Source() == model.SourceChinaEMS {
			if st, ok := p.(*SeventeenTrackProvider); ok && !st.IsConfigured() {
				continue
			}
		}
		applicable = append(applicable, p)
	}

	if len(applicable) == 0 {
		return nil, fmt.Errorf("no applicable providers for %s", trackingNumber)
	}

	type result struct {
		info *model.TrackingInfo
		err  error
		name string
	}

	results := make(chan result, len(applicable))
	var wg sync.WaitGroup

	for _, p := range applicable {
		wg.Add(1)
		go func(prov Provider) {
			defer wg.Done()
			info, err := prov.FetchTrackingInfo(ctx, trackingNumber)
			results <- result{info: info, err: err, name: prov.Name()}
		}(p)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var infos []*model.TrackingInfo
	var errs []string
	for r := range results {
		if r.err != nil {
			log.Printf("[%s] fetch error for %s: %v", r.name, trackingNumber, r.err)
			errs = append(errs, fmt.Sprintf("%s: %v", r.name, r.err))
			continue
		}
		if r.info != nil {
			infos = append(infos, r.info)
		}
	}

	if len(infos) == 0 {
		return nil, fmt.Errorf("all providers failed: %s", strings.Join(errs, "; "))
	}

	return mergeTrackingInfos(infos), nil
}

func mergeTrackingInfos(infos []*model.TrackingInfo) *model.TrackingInfo {
	if len(infos) == 1 {
		return infos[0]
	}

	// Deterministic order: sort by Source so hash stays stable across runs
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Source < infos[j].Source
	})

	merged := &model.TrackingInfo{
		TrackingNumber: infos[0].TrackingNumber,
		Details:        []model.TrackingDetail{},
	}

	var statusParts []string
	var latestUpdate time.Time
	for _, info := range infos {
		if info.Status != "" {
			statusParts = append(statusParts, fmt.Sprintf("%s %s", model.SourceFlag(info.Source), info.Status))
		}
		if info.LastUpdate.After(latestUpdate) {
			latestUpdate = info.LastUpdate
		}
		merged.Details = append(merged.Details, info.Details...)
	}
	merged.Status = strings.Join(statusParts, " | ")
	if latestUpdate.IsZero() {
		latestUpdate = time.Now()
	}
	merged.LastUpdate = latestUpdate

	sortDetailsByTime(merged.Details)
	return merged
}

func sortDetailsByTime(details []model.TrackingDetail) {
	for i := 1; i < len(details); i++ {
		for j := i; j > 0 && compareDateTime(details[j-1].DateTime, details[j].DateTime) > 0; j-- {
			details[j-1], details[j] = details[j], details[j-1]
		}
	}
}

func compareDateTime(a, b string) int {
	ta := model.ParseFlexibleTime(a)
	tb := model.ParseFlexibleTime(b)
	if ta.Before(tb) {
		return -1
	}
	if ta.After(tb) {
		return 1
	}
	return 0
}
