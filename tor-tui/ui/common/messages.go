package common

import (
	"time"

	"github.com/amirk1998/tor-manager/core/proxy"
	"github.com/amirk1998/tor-manager/core/tor"
)

// TickMsg is sent on a regular interval to trigger IP auto-refresh.
type TickMsg time.Time

// IPCheckResultMsg carries the result of a proxy IP check.
type IPCheckResultMsg struct {
	Result proxy.CheckResult
}

// NewNymResultMsg carries the result of a NEWNYM request.
type NewNymResultMsg struct {
	Err error
}

// BootstrapEventMsg carries a Tor bootstrap progress event.
type BootstrapEventMsg struct {
	Event tor.BootstrapEvent
}

// BridgesFetchedMsg carries bridges received from BridgeDB.
type BridgesFetchedMsg struct {
	Bridges []tor.Bridge
	Err     error
}

// BridgeAppliedMsg signals that a bridge config was applied.
type BridgeAppliedMsg struct {
	Err error
}

// ExitCountryAppliedMsg signals that an exit country was applied.
type ExitCountryAppliedMsg struct {
	Err error
}

// LogLineMsg carries a new log line from the Tor process.
type LogLineMsg struct {
	Line string
}

// ControllerReadyMsg signals the controller connected successfully.
type ControllerReadyMsg struct {
	Err error
}

// ToastMsg is a short ephemeral notification shown in the footer.
type ToastMsg struct {
	Message string
	IsError bool
}

// ClipboardMsg requests the proxy address be placed on clipboard.
type ClipboardMsg struct {
	Text string
}
