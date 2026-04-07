package discord

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/channinghe/ems2sns/internal/config"
	"github.com/channinghe/ems2sns/internal/model"
	"github.com/channinghe/ems2sns/internal/tracker"
)

// Bot implements notifier.Notifier for Discord
type Bot struct {
	session *discordgo.Session
	cfg     config.DiscordConfig
	tracker *tracker.Tracker
	ctx     context.Context
	cancel  context.CancelFunc
}

func New(cfg config.DiscordConfig, trk *tracker.Tracker) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("creating discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	ctx, cancel := context.WithCancel(context.Background())
	b := &Bot{
		session: session,
		cfg:     cfg,
		tracker: trk,
		ctx:     ctx,
		cancel:  cancel,
	}

	session.AddHandler(b.onInteraction)
	trk.RegisterNotifyFunc(b.onTrackingUpdate)

	return b, nil
}

func (b *Bot) Platform() string { return "discord" }

func (b *Bot) Start(ctx context.Context) error {
	b.ctx = ctx

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("opening discord session: %w", err)
	}

	log.Printf("[discord] connected as %s#%s", b.session.State.User.Username, b.session.State.User.Discriminator)

	if err := b.registerCommands(); err != nil {
		return fmt.Errorf("registering slash commands: %w", err)
	}

	// Block until context is cancelled
	<-ctx.Done()
	return nil
}

func (b *Bot) Stop() error {
	b.cancel()
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("closing discord session: %w", err)
	}
	log.Println("[discord] bot stopped")
	return nil
}

func (b *Bot) SendUpdate(sub *model.Subscription, info *model.TrackingInfo, delivered bool) error {
	embed := trackingEmbed(info, delivered)
	_, err := b.session.ChannelMessageSendEmbed(sub.ChannelID, embed)
	if err != nil {
		return fmt.Errorf("sending to channel %s: %w", sub.ChannelID, err)
	}
	return nil
}

func (b *Bot) onTrackingUpdate(sub *model.Subscription, info *model.TrackingInfo, delivered bool) {
	if sub.Platform != "discord" {
		return
	}

	targets := b.getPushTargets(sub.ChannelID)
	for _, chID := range targets {
		embed := trackingEmbed(info, delivered)
		if _, err := b.session.ChannelMessageSendEmbed(chID, embed); err != nil {
			log.Printf("[discord] notification error to %s: %v", chID, err)
		}
	}
}

func (b *Bot) getPushTargets(originalChannelID string) []string {
	if len(b.cfg.PushChannelIDs) > 0 {
		return b.cfg.PushChannelIDs
	}
	return []string{originalChannelID}
}
