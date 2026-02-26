// Package tor provides a high-level interface for managing a Tor daemon,
// including control port communication, bridge management, and configuration.
package tor

import "errors"

// Sentinel errors for the tor package.
var (
	ErrNotConnected      = errors.New("tor: not connected to control port")
	ErrAuthFailed        = errors.New("tor: authentication failed")
	ErrControlPortClosed = errors.New("tor: control port connection closed")
	ErrInvalidResponse   = errors.New("tor: invalid response from control port")
	ErrTorNotRunning     = errors.New("tor: daemon is not running")
	ErrBridgeInvalid     = errors.New("tor: invalid bridge line")
	ErrNoBridgesReturned = errors.New("tor: BridgeDB returned no bridges")
	ErrOperationTimeout  = errors.New("tor: operation timed out")
	ErrPermissionDenied  = errors.New("tor: permission denied (check control port password)")
)
