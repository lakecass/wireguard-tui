package wg

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LinuxClient implements Client using the `wg` command line tool
type LinuxClient struct{}

func NewLinuxClient() *LinuxClient {
	return &LinuxClient{}
}

func (c *LinuxClient) GetInterfaces() ([]Interface, error) {
	// 1. Get active interfaces from wg show
	output, err := runWgDump()
	if err != nil {
		// If wg fails (e.g. no permission), we might still want to list config files?
		// For now, return error if we can't run wg (likely need sudo)
		return nil, err
	}
	activeInterfaces := parseInterfaces(output)

	// Mark all parsed interfaces as UP (they came from `wg show`, so they are running)
	for i := range activeInterfaces {
		activeInterfaces[i].Status = InterfaceUp
	}

	// Build a set of active interface names for quick lookup
	seen := make(map[string]bool)
	var allInterfaces []Interface

	// Add active interfaces first
	for _, iface := range activeInterfaces {
		allInterfaces = append(allInterfaces, iface)
		seen[iface.Name] = true
	}

	// 2. Scan /etc/wireguard/*.conf for all available configs
	configFiles, _ := filepath.Glob("/etc/wireguard/*.conf")

	// Add inactive interfaces from config files
	for _, file := range configFiles {
		name := strings.TrimSuffix(filepath.Base(file), ".conf")
		if !seen[name] {
			allInterfaces = append(allInterfaces, Interface{
				Name:   name,
				Status: InterfaceDown,
			})
			seen[name] = true
		}
	}

	return allInterfaces, nil
}

func (c *LinuxClient) GetPeers(interfaceName string) ([]Peer, error) {
	output, err := runWgDump()
	if err != nil {
		return nil, err
	}
	// Note: If interface is down, `wg show` won't return peers.
	// We could parse the conf file, but typically peers are only relevant when UP.
	return parsePeers(output, interfaceName), nil
}

func (c *LinuxClient) ToggleInterface(name string, up bool) error {
	var action string
	if up {
		action = "up"
	} else {
		action = "down"
	}

	cmd := exec.Command("wg-quick", action, name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg-quick failed: %v, output: %s", err, string(output))
	}
	return nil
}

// Helper to run `wg show all dump`
func runWgDump() (string, error) {
	cmd := exec.Command("wg", "show", "all", "dump")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run wg show: %v", err)
	}
	return out.String(), nil
}

func parseInterfaces(output string) []Interface {
	var interfaces []Interface
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 1 {
			continue
		}

		if len(parts) == 5 {
			// Interface line
			name := parts[0]
			if seen[name] {
				continue
			}

			port, _ := strconv.Atoi(parts[3])
			fwMark, _ := strconv.Atoi(parts[4])

			interfaces = append(interfaces, Interface{
				Name:         name,
				PublicKey:    parts[2],
				ListenPort:   port,
				FirewallMark: fwMark,
			})
			seen[name] = true
		} else if len(parts) > 5 {
			// Peer line, check if we missed the interface
			name := parts[0]
			if !seen[name] {
				interfaces = append(interfaces, Interface{Name: name})
				seen[name] = true
			}
		}
	}
	return interfaces
}

func parsePeers(output string, interfaceName string) []Peer {
	var peers []Peer
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue
		}

		if parts[0] != interfaceName {
			continue
		}

		handshakeInt, _ := strconv.ParseInt(parts[5], 10, 64)
		rx, _ := strconv.ParseInt(parts[6], 10, 64)
		tx, _ := strconv.ParseInt(parts[7], 10, 64)
		keepalive, _ := strconv.Atoi(parts[8])

		var handshakeTime time.Time
		if handshakeInt > 0 {
			handshakeTime = time.Unix(handshakeInt, 0)
		}

		peers = append(peers, Peer{
			PublicKey:           parts[1],
			Endpoint:            parts[3],
			AllowedIPs:          strings.Split(parts[4], ","),
			LatestHandshake:     handshakeTime,
			TransferRx:          rx,
			TransferTx:          tx,
			PersistentKeepalive: keepalive,
		})
	}
	return peers
}
