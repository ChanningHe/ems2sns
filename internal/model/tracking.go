package model

import (
	"strings"
	"time"
)

type TrackingSource string

const (
	SourceJapanPostJA TrackingSource = "JP"
	SourceJapanPostEN TrackingSource = "EN"
	SourceChinaEMS    TrackingSource = "CN"
)

type TrackingInfo struct {
	TrackingNumber string
	Status         string
	Details        []TrackingDetail
	LastUpdate     time.Time
	Source         TrackingSource
}

type TrackingDetail struct {
	DateTime    string
	Description string
	Details     string
	Office      string
	Region      string
	Source      TrackingSource
}

// SourceSegment is the per-source slice that makes up a TrackingUpdate.
type SourceSegment struct {
	Source     TrackingSource
	Status     string
	LastUpdate time.Time
	Details    []TrackingDetail
	Delivered  bool
}

// TrackingUpdate is what notifiers render. It carries one segment per source;
// for scheduled pushes only changed sources are included, for /check all
// successful sources are included.
type TrackingUpdate struct {
	TrackingNumber string
	Segments       []SourceSegment
	Delivered      bool // any segment marked delivered
	Initial        bool // true for the first fetch after a new subscription
}

func NewSegment(info *TrackingInfo, delivered bool) SourceSegment {
	return SourceSegment{
		Source:     info.Source,
		Status:     info.Status,
		LastUpdate: info.LastUpdate,
		Details:    info.Details,
		Delivered:  delivered,
	}
}

func SourceFlag(src TrackingSource) string {
	switch src {
	case SourceChinaEMS:
		return "🇨🇳"
	case SourceJapanPostJA:
		return "🇯🇵"
	case SourceJapanPostEN:
		return "🇬🇧"
	default:
		return "📦"
	}
}

func IsChinaTrackingNumber(trackingNumber string) bool {
	return len(trackingNumber) >= 2 && strings.HasSuffix(strings.ToUpper(trackingNumber), "CN")
}

// ParseFlexibleTime tries multiple datetime formats
func ParseFlexibleTime(s string) time.Time {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04",
		"2006/01/02",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
