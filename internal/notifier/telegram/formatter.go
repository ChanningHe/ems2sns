package telegram

import (
	"fmt"
	"strings"

	"github.com/channinghe/ems2sns/internal/i18n"
	"github.com/channinghe/ems2sns/internal/model"
)

func formatTrackingUpdate(msg *i18n.Messages, update *model.TrackingUpdate) string {
	var sb strings.Builder

	header := msg.TrackingUpdate
	icon := "🔔"
	if update.Initial {
		header = msg.InitialStatus
		icon = "📦"
	}
	if update.Delivered {
		icon = "✅"
	}
	sb.WriteString(fmt.Sprintf("%s *%s*\n", icon, header))
	sb.WriteString(fmt.Sprintf("📦 `%s`\n\n", update.TrackingNumber))

	for i, seg := range update.Segments {
		if i > 0 {
			sb.WriteString("\n────────────\n\n")
		}
		sb.WriteString(renderSegment(seg))
	}

	if update.Delivered {
		tail := msg.DeliveredEnd
		if update.Initial {
			tail = msg.DeliveredNoTrack
		}
		sb.WriteString("\n\n✅ " + tail)
	}
	return sb.String()
}

func renderSegment(seg model.SourceSegment) string {
	var sb strings.Builder
	flag := model.SourceFlag(seg.Source)

	sb.WriteString(fmt.Sprintf("%s *%s*\n", flag, seg.Source))
	if seg.Status != "" {
		sb.WriteString(fmt.Sprintf("📊 %s\n", seg.Status))
	}
	if !seg.LastUpdate.IsZero() {
		sb.WriteString(fmt.Sprintf("🕐 %s\n", seg.LastUpdate.Format("2006-01-02 15:04:05")))
	}

	if len(seg.Details) == 0 {
		return sb.String()
	}

	sb.WriteString("\n")
	for i, d := range seg.Details {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, d.DateTime))
		if d.Details != "" {
			sb.WriteString(fmt.Sprintf("   📍 %s\n", d.Details))
		}
		if d.Description != "" {
			sb.WriteString(fmt.Sprintf("   ℹ️  %s\n", d.Description))
		}
		if d.Office != "" || d.Region != "" {
			loc := d.Office
			if d.Region != "" {
				if loc != "" {
					loc += " - " + d.Region
				} else {
					loc = d.Region
				}
			}
			sb.WriteString(fmt.Sprintf("   🏢 %s\n", loc))
		}
	}
	return sb.String()
}

func formatSubscriberPrefix(msg *i18n.Messages, sub *model.Subscription) string {
	return "👤 " + fmt.Sprintf(msg.SubscriberFmt, sub.Username, sub.UserID) + "\n\n"
}

func formatHelp(msg *i18n.Messages) string {
	return fmt.Sprintf("*📚 %s*\n\n", msg.HelpTitle) +
		fmt.Sprintf("*Commands:*\n%s\n%s\n%s\n%s\n%s\n%s\n\n",
			msg.HelpCmdSub, msg.HelpCmdUnsub, msg.HelpCmdList,
			msg.HelpCmdCheck, msg.HelpCmdPush, msg.HelpCmdHelp) +
		fmt.Sprintf("*Features:*\n• %s\n• %s\n• %s\n• %s\n• %s\n\n",
			msg.HelpFeatureAuto, msg.HelpFeatureNotify, msg.HelpFeatureStop,
			msg.HelpFeatureCN, msg.HelpFeatureMerge) +
		"💡 " + msg.HelpButtonTip
}
