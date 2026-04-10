package telegram

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		b.handleCallback(update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	msg := update.Message
	userID := msg.From.ID
	chatID := msg.Chat.ID
	username := msg.From.UserName

	log.Printf("[telegram] msg from %s (uid=%d, chat=%d): %s", username, userID, chatID, msg.Text)

	if !b.isAuthorized(msg) {
		b.send(chatID, "❌ "+b.msg.Unauthorized)
		return
	}

	text := strings.TrimSpace(msg.Text)

	if strings.HasPrefix(text, "/") {
		b.handleCommand(msg)
		return
	}

	// Handle "@BotName /command args" pattern in group chats
	if stripped, ok := stripLeadingMention(msg); ok && strings.HasPrefix(stripped, "/") {
		cmd, args := parseCommand(stripped)
		log.Printf("[telegram] mention-prefixed command from %s: cmd=%s args=%s", username, cmd, args)
		b.dispatchCommand(chatID, userID, username, cmd, args)
		return
	}

	// Legacy text commands (Chinese)
	subRe := regexp.MustCompile(`(?i)订阅\s*EMS\s+([A-Z0-9]+)`)
	if m := subRe.FindStringSubmatch(text); len(m) > 1 {
		b.doSubscribe(chatID, userID, username, strings.ToUpper(m[1]))
		return
	}

	unsubRe := regexp.MustCompile(`(?i)取消订阅\s*EMS\s+([A-Z0-9]+)`)
	if m := unsubRe.FindStringSubmatch(text); len(m) > 1 {
		b.doUnsubscribe(chatID, strings.ToUpper(m[1]))
		return
	}
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	b.dispatchCommand(msg.Chat.ID, msg.From.ID, msg.From.UserName, msg.Command(), msg.CommandArguments())
}

func (b *Bot) dispatchCommand(chatID, userID int64, username, cmd, args string) {
	switch cmd {
	case "start":
		b.doStart(chatID)
	case "help":
		b.sendMarkdown(chatID, formatHelp(b.msg))
	case "list":
		b.doList(chatID)
	case "sub", "subscribe":
		if args == "" {
			b.sendMarkdown(chatID, b.msg.SubNeedArgs)
			return
		}
		b.doSubscribe(chatID, userID, username, cleanTrackingNumber(args))
	case "unsub", "unsubscribe":
		if args == "" {
			b.sendMarkdown(chatID, b.msg.UnsubNeedArgs)
			return
		}
		b.doUnsubscribe(chatID, cleanTrackingNumber(args))
	case "status", "check":
		if args == "" {
			b.sendMarkdown(chatID, b.msg.CheckNeedArgs)
			return
		}
		b.doCheck(chatID, cleanTrackingNumber(args))
	case "push":
		b.doPush(chatID, userID)
	default:
		b.send(chatID, b.msg.UnknownCmd)
	}
}

func (b *Bot) handleCallback(cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID

	ack := tgbotapi.NewCallback(cb.ID, "")
	b.api.Send(ack)

	data := cb.Data

	if strings.HasPrefix(data, "check_") {
		tn := strings.TrimPrefix(data, "check_")
		b.doCheck(chatID, tn)
		return
	}

	switch data {
	case "btn_subscribe":
		b.sendMarkdown(chatID, "📦 "+b.msg.PromptSub)
	case "btn_list":
		b.doList(chatID)
	case "btn_unsubscribe":
		b.sendMarkdown(chatID, "❌ "+b.msg.PromptUnsub)
	case "btn_push":
		b.doPush(chatID, cb.From.ID)
	case "btn_help":
		b.sendMarkdown(chatID, formatHelp(b.msg))
	}
}

// --- command implementations ---

func (b *Bot) doStart(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ "+b.msg.BtnSubscribe, "btn_subscribe"),
			tgbotapi.NewInlineKeyboardButtonData("📋 "+b.msg.BtnList, "btn_list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ "+b.msg.BtnUnsubscribe, "btn_unsubscribe"),
			tgbotapi.NewInlineKeyboardButtonData("🚀 "+b.msg.BtnPush, "btn_push"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ "+b.msg.BtnHelp, "btn_help"),
		),
	)

	m := tgbotapi.NewMessage(chatID, "👋 "+b.msg.Welcome)
	m.ReplyMarkup = keyboard
	b.api.Send(m)
}

func (b *Bot) doSubscribe(chatID, userID int64, username, trackingNumber string) {
	if len(trackingNumber) < 10 {
		b.send(chatID, "❌ "+b.msg.InvalidTrackingNo)
		return
	}

	chStr := fmt.Sprintf("%d", chatID)
	uStr := fmt.Sprintf("%d", userID)

	if err := b.tracker.Subscribe(b.ctx, "telegram", chStr, uStr, username, trackingNumber); err != nil {
		b.send(chatID, "❌ "+fmt.Sprintf(b.msg.SubFailed, err))
		return
	}

	b.sendMarkdown(chatID, "✅ "+fmt.Sprintf(b.msg.Subscribed, trackingNumber))
}

