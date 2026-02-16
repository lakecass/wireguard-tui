package wg

import (
	"math/rand"
	"sync"
	"time"
)

// MockClient implements Client with fake data
type MockClient struct {
	mu         sync.Mutex
	interfaces []Interface
}

func NewMockClient() *MockClient {
	return &MockClient{
		interfaces: []Interface{
			{
				Name:       "wg0",
				PublicKey:  "OHkMwM9QyK9f9...",
				ListenPort: 51820,
				Status:     InterfaceUp,
			},
			{
				Name:       "wg1 (vpn)",
				PublicKey:  "AbCmK1239...",
				ListenPort: 51821,
				Status:     InterfaceUp,
			},
			{
				Name:      "wg2 (inactive)",
				PublicKey: "InAcTiVe...",
				Status:    InterfaceDown,
			},
		},
	}
}

func (m *MockClient) GetInterfaces() ([]Interface, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy
	dummy := make([]Interface, len(m.interfaces))
	copy(dummy, m.interfaces)
	return dummy, nil
}

func (m *MockClient) GetPeers(interfaceName string) ([]Peer, error) {
	// Check if interface is up?
	// In real world, if interface is down, maybe no peers?
	// But let's verify interface status first?
	// Getting status requires looking up.

	// For mock, let's just return peers if it's not "wg2 (inactive)"
	// Or better, check internal state.
	m.mu.Lock()
	var iface *Interface
	for i := range m.interfaces {
		if m.interfaces[i].Name == interfaceName {
			iface = &m.interfaces[i]
			break
		}
	}
	m.mu.Unlock()

	if iface == nil || iface.Status == InterfaceDown {
		return nil, nil
	}

	// Simulate some random traffic updates
	peers := []Peer{
		{
			PublicKey:       "PeEr1...",
			Endpoint:        "192.168.1.10:51820",
			AllowedIPs:      []string{"10.0.0.2/32"},
			LatestHandshake: time.Now().Add(-time.Duration(rand.Intn(120)) * time.Second),
			TransferRx:      int64(rand.Intn(1000000) + 500000),
			TransferTx:      int64(rand.Intn(500000) + 100000),
		},
		{
			PublicKey:       "PeEr2...",
			Endpoint:        "203.0.113.5:12345",
			AllowedIPs:      []string{"10.0.0.3/32", "10.0.0.4/32"},
			LatestHandshake: time.Now().Add(-time.Duration(rand.Intn(300)) * time.Second),
			TransferRx:      int64(rand.Intn(200000)),
			TransferTx:      int64(rand.Intn(9000000)),
		},
		{
			PublicKey:       "PeEr3 (Inactive)...",
			Endpoint:        "Unknown",
			AllowedIPs:      []string{"10.0.0.5/32"},
			LatestHandshake: time.Now().Add(-48 * time.Hour), // Old handshake
			TransferRx:      1024,
			TransferTx:      2048,
		},
	}
	return peers, nil
}

func (m *MockClient) ToggleInterface(name string, up bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.interfaces {
		if m.interfaces[i].Name == name {
			if up {
				m.interfaces[i].Status = InterfaceUp
			} else {
				m.interfaces[i].Status = InterfaceDown
			}
			return nil
		}
	}
	return nil
}
