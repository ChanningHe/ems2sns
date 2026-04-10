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
		b.send(chatID, "\u274C \u60A8\u6CA1\u6709\u6743\u9650\u4F7F\u7528\u6B64\u673A\u5668\u4EBA\u3002")
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

	// Legacy text commands
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
		b.sendMarkdown(chatID, formatHelp())
	case "list":
		b.doList(chatID)
	case "sub", "subscribe":
		if args == "" {
			b.sendMarkdown(chatID, "\u8BF7\u63D0\u4F9B\u8FFD\u8E2A\u53F7\uFF0C\u4F8B\u5982\uFF1A\n`/sub EB123456789CN`")
			return
		}
		b.doSubscribe(chatID, userID, username, cleanTrackingNumber(args))
	case "unsub", "unsubscribe":
		if args == "" {
			b.sendMarkdown(chatID, "\u8BF7\u63D0\u4F9B\u8FFD\u8E2A\u53F7\uFF0C\u4F8B\u5982\uFF1A\n`/unsub EB123456789CN`")
			return
		}
		b.doUnsubscribe(chatID, cleanTrackingNumber(args))
	case "status", "check":
		if args == "" {
			b.sendMarkdown(chatID, "\u8BF7\u63D0\u4F9B\u8FFD\u8E2A\u53F7\uFF0C\u4F8B\u5982\uFF1A\n`/check EB123456789CN`")
			return
		}
		b.doCheck(chatID, cleanTrackingNumber(args))
	case "push":
		b.doPush(chatID, userID)
	default:
		b.send(chatID, "\u672A\u77E5\u547D\u4EE4\u3002\u4F7F\u7528 /help \u67E5\u770B\u53EF\u7528\u547D\u4EE4\u3002")
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
		b.sendMarkdown(chatID, "\U0001F4E6 \u8BF7\u8F93\u5165\u8981\u8BA2\u9605\u7684\u8FFD\u8E2A\u53F7\uFF1A\n\n\u4F7F\u7528\u547D\u4EE4\uFF1A`/sub \u8FFD\u8E2A\u53F7`\n\u4F8B\u5982\uFF1A`/sub EB123456789CN`")
	case "btn_list":
		b.doList(chatID)
	case "btn_unsubscribe":
		b.sendMarkdown(chatID, "\u274C \u8BF7\u8F93\u5165\u8981\u53D6\u6D88\u8BA2\u9605\u7684\u8FFD\u8E2A\u53F7\uFF1A\n\n\u4F7F\u7528\u547D\u4EE4\uFF1A`/unsub \u8FFD\u8E2A\u53F7`")
	case "btn_push":
		b.doPush(chatID, cb.From.ID)
	case "btn_help":
		b.sendMarkdown(chatID, formatHelp())
	}
}

// --- command implementations ---

func (b *Bot) doStart(chatID int64) {
	text := "\U0001F44B \u6B22\u8FCE\u4F7F\u7528 EMS \u8FFD\u8E2A\u673A\u5668\u4EBA\uFF01\n\n" +
		"\u6211\u53EF\u4EE5\u5E2E\u4F60\u8FFD\u8E2A EMS \u5FEB\u9012\uFF0C\u652F\u6301\uFF1A\n" +
		"\U0001F1EF\U0001F1F5 \u65E5\u672C\u90AE\u653F\uFF08Japan Post\uFF09\n" +
		"\U0001F1E8\U0001F1F3 \u4E2D\u56FDEMS\uFF08via 17track\uFF09\n" +
		"\U0001F1EC\U0001F1E7 \u82F1\u6587\u7248\u8FFD\u8E2A\u4FE1\u606F\n\n" +
		"CN\u7ED3\u5C3E\u7684\u5355\u53F7\u4F1A\u81EA\u52A8\u5408\u5E76\u4E2D\u65E5\u82F1\u4E09\u6BB5\u7269\u6D41\u4FE1\u606F\u3002\n\n" +
		"\u8BF7\u9009\u62E9\u64CD\u4F5C\u6216\u4F7F\u7528\u4E0B\u65B9\u547D\u4EE4\uFF1A"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\u2795 \u8BA2\u9605\u8FFD\u8E2A", "btn_subscribe"),
			tgbotapi.NewInlineKeyboardButtonData("\U0001F4CB \u67E5\u770B\u5217\u8868", "btn_list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\u274C \u53D6\u6D88\u8BA2\u9605", "btn_unsubscribe"),
			tgbotapi.NewInlineKeyboardButtonData("\U0001F680 \u7ACB\u5373\u63A8\u9001", "btn_push"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\u2753 \u5E2E\u52A9", "btn_help"),
		),
	)

	m := tgbotapi.NewMessage(chatID, text)
	m.ReplyMarkup = keyboard
	b.api.Send(m)
}

