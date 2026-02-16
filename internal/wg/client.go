package wg

import (
	"time"
)

// InterfaceStatus represents the state of an interface
type InterfaceStatus int

const (
	InterfaceDown InterfaceStatus = iota
	InterfaceUp
)

func (s InterfaceStatus) String() string {
	switch s {
	case InterfaceUp:
		return "UP"
	default:
		return "DOWN"
	}
}

// Interface represents a WireGuard interface (e.g., wg0)
type Interface struct {
	Name         string
	PublicKey    string
	ListenPort   int
	FirewallMark int
	Status       InterfaceStatus
}

// Peer represents a connected peer
type Peer struct {
	PublicKey           string
	Endpoint            string
	AllowedIPs          []string
	LatestHandshake     time.Time
	TransferRx          int64
	TransferTx          int64
	PersistentKeepalive int
}

// Client defines the methods for interacting with WireGuard
type Client interface {
	GetInterfaces() ([]Interface, error)
	GetPeers(interfaceName string) ([]Peer, error)
	ToggleInterface(name string, up bool) error
}
