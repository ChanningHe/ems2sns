package i18n

import "fmt"

type Messages struct {
	// Auth
	Unauthorized string

	// Command errors
	UnknownCmd        string
	InvalidTrackingNo string

	// Subscribe
	SubNeedArgs string
	Subscribed  string // %s = tracking number
	SubFailed   string // %v = error

	// Unsubscribe
	UnsubNeedArgs string
	Unsubscribed  string // %s = tracking number
	UnsubFailed   string // %v = error

	// List
	ListEmpty string
	ListTitle string // %d = count

	// Check
	CheckNeedArgs string
	Checking      string // %s = tracking number
	CheckFailed   string // %v = error

	// Push
	PushStarting      string
	PushDoneUpdates   string // %d, %d, %d = checked, updates, delivered
	PushDoneNoUpdates string // %d = checked

	// Welcome / Start
	Welcome string

	// Telegram inline buttons
	BtnSubscribe   string
	BtnList        string
	BtnUnsubscribe string
	BtnPush        string
	BtnHelp        string

	// Telegram callback prompts
	PromptSub   string
	PromptUnsub string

	// Tracking update notifications
	TrackingUpdate   string
	InitialStatus    string
	DeliveredEnd     string
	DeliveredNoTrack string
	SubscriberFmt    string // %s, %s = username, userID
	PackageDelivered string

	// Help sections (assembled by each platform's formatter)
	HelpTitle         string
	HelpCmdSub        string
	HelpCmdUnsub      string
	HelpCmdList       string
	HelpCmdCheck      string
	HelpCmdPush       string
	HelpCmdHelp       string
	HelpFeatureAuto   string
	HelpFeatureNotify string
	HelpFeatureStop   string
	HelpFeatureCN     string
	HelpFeatureMerge  string
	HelpButtonTip     string
}

func Load(lang string) *Messages {
	switch lang {
	case "zh":
		return chinese()
	case "ja":
		return japanese()
	default:
		return english()
	}
}

func Supported() []string {
	return []string{"en", "zh", "ja"}
}

func Validate(lang string) error {
	for _, l := range Supported() {
		if lang == l {
			return nil
		}
	}
	return fmt.Errorf("unsupported language %q, supported: en, zh, ja", lang)
}