func (b *Bot) doSubscribe(chatID, userID int64, username, trackingNumber string) {
	if len(trackingNumber) < 10 {
		b.send(chatID, "\u274C \u65E0\u6548\u7684\u8FFD\u8E2A\u53F7\u683C\u5F0F\u3002")
		return
	}

	chStr := fmt.Sprintf("%d", chatID)
	uStr := fmt.Sprintf("%d", userID)

	if err := b.tracker.Subscribe(b.ctx, "telegram", chStr, uStr, username, trackingNumber); err != nil {
		b.send(chatID, fmt.Sprintf("\u274C \u8BA2\u9605\u5931\u8D25\uFF1A%v", err))
		return
	}

	b.sendMarkdown(chatID, fmt.Sprintf("\u2705 \u5DF2\u8BA2\u9605\u8FFD\u8E2A\u53F7\uFF1A`%s`\n\n\u6B63\u5728\u83B7\u53D6\u5F53\u524D\u72B6\u6001...", trackingNumber))
}

func (b *Bot) doUnsubscribe(chatID int64, trackingNumber string) {
	chStr := fmt.Sprintf("%d", chatID)

	if err := b.tracker.Unsubscribe("telegram", chStr, trackingNumber); err != nil {
		b.send(chatID, fmt.Sprintf("\u274C \u53D6\u6D88\u8BA2\u9605\u5931\u8D25\uFF1A%v", err))
		return
	}

	b.sendMarkdown(chatID, fmt.Sprintf("\u2705 \u5DF2\u53D6\u6D88\u8BA2\u9605\uFF1A`%s`", trackingNumber))
}

func (b *Bot) doList(chatID int64) {
	chStr := fmt.Sprintf("%d", chatID)
	subs := b.tracker.List("telegram", chStr)

	if len(subs) == 0 {
		b.sendMarkdown(chatID, "\U0001F4ED \u5F53\u524D\u6CA1\u6709\u4EFB\u4F55\u8BA2\u9605\u3002\n\n\u4F7F\u7528 `/sub \u8FFD\u8E2A\u53F7` \u6765\u6DFB\u52A0\u8BA2\u9605\u3002")
		return
	}

	text := fmt.Sprintf("\U0001F4E6 *\u5F53\u524D\u8BA2\u9605\u5217\u8868* (%d)\uFF1A\n\n\u70B9\u51FB\u4E0B\u65B9\u6309\u94AE\u67E5\u8BE2\u8FFD\u8E2A\u72B6\u6001\uFF1A", len(subs))

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, sub := range subs {
		icon := "\U0001F504"
		if sub.IsDelivered {
			icon = "\u2705"
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
	b.sendMarkdown(chatID, fmt.Sprintf("\U0001F50D \u6B63\u5728\u67E5\u8BE2 `%s` \u7684\u72B6\u6001...", trackingNumber))

	info, err := b.tracker.Check(b.ctx, trackingNumber)
	if err != nil {
		b.send(chatID, fmt.Sprintf("\u274C \u67E5\u8BE2\u5931\u8D25\uFF1A%v", err))
		return
	}

	b.send(chatID, info.FormatText())
}

func (b *Bot) doPush(chatID int64, userID int64) {
	log.Printf("[telegram] manual push by uid=%d chat=%d", userID, chatID)
	b.send(chatID, "\U0001F680 \u6B63\u5728\u7ACB\u5373\u68C0\u67E5\u6240\u6709\u8BA2\u9605\u7684\u66F4\u65B0...")

	chStr := fmt.Sprintf("%d", chatID)
	result := b.tracker.ManualPush(b.ctx, "telegram", chStr)

	var text string
	if result.UpdatesFound > 0 {
		text = fmt.Sprintf("\u2705 *\u624B\u52A8\u63A8\u9001\u5B8C\u6210*\n\n"+
			"\U0001F4CA \u68C0\u67E5\u4E86 %d \u4E2A\u8BA2\u9605\n"+
			"\U0001F514 \u53D1\u73B0 %d \u4E2A\u66F4\u65B0\n"+
			"\U0001F4E6 \u5176\u4E2D %d \u4E2A\u5305\u88F9\u5DF2\u9001\u8FBE",
			result.TotalChecked, result.UpdatesFound, result.DeliveredCount)
	} else {
		text = fmt.Sprintf("\u2705 *\u624B\u52A8\u63A8\u9001\u5B8C\u6210*\n\n"+
			"\U0001F4CA \u68C0\u67E5\u4E86 %d \u4E2A\u8BA2\u9605\n"+
			"\U0001F634 \u6682\u65E0\u65B0\u66F4\u65B0",
			result.TotalChecked)
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
	// If no user IDs configured, allow all (rely on chat ID restrictions)
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

// stripLeadingMention removes a leading @BotUsername mention from the message
// text when the first entity is a mention at offset 0. Returns the remaining
// text (trimmed) and whether a mention was actually stripped.
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

// parseCommand extracts a command name and its arguments from raw text like
// "/sub@BotName EB123". The leading "/" and optional @BotName suffix on the
// command token are both removed.
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

// getPushTargets determines where notifications should be sent
func (b *Bot) getPushTargets(originalChatID int64) []int64 {
	if len(b.cfg.PushChatIDs) > 0 && originalChatID > 0 {
		return b.cfg.PushChatIDs
	}
	return []int64{originalChatID}
}

