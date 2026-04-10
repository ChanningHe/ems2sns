package telegram

import (
	"fmt"
	"strings"

	"github.com/channinghe/ems2sns/internal/i18n"
	"github.com/channinghe/ems2sns/internal/model"
)

func formatTrackingUpdate(msg *i18n.Messages, info *model.TrackingInfo, delivered bool) string {
	var sb strings.Builder
	if delivered {
		sb.WriteString(fmt.Sprintf("✅ *%s*\n\n", msg.TrackingUpdate))
	} else {
		sb.WriteString(fmt.Sprintf("🔔 *%s*\n\n", msg.TrackingUpdate))
	}
	sb.WriteString(info.FormatText())

	if delivered {
		sb.WriteString("\n\n✅ " + msg.DeliveredEnd)
	}
	return sb.String()
}

func formatInitialStatus(msg *i18n.Messages, info *model.TrackingInfo, delivered bool) string {
	var sb strings.Builder
	if delivered {
		sb.WriteString(fmt.Sprintf("✅ *%s*\n\n", msg.InitialStatus))
	} else {
		sb.WriteString(fmt.Sprintf("📦 *%s*\n\n", msg.InitialStatus))
	}
	sb.WriteString(info.FormatText())

	if delivered {
		sb.WriteString("\n\n✅ " + msg.DeliveredNoTrack)
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
