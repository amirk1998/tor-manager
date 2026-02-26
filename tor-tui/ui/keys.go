package ui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeys are available in every view.
type GlobalKeyMap struct {
	Tab      key.Binding
	ShiftTab key.Binding
	Quit     key.Binding
	Help     key.Binding

	// Tab shortcuts
	GoStatus    key.Binding
	GoBridges   key.Binding
	GoCountries key.Binding
	GoLogs      key.Binding
	GoSettings  key.Binding
}

// ViewKeys are navigation keys used within list/table views.
type ViewKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	PageUp key.Binding
	PageDn key.Binding
	Top    key.Binding
	Bottom key.Binding
	Select key.Binding
	Back   key.Binding
}

// StatusKeys are actions available on the dashboard.
type StatusKeyMap struct {
	Refresh     key.Binding
	NewIdentity key.Binding
	CopyProxy   key.Binding
}

// BridgeKeys are actions on the bridge manager view.
type BridgeKeyMap struct {
	ModeBuiltIn key.Binding
	ModeCustom  key.Binding
	RequestNew  key.Binding
	Apply       key.Binding
	Delete      key.Binding
	Edit        key.Binding
}

// Global is the shared global keymap instance.
var Global = GlobalKeyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	GoStatus: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "dashboard"),
	),
	GoBridges: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "bridges"),
	),
	GoCountries: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "countries"),
	),
	GoLogs: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "logs"),
	),
	GoSettings: key.NewBinding(
		key.WithKeys("5"),
		key.WithHelp("5", "settings"),
	),
}

var NavKeys = ViewKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+u"),
		key.WithHelp("pgup", "page up"),
	),
	PageDn: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+d"),
		key.WithHelp("pgdn", "page down"),
	),
	Top: key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G", "end"),
		key.WithHelp("G", "bottom"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

var StatusKeys = StatusKeyMap{
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh IP"),
	),
	NewIdentity: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new identity"),
	),
	CopyProxy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy proxy addr"),
	),
}

var BridgeKeys = BridgeKeyMap{
	ModeBuiltIn: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "built-in bridges"),
	),
	ModeCustom: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "custom bridge"),
	),
	RequestNew: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fetch from BridgeDB"),
	),
	Apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "delete"),
		key.WithHelp("d", "delete"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
}
