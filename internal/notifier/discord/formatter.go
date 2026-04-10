package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/channinghe/ems2sns/internal/i18n"
	"github.com/channinghe/ems2sns/internal/model"
)

func trackingEmbed(msg *i18n.Messages, info *model.TrackingInfo, delivered bool) *discordgo.MessageEmbed {
	color := 0x3498db
	if delivered {
		color = 0x2ecc71
	}

	var desc strings.Builder
	desc.WriteString(fmt.Sprintf("**Status:** %s\n", info.Status))
	desc.WriteString(fmt.Sprintf("**Updated:** %s\n\n", info.LastUpdate.Format("2006-01-02 15:04:05")))

	if len(info.Details) > 0 {
		limit := len(info.Details)
		if limit > 15 {
			limit = 15
		}
		for i := 0; i < limit; i++ {
			d := info.Details[i]
			flag := model.SourceFlag(d.Source)
			line := fmt.Sprintf("%s **%s** %s", flag, d.DateTime, d.Description)
			if d.Office != "" {
				line += fmt.Sprintf(" (%s)", d.Office)
			}
			desc.WriteString(line + "\n")
		}
		if len(info.Details) > 15 {
			desc.WriteString(fmt.Sprintf("\n... and %d more events", len(info.Details)-15))
		}
	}

	if delivered {
		desc.WriteString("\n✅ " + msg.PackageDelivered)
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📦 %s", info.TrackingNumber),
		Description: desc.String(),
		Color:       color,
	}
}

func helpEmbed(msg *i18n.Messages) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("📚 %s", msg.HelpTitle),
		Description: fmt.Sprintf("**Commands:**\n%s\n%s\n%s\n%s\n%s\n%s\n\n",
			msg.HelpCmdSub, msg.HelpCmdUnsub, msg.HelpCmdList,
			msg.HelpCmdCheck, msg.HelpCmdPush, msg.HelpCmdHelp) +
			fmt.Sprintf("**Features:**\n• %s\n• %s\n• %s\n• %s\n• %s",
				msg.HelpFeatureAuto, msg.HelpFeatureNotify, msg.HelpFeatureStop,
				msg.HelpFeatureCN, msg.HelpFeatureMerge),
		Color: 0x3498db,
	}
}
