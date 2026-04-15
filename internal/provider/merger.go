package provider

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/channinghe/ems2sns/internal/model"
)

// SourceResult is the per-provider outcome of a fetch.
type SourceResult struct {
	Source model.TrackingSource
	Info   *model.TrackingInfo
	Err    error
}

// Merger queries multiple providers concurrently and returns per-source results.
type Merger struct {
	providers []Provider
}

func NewMerger(providers ...Provider) *Merger {
	return &Merger{providers: providers}
}

// FetchAll queries all applicable providers concurrently and returns one
// SourceResult per provider. Results are sorted by source for deterministic
// ordering downstream. An error is returned only when no provider is applicable
// to the given tracking number (otherwise per-provider errors ride on the
// SourceResult).
func (m *Merger) FetchAll(ctx context.Context, trackingNumber string) ([]SourceResult, error) {
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

	results := make([]SourceResult, len(applicable))
	var wg sync.WaitGroup

	for i, p := range applicable {
		wg.Add(1)
		go func(idx int, prov Provider) {
			defer wg.Done()
			info, err := prov.FetchTrackingInfo(ctx, trackingNumber)
			if err != nil {
				log.Printf("[%s] fetch error for %s: %v", prov.Name(), trackingNumber, err)
			}
			results[idx] = SourceResult{Source: prov.Source(), Info: info, Err: err}
		}(i, p)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Source < results[j].Source
	})
	return results, nil
}
