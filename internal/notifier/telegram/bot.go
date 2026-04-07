package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/channinghe/ems2sns/internal/config"
	"github.com/channinghe/ems2sns/internal/model"
	"github.com/channinghe/ems2sns/internal/tracker"
)

// Bot implements notifier.Notifier for Telegram
type Bot struct {
	api     *tgbotapi.BotAPI
	cfg     config.TelegramConfig
	tracker *tracker.Tracker
	ctx     context.Context
	cancel  context.CancelFunc
}

func New(cfg config.TelegramConfig, trk *tracker.Tracker) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot: %w", err)
	}

	log.Printf("[telegram] authorized as @%s", api.Self.UserName)

	ctx, cancel := context.WithCancel(context.Background())
	b := &Bot{
		api:     api,
		cfg:     cfg,
		tracker: trk,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Register this bot's notification callback with the tracker
	trk.RegisterNotifyFunc(b.onTrackingUpdate)

	return b, nil
}

func (b *Bot) Platform() string { return "telegram" }

func (b *Bot) Start(ctx context.Context) error {
	b.ctx = ctx
	log.Println("[telegram] bot handler started")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return nil
		case update := <-updates:
			b.handleUpdate(update)
		}
	}
}

func (b *Bot) Stop() error {
	b.cancel()
	b.api.StopReceivingUpdates()
	log.Println("[telegram] bot stopped")
	return nil
}

func (b *Bot) SendUpdate(sub *model.Subscription, info *model.TrackingInfo, delivered bool) error {
	chatID, err := strconv.ParseInt(sub.ChannelID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID %q: %w", sub.ChannelID, err)
	}

	message := formatTrackingUpdate(info, delivered)
	targets := b.getPushTargets(chatID)

	for _, targetID := range targets {
		text := message
		if targetID != chatID {
			text = formatSubscriberPrefix(sub) + text
		}
		b.sendMarkdown(targetID, text)
	}

	return nil
}

// onTrackingUpdate is registered as a NotifyFunc with the tracker.
// It only handles notifications for Telegram subscriptions.
func (b *Bot) onTrackingUpdate(sub *model.Subscription, info *model.TrackingInfo, delivered bool) {
	if sub.Platform != "telegram" {
		return
	}
	if err := b.SendUpdate(sub, info, delivered); err != nil {
		log.Printf("[telegram] notification error: %v", err)
	}
}