func (b *Bot) doUnsubscribe(chatID int64, trackingNumber string) {
	chStr := fmt.Sprintf("%d", chatID)

	if err := b.tracker.Unsubscribe("telegram", chStr, trackingNumber); err != nil {
		b.send(chatID, "❌ "+fmt.Sprintf(b.msg.UnsubFailed, err))
		return
	}

	b.sendMarkdown(chatID, "✅ "+fmt.Sprintf(b.msg.Unsubscribed, trackingNumber))
}

func (b *Bot) doList(chatID int64) {
	chStr := fmt.Sprintf("%d", chatID)
	subs := b.tracker.List("telegram", chStr)

	if len(subs) == 0 {
		b.sendMarkdown(chatID, "📭 "+b.msg.ListEmpty)
		return
	}

	text := fmt.Sprintf("📦 *"+b.msg.ListTitle+"*", len(subs))

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, sub := range subs {
		icon := "🔄"
		if sub.IsDelivered {
			icon = "✅"
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s %s", icon, sub.TrackingNumber),
			fmt.Sprintf("check_%s", sub.TrackingNumber),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = "Markdown"
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.api.Send(m)
}

func (b *Bot) doCheck(chatID int64, trackingNumber string) {
	b.sendMarkdown(chatID, "🔍 "+fmt.Sprintf(b.msg.Checking, trackingNumber))

	info, err := b.tracker.Check(b.ctx, trackingNumber)
	if err != nil {
		b.send(chatID, "❌ "+fmt.Sprintf(b.msg.CheckFailed, err))
		return
	}

	b.send(chatID, info.FormatText())
}

func (b *Bot) doPush(chatID int64, userID int64) {
	log.Printf("[telegram] manual push by uid=%d chat=%d", userID, chatID)
	b.send(chatID, "🚀 "+b.msg.PushStarting)

	chStr := fmt.Sprintf("%d", chatID)
	result := b.tracker.ManualPush(b.ctx, "telegram", chStr)

	var text string
	if result.UpdatesFound > 0 {
		text = "✅ *" + fmt.Sprintf(b.msg.PushDoneUpdates,
			result.TotalChecked, result.UpdatesFound, result.DeliveredCount) + "*"
	} else {
		text = "✅ *" + fmt.Sprintf(b.msg.PushDoneNoUpdates, result.TotalChecked) + "*"
	}

	b.sendMarkdown(chatID, text)
}

// --- helpers ---

func (b *Bot) isAuthorized(msg *tgbotapi.Message) bool {
	if msg.Chat.IsPrivate() {
		return b.isUserAllowed(msg.From.ID)
	}
	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		return b.isChatAllowed(msg.Chat.ID)
	}
	return false
}

func (b *Bot) isUserAllowed(uid int64) bool {
	for _, id := range b.cfg.AllowedUserIDs {
		if id == uid {
			return true
		}
	}
	return len(b.cfg.AllowedUserIDs) == 0
}

func (b *Bot) isChatAllowed(chatID int64) bool {
	for _, id := range b.cfg.AllowedChatIDs {
		if id == chatID {
			return true
		}
	}
	return len(b.cfg.AllowedChatIDs) == 0
}

func (b *Bot) send(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(m); err != nil {
		log.Printf("[telegram] send error to %d: %v", chatID, err)
	}
}

func (b *Bot) sendMarkdown(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = "Markdown"
	if _, err := b.api.Send(m); err != nil {
		log.Printf("[telegram] send error to %d: %v", chatID, err)
	}
}

func stripLeadingMention(msg *tgbotapi.Message) (string, bool) {
	if len(msg.Entities) == 0 {
		return strings.TrimSpace(msg.Text), false
	}
	first := msg.Entities[0]
	if first.Type != "mention" || first.Offset != 0 {
		return strings.TrimSpace(msg.Text), false
	}
	stripped := strings.TrimSpace(msg.Text[first.Length:])
	return stripped, true
}

func parseCommand(text string) (cmd, args string) {
	if !strings.HasPrefix(text, "/") {
		return "", text
	}
	text = text[1:]
	parts := strings.SplitN(text, " ", 2)
	cmd = parts[0]
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}
	if idx := strings.Index(cmd, "@"); idx != -1 {
		cmd = cmd[:idx]
	}
	return strings.ToLower(cmd), args
}

func cleanTrackingNumber(args string) string {
	if idx := strings.Index(args, "@"); idx != -1 {
		args = args[:idx]
	}
	tn := strings.ToUpper(strings.TrimSpace(args))
	if parts := strings.Fields(tn); len(parts) > 0 {
		return parts[0]
	}
	return tn
}

func (b *Bot) getPushTargets(originalChatID int64) []int64 {
	if len(b.cfg.PushChatIDs) > 0 && originalChatID > 0 {
		return b.cfg.PushChatIDs
	}
	return []int64{originalChatID}
}
