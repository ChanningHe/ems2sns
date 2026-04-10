package i18n

func japanese() *Messages {
	return &Messages{
		Unauthorized:      "このボットの使用権限がありません。",
		UnknownCmd:        "不明なコマンドです。/help で利用可能なコマンドを確認できます。",
		InvalidTrackingNo: "無効な追跡番号です。",

		SubNeedArgs: "追跡番号を入力してください。例：\n`/sub EB123456789CN`",
		Subscribed:  "`%s` を登録しました\n\n現在のステータスを取得中...",
		SubFailed:   "登録に失敗しました：%v",

		UnsubNeedArgs: "追跡番号を入力してください。例：\n`/unsub EB123456789CN`",
		Unsubscribed:  "`%s` の登録を解除しました。",
		UnsubFailed:   "登録解除に失敗しました：%v",

		ListEmpty: "登録中の追跡番号はありません。\n\n`/sub 追跡番号` で追加できます。",
		ListTitle: "登録リスト (%d)：",

		CheckNeedArgs: "追跡番号を入力してください。例：\n`/check EB123456789CN`",
		Checking:      "`%s` のステータスを確認中...",
		CheckFailed:   "確認に失敗しました：%v",

		PushStarting:      "全ての登録を確認中...",
		PushDoneUpdates:   "手動プッシュ完了\n\n確認: %d 件\n更新: %d 件\n配達済み: %d 件",
		PushDoneNoUpdates: "手動プッシュ完了\n\n確認: %d 件\n新しい更新はありません。",

		Welcome: "EMS追跡ボットへようこそ！\n\n" +
			"EMS荷物の追跡ができます：\n" +
			"日本郵便（日本語/英語）\n" +
			"中国EMS（17track経由）\n\n" +
			"CN末尾の番号は日中英の追跡情報を自動統合します。\n\n" +
			"操作を選択するか、以下のコマンドを使用してください：",

		BtnSubscribe:   "登録",
		BtnList:        "一覧",
		BtnUnsubscribe: "解除",
		BtnPush:        "即時確認",
		BtnHelp:        "ヘルプ",

		PromptSub:   "登録する追跡番号を入力してください：\n\nコマンド：`/sub 追跡番号`\n例：`/sub EB123456789CN`",
		PromptUnsub: "解除する追跡番号を入力してください：\n\nコマンド：`/unsub 追跡番号`",

		TrackingUpdate:   "追跡更新！",
		InitialStatus:    "初期ステータス",
		DeliveredEnd:     "配達完了！追跡を終了します。",
		DeliveredNoTrack: "配達済みです。これ以上の追跡は行いません。",
		SubscriberFmt:    "登録者: @%s (ID: %s)",
		PackageDelivered: "配達完了！",

		HelpTitle:         "ヘルプ",
		HelpCmdSub:        "`/sub 番号` - 追跡を登録",
		HelpCmdUnsub:      "`/unsub 番号` - 登録解除",
		HelpCmdList:       "`/list` - 登録一覧",
		HelpCmdCheck:      "`/check 番号` - ステータス確認",
		HelpCmdPush:       "`/push` - 全件即時確認",
		HelpCmdHelp:       "`/help` - ヘルプ表示",
		HelpFeatureAuto:   "定期的に自動確認",
		HelpFeatureNotify: "ステータス変更時に自動通知",
		HelpFeatureStop:   "配達後に自動停止",
		HelpFeatureCN:     "CN末尾番号は中国EMSを自動照会",
		HelpFeatureMerge:  "日中英の追跡情報を統合表示",
		HelpButtonTip:     "/start でボタンメニューを表示",
	}
}
