package templates

// LoginData renders the sign-in card. Errors is keyed by field name with
// "form" reserved for top-level messages.
type LoginData struct {
	Title      string
	Theme      string
	CSRF       string
	PanelTitle string
	PanelText  string
	FlashType  string
	FlashMsg   string
	Email      string
	Errors     map[string]string
	DemoHint   bool
	DemoAdmin  string
	DemoUser   string
}

type RegisterData struct {
	Title      string
	Theme      string
	CSRF       string
	PanelTitle string
	PanelText  string
	FlashType  string
	FlashMsg   string
	Name       string
	Email      string
	Errors     map[string]string
}

type ForgotData struct {
	Title      string
	Theme      string
	CSRF       string
	PanelTitle string
	PanelText  string
	FlashType  string
	FlashMsg   string
	Email      string
	Errors     map[string]string
}

type ResetData struct {
	Title  string
	Theme  string
	CSRF   string
	Token  string
	Errors map[string]string
	Valid  bool // false renders the "link expired" state
}
