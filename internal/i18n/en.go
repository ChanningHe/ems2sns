package i18n

func english() *Messages {
	return &Messages{
		Unauthorized:      "You are not authorized to use this bot.",
		UnknownCmd:        "Unknown command. Use /help to see available commands.",
		InvalidTrackingNo: "Invalid tracking number format.",

		SubNeedArgs: "Please provide a tracking number, e.g.:\n`/sub EB123456789CN`",
		Subscribed:  "Subscribed to `%s`\n\nFetching current status...",
		SubFailed:   "Subscribe failed: %v",

		UnsubNeedArgs: "Please provide a tracking number, e.g.:\n`/unsub EB123456789CN`",
		Unsubscribed:  "Unsubscribed from `%s`.",
		UnsubFailed:   "Unsubscribe failed: %v",

		ListEmpty: "No active subscriptions.\n\nUse `/sub <tracking_number>` to add one.",
		ListTitle: "Subscriptions (%d):",

		CheckNeedArgs: "Please provide a tracking number, e.g.:\n`/check EB123456789CN`",
		Checking:      "Checking status of `%s`...",
		CheckFailed:   "Check failed: %v",

		PushStarting:      "Checking all subscriptions for updates...",
		PushDoneUpdates:   "Manual push done\n\nChecked: %d\nUpdates: %d\nDelivered: %d",
		PushDoneNoUpdates: "Manual push done\n\nChecked: %d\nNo new updates.",

		Welcome: "Welcome to EMS Tracking Bot!\n\n" +
			"I can help you track EMS packages:\n" +
			"Japan Post (JP/EN)\n" +
			"China EMS (via 17track)\n\n" +
			"CN-suffix numbers auto-merge JP + CN + EN tracking.\n\n" +
			"Choose an action or use commands below:",

		BtnSubscribe:   "Subscribe",
		BtnList:        "List",
		BtnUnsubscribe: "Unsubscribe",
		BtnPush:        "Push Now",
		BtnHelp:        "Help",

		PromptSub:   "Enter tracking number to subscribe:\n\nCommand: `/sub <number>`\nExample: `/sub EB123456789CN`",
		PromptUnsub: "Enter tracking number to unsubscribe:\n\nCommand: `/unsub <number>`",

		TrackingUpdate:   "Tracking Update!",
		InitialStatus:    "Initial Status",
		DeliveredEnd:     "Package delivered! Tracking ended.",
		DeliveredNoTrack: "Package already delivered. No further tracking.",
		SubscriberFmt:    "Subscriber: @%s (ID: %s)",
		PackageDelivered: "Package delivered!",

		HelpTitle:         "Help",
		HelpCmdSub:        "`/sub <number>` - Subscribe to tracking",
		HelpCmdUnsub:      "`/unsub <number>` - Unsubscribe",
		HelpCmdList:       "`/list` - List subscriptions",
		HelpCmdCheck:      "`/check <number>` - Check status",
		HelpCmdPush:       "`/push` - Check all subscriptions now",
		HelpCmdHelp:       "`/help` - Show help",
		HelpFeatureAuto:   "Auto-check at regular intervals",
		HelpFeatureNotify: "Push notifications on status changes",
		HelpFeatureStop:   "Auto-stop tracking after delivery",
		HelpFeatureCN:     "CN-suffix numbers auto-query China EMS",
		HelpFeatureMerge:  "JP/CN/EN tracking info merged",
		HelpButtonTip:     "Use /start for button menu",
	}
}
