package discord

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

var slashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "sub",
		Description: "Subscribe to a tracking number",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "tracking_number",
				Description: "EMS tracking number",
				Required:    true,
			},
		},
	},
	{
		Name:        "unsub",
		Description: "Unsubscribe from a tracking number",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "tracking_number",
				Description: "EMS tracking number",
				Required:    true,
			},
		},
	},
	{
		Name:        "list",
		Description: "List current subscriptions",
	},
	{
		Name:        "check",
		Description: "Check tracking status",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "tracking_number",
				Description: "EMS tracking number",
				Required:    true,
			},
		},
	},
	{
		Name:        "push",
		Description: "Manually check all subscriptions now",
	},
	{
		Name:        "emshelp",
		Description: "Show help information",
	},
}

func (b *Bot) registerCommands() error {
	for _, cmd := range slashCommands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("registering command %s: %w", cmd.Name, err)
		}
	}
	log.Println("[discord] slash commands registered")
	return nil
}

func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	channelID := i.ChannelID
	userID := ""
	username := ""
	if i.Member != nil {
		userID = i.Member.User.ID
		username = i.Member.User.Username
	} else if i.User != nil {
		userID = i.User.ID
		username = i.User.Username
	}

	if !b.isAuthorized(i) {
		b.respond(s, i, "❌ "+b.msg.Unauthorized)
		return
	}

	data := i.ApplicationCommandData()
	switch data.Name {
	case "sub":
		tn := data.Options[0].StringValue()
		b.doSubscribe(s, i, channelID, userID, username, tn)
	case "unsub":
		tn := data.Options[0].StringValue()
		b.doUnsubscribe(s, i, channelID, tn)
	case "list":
		b.doList(s, i, channelID)
	case "check":
		tn := data.Options[0].StringValue()
		b.doCheck(s, i, tn)
	case "push":
		b.doPush(s, i, channelID)
	case "emshelp":
		b.respondEmbed(s, i, helpEmbed(b.msg))
	}
}

func (b *Bot) doSubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, channelID, userID, username, tn string) {
	if len(tn) < 10 {
		b.respond(s, i, "❌ "+b.msg.InvalidTrackingNo)
		return
	}

	if err := b.tracker.Subscribe(b.ctx, "discord", channelID, userID, username, tn); err != nil {
		b.respond(s, i, "❌ "+fmt.Sprintf(b.msg.SubFailed, err))
		return
	}

	b.respond(s, i, "✅ "+fmt.Sprintf(b.msg.Subscribed, tn))
}

func (b *Bot) doUnsubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, channelID, tn string) {
	if err := b.tracker.Unsubscribe("discord", channelID, tn); err != nil {
		b.respond(s, i, "❌ "+fmt.Sprintf(b.msg.UnsubFailed, err))
		return
	}

	b.respond(s, i, "✅ "+fmt.Sprintf(b.msg.Unsubscribed, tn))
}

func (b *Bot) doList(s *discordgo.Session, i *discordgo.InteractionCreate, channelID string) {
	subs := b.tracker.List("discord", channelID)
	if len(subs) == 0 {
		b.respond(s, i, "📭 "+b.msg.ListEmpty)
		return
	}

	text := fmt.Sprintf("📦 **"+b.msg.ListTitle+"**\n\n", len(subs))
	for _, sub := range subs {
		icon := "🔄"
		if sub.IsDelivered {
			icon = "✅"
		}
		text += fmt.Sprintf("%s `%s`\n", icon, sub.TrackingNumber)
	}
	b.respond(s, i, text)
}

func (b *Bot) doCheck(s *discordgo.Session, i *discordgo.InteractionCreate, tn string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	update, err := b.tracker.Check(b.ctx, tn)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("❌ " + fmt.Sprintf(b.msg.CheckFailed, err)),
		})
		return
	}

	embed := trackingEmbed(b.msg, update)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

func (b *Bot) doPush(s *discordgo.Session, i *discordgo.InteractionCreate, channelID string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	result := b.tracker.ManualPush(b.ctx, "discord", channelID)

	var text string
	if result.UpdatesFound > 0 {
		text = "✅ **" + fmt.Sprintf(b.msg.PushDoneUpdates,
			result.TotalChecked, result.UpdatesFound, result.DeliveredCount) + "**"
	} else {
		text = "✅ **" + fmt.Sprintf(b.msg.PushDoneNoUpdates, result.TotalChecked) + "**"
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &text,
	})
}

// --- helpers ---

func (b *Bot) isAuthorized(i *discordgo.InteractionCreate) bool {
	if len(b.cfg.AllowedGuildIDs) == 0 && len(b.cfg.AllowedChannelIDs) == 0 {
		return true
	}

	if i.GuildID != "" {
		for _, gid := range b.cfg.AllowedGuildIDs {
			if gid == i.GuildID {
				return true
			}
		}
	}

	for _, cid := range b.cfg.AllowedChannelIDs {
		if cid == i.ChannelID {
			return true
		}
	}

	return false
}

func (b *Bot) respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
}

func (b *Bot) respondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
	})
}

func strPtr(s string) *string { return &s }
