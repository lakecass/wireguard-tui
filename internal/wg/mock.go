package wg

import (
	"math/rand"
	"time"
)

// MockClient implements Client with fake data
type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) GetInterfaces() ([]Interface, error) {
	return []Interface{
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
	}, nil
}

func (m *MockClient) GetPeers(interfaceName string) ([]Peer, error) {
	if interfaceName == "wg2 (inactive)" {
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
	// In mock, we could update internal state, but for now just return nil
	return nil
}
