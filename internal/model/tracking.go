package model

import (
	"fmt"
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

func SourceFlag(src TrackingSource) string {
	switch src {
	case SourceChinaEMS:
		return "\U0001F1E8\U0001F1F3"
	case SourceJapanPostJA:
		return "\U0001F1EF\U0001F1F5"
	case SourceJapanPostEN:
		return "\U0001F1EC\U0001F1E7"
	default:
		return "\U0001F4E6"
	}
}

func IsChinaTrackingNumber(trackingNumber string) bool {
	return len(trackingNumber) >= 2 && strings.HasSuffix(strings.ToUpper(trackingNumber), "CN")
}

func (t *TrackingInfo) FormatText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\U0001F4E6 Tracking: %s\n", t.TrackingNumber))
	sb.WriteString(fmt.Sprintf("\U0001F4CA Status: %s\n", t.Status))
	sb.WriteString(fmt.Sprintf("\U0001F550 Updated: %s\n\n", t.LastUpdate.Format("2006-01-02 15:04:05")))

	if len(t.Details) > 0 {
		sb.WriteString("\U0001F4CB Details:\n")
		for i, d := range t.Details {
			flag := SourceFlag(d.Source)
			sb.WriteString(fmt.Sprintf("\n%d. %s %s\n", i+1, flag, d.DateTime))
			if d.Details != "" {
				sb.WriteString(fmt.Sprintf("   \U0001F4CD %s\n", d.Details))
			}
			sb.WriteString(fmt.Sprintf("   \u2139\uFE0F  %s\n", d.Description))
			if d.Office != "" || d.Region != "" {
				loc := d.Office
				if d.Region != "" {
					if loc != "" {
						loc += " - " + d.Region
					} else {
						loc = d.Region
					}
				}
				sb.WriteString(fmt.Sprintf("   \U0001F3E2 %s\n", loc))
			}
		}
	}

	return sb.String()
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
