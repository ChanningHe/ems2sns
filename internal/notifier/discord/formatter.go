package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/channinghe/ems2sns/internal/i18n"
	"github.com/channinghe/ems2sns/internal/model"
)

const detailsPerSourceLimit = 10

func trackingEmbed(msg *i18n.Messages, update *model.TrackingUpdate) *discordgo.MessageEmbed {
	color := 0x3498db
	if update.Delivered {
		color = 0x2ecc71
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(update.Segments))
	for _, seg := range update.Segments {
		fields = append(fields, segmentField(seg))
	}

	var desc strings.Builder
	if update.Initial {
		desc.WriteString("📦 " + msg.InitialStatus)
	} else {
		desc.WriteString("🔔 " + msg.TrackingUpdate)
	}
	if update.Delivered {
		desc.WriteString("\n✅ " + msg.PackageDelivered)
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📦 %s", update.TrackingNumber),
		Description: desc.String(),
		Color:       color,
		Fields:      fields,
	}
}

func segmentField(seg model.SourceSegment) *discordgo.MessageEmbedField {
	flag := model.SourceFlag(seg.Source)
	var body strings.Builder

	if seg.Status != "" {
		body.WriteString(fmt.Sprintf("**Status:** %s\n", seg.Status))
	}
	if !seg.LastUpdate.IsZero() {
		body.WriteString(fmt.Sprintf("**Updated:** %s\n", seg.LastUpdate.Format("2006-01-02 15:04:05")))
	}

	if len(seg.Details) > 0 {
		body.WriteString("\n")
		limit := len(seg.Details)
		if limit > detailsPerSourceLimit {
			limit = detailsPerSourceLimit
		}
		for i := 0; i < limit; i++ {
			d := seg.Details[i]
			line := fmt.Sprintf("**%s** %s", d.DateTime, d.Description)
			if d.Office != "" {
				line += fmt.Sprintf(" (%s)", d.Office)
			}
			body.WriteString(line + "\n")
		}
		if len(seg.Details) > detailsPerSourceLimit {
			body.WriteString(fmt.Sprintf("... and %d more events\n", len(seg.Details)-detailsPerSourceLimit))
		}
	}

	return &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s %s", flag, seg.Source),
		Value:  body.String(),
		Inline: false,
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
