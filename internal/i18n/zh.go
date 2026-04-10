package i18n

func chinese() *Messages {
	return &Messages{
		Unauthorized:      "您没有权限使用此机器人。",
		UnknownCmd:        "未知命令。使用 /help 查看可用命令。",
		InvalidTrackingNo: "无效的追踪号格式。",

		SubNeedArgs: "请提供追踪号，例如：\n`/sub EB123456789CN`",
		Subscribed:  "已订阅追踪号：`%s`\n\n正在获取当前状态...",
		SubFailed:   "订阅失败：%v",

		UnsubNeedArgs: "请提供追踪号，例如：\n`/unsub EB123456789CN`",
		Unsubscribed:  "已取消订阅：`%s`",
		UnsubFailed:   "取消订阅失败：%v",

		ListEmpty: "当前没有任何订阅。\n\n使用 `/sub 追踪号` 来添加订阅。",
		ListTitle: "当前订阅列表 (%d)：",

		CheckNeedArgs: "请提供追踪号，例如：\n`/check EB123456789CN`",
		Checking:      "正在查询 `%s` 的状态...",
		CheckFailed:   "查询失败：%v",

		PushStarting:      "正在立即检查所有订阅的更新...",
		PushDoneUpdates:   "手动推送完成\n\n检查了 %d 个订阅\n发现 %d 个更新\n其中 %d 个包裹已送达",
		PushDoneNoUpdates: "手动推送完成\n\n检查了 %d 个订阅\n暂无新更新",

		Welcome: "欢迎使用 EMS 追踪机器人！\n\n" +
			"我可以帮你追踪 EMS 快递，支持：\n" +
			"日本邮政（Japan Post）\n" +
			"中国EMS（via 17track）\n\n" +
			"CN结尾的单号会自动合并中日英三段物流信息。\n\n" +
			"请选择操作或使用下方命令：",

		BtnSubscribe:   "订阅追踪",
		BtnList:        "查看列表",
		BtnUnsubscribe: "取消订阅",
		BtnPush:        "立即推送",
		BtnHelp:        "帮助",

		PromptSub:   "请输入要订阅的追踪号：\n\n使用命令：`/sub 追踪号`\n例如：`/sub EB123456789CN`",
		PromptUnsub: "请输入要取消订阅的追踪号：\n\n使用命令：`/unsub 追踪号`",

		TrackingUpdate:   "追踪更新！",
		InitialStatus:    "初始状态",
		DeliveredEnd:     "包裹已送达！追踪结束。",
		DeliveredNoTrack: "包裹已送达，将不再继续追踪。",
		SubscriberFmt:    "订阅者: @%s (ID: %s)",
		PackageDelivered: "包裹已送达！",

		HelpTitle:         "帮助信息",
		HelpCmdSub:        "`/sub 追踪号` - 订阅追踪",
		HelpCmdUnsub:      "`/unsub 追踪号` - 取消订阅",
		HelpCmdList:       "`/list` - 查看订阅列表",
		HelpCmdCheck:      "`/check 追踪号` - 查询状态",
		HelpCmdPush:       "`/push` - 立即检查所有订阅",
		HelpCmdHelp:       "`/help` - 显示帮助",
		HelpFeatureAuto:   "定期自动检查更新",
		HelpFeatureNotify: "有变化时自动推送通知",
		HelpFeatureStop:   "包裹送达后自动停止追踪",
		HelpFeatureCN:     "CN结尾单号自动查询中国EMS",
		HelpFeatureMerge:  "中日英三语物流信息合并展示",
		HelpButtonTip:     "点击 /start 可以使用按钮菜单",
	}
}
