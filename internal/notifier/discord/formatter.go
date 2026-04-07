package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/channinghe/ems2sns/internal/model"
)

func trackingEmbed(info *model.TrackingInfo, delivered bool) *discordgo.MessageEmbed {
	color := 0x3498db // blue
	if delivered {
		color = 0x2ecc71 // green
	}

	var desc strings.Builder
	desc.WriteString(fmt.Sprintf("**Status:** %s\n", info.Status))
	desc.WriteString(fmt.Sprintf("**Updated:** %s\n\n", info.LastUpdate.Format("2006-01-02 15:04:05")))

	if len(info.Details) > 0 {
		// Show at most 15 events in embed to avoid length limit
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
		desc.WriteString("\n\u2705 Package delivered!")
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("\U0001F4E6 %s", info.TrackingNumber),
		Description: desc.String(),
		Color:       color,
	}
}

func helpEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "\U0001F4DA EMS Tracking Bot - Help",
		Description: "**Commands:**\n" +
			"`/sub <tracking_number>` - Subscribe\n" +
			"`/unsub <tracking_number>` - Unsubscribe\n" +
			"`/list` - List subscriptions\n" +
			"`/check <tracking_number>` - Check status\n" +
			"`/push` - Manual check all subscriptions\n" +
			"`/help` - Show this help\n\n" +
			"**Features:**\n" +
			"\u2022 Auto-check at regular intervals\n" +
			"\u2022 Push notifications on updates\n" +
			"\u2022 \U0001F1EF\U0001F1F5\U0001F1E8\U0001F1F3\U0001F1EC\U0001F1E7 JP/CN/EN tracking merged\n" +
			"\u2022 CN-suffix numbers auto-query China EMS",
		Color: 0x3498db,
	}
}
