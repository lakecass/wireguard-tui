package wg

import (
	"fmt"
	"math/rand"
	"time"
)

type MockClient struct {
	Interfaces []Interface
	Peers      map[string][]Peer
}

func NewMockClient() *MockClient {
	// Initialize with some dummy data
	ifaces := []Interface{
		{Name: "wg0", PublicKey: "OHkMwM9QyK9f9...", ListenPort: 51820, FirewallMark: 0, Status: InterfaceUp},
		{Name: "wg1", PublicKey: "AbCmK1239...", ListenPort: 51821, FirewallMark: 0, Status: InterfaceUp},
		{Name: "wg2", PublicKey: "InAcTiVe...", ListenPort: 0, FirewallMark: 0, Status: InterfaceDown},
	}

	peers := make(map[string][]Peer)
	peers["wg0"] = []Peer{
		{PublicKey: "PeEr1...", Endpoint: "192.168.1.10:51820", AllowedIPs: []string{"10.0.0.2/32"}, LatestHandshake: time.Now().Add(-49 * time.Second), TransferRx: 896432, TransferTx: 408123, PersistentKeepalive: 25},
		{PublicKey: "PeEr2...", Endpoint: "203.0.113.5:12345", AllowedIPs: []string{"10.0.0.3/32"}, LatestHandshake: time.Now().Add(-48 * time.Second), TransferRx: 78902, TransferTx: 3242634, PersistentKeepalive: 25},
		{PublicKey: "PeEr3 (Ina...", Endpoint: "Unknown", AllowedIPs: []string{"10.0.0.4/32"}, LatestHandshake: time.Now().Add(-48 * time.Hour), TransferRx: 1024, TransferTx: 2048, PersistentKeepalive: 0},
	}
	peers["wg1"] = []Peer{
		{PublicKey: "PeEr1...", Endpoint: "192.168.1.10:51820", AllowedIPs: []string{"192.168.2.2/32"}, LatestHandshake: time.Now().Add(-2 * time.Minute), TransferRx: 1324354, TransferTx: 216123, PersistentKeepalive: 25},
		{PublicKey: "PeEr2...", Endpoint: "203.0.113.5:12345", AllowedIPs: []string{"192.168.2.3/32"}, LatestHandshake: time.Now().Add(-18 * time.Second), TransferRx: 85000, TransferTx: 1350000, PersistentKeepalive: 25},
		{PublicKey: "PeEr3 (Ina...", Endpoint: "Unknown", AllowedIPs: []string{"192.168.2.4/32"}, LatestHandshake: time.Now().Add(-48 * time.Hour), TransferRx: 1024, TransferTx: 2048, PersistentKeepalive: 0},
	}

	return &MockClient{
		Interfaces: ifaces,
		Peers:      peers,
	}
}

func (c *MockClient) GetInterfaces() ([]Interface, error) {
	// Simulate delay?
	// time.Sleep(100 * time.Millisecond)
	return c.Interfaces, nil
}

func (c *MockClient) GetPeers(interfaceName string) ([]Peer, error) {
	if p, ok := c.Peers[interfaceName]; ok {
		// Randomize some data for liveness
		for i := range p {
			p[i].TransferRx += int64(rand.Intn(1024))
			p[i].TransferTx += int64(rand.Intn(1024))
			if rand.Intn(10) > 8 {
				p[i].LatestHandshake = time.Now()
			}
		}
		return p, nil
	}
	return nil, nil
}

func (c *MockClient) ToggleInterface(name string, up bool) error {
	// Find interface and update status
	for i, iface := range c.Interfaces {
		if iface.Name == name {
			if up {
				c.Interfaces[i].Status = InterfaceUp
			} else {
				c.Interfaces[i].Status = InterfaceDown
			}
			return nil
		}
	}
	return fmt.Errorf("interface not found")
}
