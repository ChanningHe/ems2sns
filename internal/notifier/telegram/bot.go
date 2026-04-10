package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/channinghe/ems2sns/internal/config"
	"github.com/channinghe/ems2sns/internal/i18n"
	"github.com/channinghe/ems2sns/internal/model"
	"github.com/channinghe/ems2sns/internal/tracker"
)

// Bot implements notifier.Notifier for Telegram
type Bot struct {
	api     *tgbotapi.BotAPI
	cfg     config.TelegramConfig
	tracker *tracker.Tracker
	msg     *i18n.Messages
	ctx     context.Context
	cancel  context.CancelFunc
	debug   bool
}

func New(cfg config.TelegramConfig, trk *tracker.Tracker, logLevel string, msg *i18n.Messages) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot: %w", err)
	}

	debug := strings.EqualFold(logLevel, "debug")
	if debug {
		api.Debug = true
		log.Println("[telegram] debug mode enabled (api.Debug=true)")
	}

	log.Printf("[telegram] authorized as @%s (log_level=%s)", api.Self.UserName, logLevel)

	ctx, cancel := context.WithCancel(context.Background())
	b := &Bot{
		api:     api,
		cfg:     cfg,
		tracker: trk,
		msg:     msg,
		ctx:     ctx,
		cancel:  cancel,
		debug:   debug,
	}

	trk.RegisterNotifyFunc(b.onTrackingUpdate)

	return b, nil
}

func (b *Bot) Platform() string { return "telegram" }

func (b *Bot) Start(ctx context.Context) error {
	b.ctx = ctx
	log.Printf("[telegram] bot handler started (debug=%v)", b.debug)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return nil
		case update := <-updates:
			if b.debug {
				b.logRawUpdate(update)
			}
			b.handleUpdate(update)
		}
	}
}

func (b *Bot) logRawUpdate(update tgbotapi.Update) {
	parts := []string{fmt.Sprintf("update_id=%d", update.UpdateID)}
	if update.Message != nil {
		m := update.Message
		parts = append(parts, fmt.Sprintf("message(id=%d chat=%d from=%s text=%q)",
			m.MessageID, m.Chat.ID, m.From.UserName, m.Text))
		if len(m.Entities) > 0 {
			for i, e := range m.Entities {
				parts = append(parts, fmt.Sprintf("  entity[%d] type=%s offset=%d length=%d",
					i, e.Type, e.Offset, e.Length))
			}
		}
	}
	if update.EditedMessage != nil {
		parts = append(parts, fmt.Sprintf("edited_message(id=%d chat=%d)",
			update.EditedMessage.MessageID, update.EditedMessage.Chat.ID))
	}
	if update.CallbackQuery != nil {
		parts = append(parts, fmt.Sprintf("callback(id=%s data=%q from=%s)",
			update.CallbackQuery.ID, update.CallbackQuery.Data, update.CallbackQuery.From.UserName))
	}
	if update.Message == nil && update.EditedMessage == nil && update.CallbackQuery == nil {
		parts = append(parts, "type=other (no message/callback)")
	}
	log.Printf("[telegram][debug] raw update: %s", strings.Join(parts, " | "))
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

	message := formatTrackingUpdate(b.msg, info, delivered)
	targets := b.getPushTargets(chatID)

	for _, targetID := range targets {
		text := message
		if targetID != chatID {
			text = formatSubscriberPrefix(b.msg, sub) + text
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
