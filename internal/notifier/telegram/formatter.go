package telegram

import (
	"fmt"
	"strings"

	"github.com/channinghe/ems2sns/internal/model"
)

func formatTrackingUpdate(info *model.TrackingInfo, delivered bool) string {
	icon := "\U0001F514"
	if delivered {
		icon = "\u2705"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s *\u8FFD\u8E2A\u66F4\u65B0\uFF01*\n\n", icon))
	sb.WriteString(info.FormatText())

	if delivered {
		sb.WriteString("\n\n\u2705 \u5305\u88F9\u5DF2\u9001\u8FBE\uFF01\u8FFD\u8E2A\u7ED3\u675F\u3002")
	}
	return sb.String()
}

func formatInitialStatus(info *model.TrackingInfo, delivered bool) string {
	icon := "\U0001F4E6"
	if delivered {
		icon = "\u2705"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s *\u521D\u59CB\u72B6\u6001*\n\n", icon))
	sb.WriteString(info.FormatText())

	if delivered {
		sb.WriteString("\n\n\u2705 \u5305\u88F9\u5DF2\u9001\u8FBE\uFF0C\u5C06\u4E0D\u518D\u7EE7\u7EED\u8FFD\u8E2A\u3002")
	}
	return sb.String()
}

func formatSubscriberPrefix(sub *model.Subscription) string {
	return fmt.Sprintf("\U0001F464 \u8BA2\u9605\u8005: @%s (ID: %s)\n\n", sub.Username, sub.UserID)
}

func formatHelp() string {
	return "*\U0001F4DA \u5E2E\u52A9\u4FE1\u606F*\n\n" +
		"*\u547D\u4EE4\u5217\u8868\uFF1A*\n" +
		"`/sub \u8FFD\u8E2A\u53F7` - \u8BA2\u9605\u8FFD\u8E2A\n" +
		"`/unsub \u8FFD\u8E2A\u53F7` - \u53D6\u6D88\u8BA2\u9605\n" +
		"`/list` - \u67E5\u770B\u8BA2\u9605\u5217\u8868\n" +
		"`/check \u8FFD\u8E2A\u53F7` - \u67E5\u8BE2\u72B6\u6001\n" +
		"`/push` - \u7ACB\u5373\u68C0\u67E5\u6240\u6709\u8BA2\u9605\n" +
		"`/help` - \u663E\u793A\u5E2E\u52A9\n\n" +
		"*\u793A\u4F8B\uFF1A*\n" +
		"`/sub EB123456789CN`\n" +
		"`/check EB123456789CN`\n\n" +
		"*\u529F\u80FD\u8BF4\u660E\uFF1A*\n" +
		"\u2022 \u8BA2\u9605\u540E\u7ACB\u5373\u63A8\u9001\u5F53\u524D\u72B6\u6001\n" +
		"\u2022 \u5B9A\u671F\u81EA\u52A8\u68C0\u67E5\u66F4\u65B0\n" +
		"\u2022 \u6709\u53D8\u5316\u65F6\u81EA\u52A8\u63A8\u9001\u901A\u77E5\n" +
		"\u2022 \u5305\u88F9\u9001\u8FBE\u540E\u81EA\u52A8\u505C\u6B62\u8FFD\u8E2A\n" +
		"\u2022 \U0001F1E8\U0001F1F3 CN\u7ED3\u5C3E\u5355\u53F7\u81EA\u52A8\u67E5\u8BE2\u4E2D\u56FDEMS\n" +
		"\u2022 \U0001F1EF\U0001F1F5\U0001F1E8\U0001F1F3\U0001F1EC\U0001F1E7 \u4E2D\u65E5\u82F1\u4E09\u8BED\u7269\u6D41\u4FE1\u606F\u5408\u5E76\u5C55\u793A\n\n" +
		"\U0001F4A1 \u70B9\u51FB /start \u53EF\u4EE5\u4F7F\u7528\u6309\u94AE\u83DC\u5355"
}
