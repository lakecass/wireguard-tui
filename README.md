# WireGuard TUI

A beautiful terminal user interface for managing WireGuard VPN tunnels, inspired by `htop`.

![WireGuard TUI](docs/screenshot.png)

## Features

- **Htop-like Interface**: Clean, intuitive tree view with interfaces as roots and peers as leaves
- **Dynamic Layout**: Columns auto-resize to fill your terminal width
- **Visual Switches**: Color-coded `[ON]` (Green) / `[OFF]` (Red) indicators for tunnel status
- **Details Panel**: Bottom panel showing comprehensive information for selected items
- **Animated Mascot**: A friendly 2D companion that animates when tunnels are active
- **Multiple Themes**: 5 beautiful themes to choose from:
  - Htop Classic
  - Dracula
  - Solarized Light
  - Nord
  - Tokyo Night
- **Real-time Monitoring**: Live stats for handshakes and traffic
- **Toggle Control**: Press `Space` to bring interfaces up/down
- **Expand/Collapse**: Press `Enter` to show/hide peer details

## Installation

### Prerequisites

- Go 1.21 or higher
- WireGuard installed and configured
- Root/sudo access (required for managing WireGuard interfaces)

### Build from Source

```bash
git clone https://github.com/yourusername/wireguard-tui.git
cd wireguard-tui
go build -o wg-tui cmd/wireguard-tui/main.go
```

### Install

```bash
sudo mv wg-tui /usr/local/bin/
```

## Usage

### Running

```bash
# On Linux with WireGuard
sudo wg-tui

# Demo mode (for testing without WireGuard)
wg-tui -mock
```

### Keyboard Shortcuts

- `↑`/`↓` or `k`/`j` - Navigate
- `Space` - Toggle interface UP/DOWN
- `Enter` - Expand/Collapse interface
- `F2` - Cycle themes
- `F10` or `q` - Quit

## Configuration

WireGuard TUI reads your existing WireGuard configuration from `/etc/wireguard/*.conf`.

Example configuration structure:
```
/etc/wireguard/
├── wg0.conf
├── wg1.conf
└── wg2.conf
```

## How It Works

- **Scanning**: Automatically detects all WireGuard configurations in `/etc/wireguard/`
- **Status Detection**: Uses `wg show` to determine which interfaces are currently active
- **Toggle Control**: Uses `wg-quick up/down` to manage interface states
- **Real-time Stats**: Continuously updates peer statistics (handshake times, traffic)

## Development

### Running in Mock Mode

For development on systems without WireGuard:

```bash
go run cmd/wireguard-tui/main.go -mock
```

### Project Structure

```
wireguard-tui/
├── cmd/
│   └── wireguard-tui/
│       └── main.go          # Entry point
├── internal/
│   ├── ui/
│   │   ├── model.go         # Bubbletea model
│   │   ├── theme.go         # Color themes
│   │   └── mascot.go        # Animated mascot
│   └── wg/
│       ├── client.go        # WireGuard client interface
│       ├── linux.go         # Linux implementation
│       └── mock.go          # Mock implementation
├── go.mod
├── go.sum
└── README.md
```

## Requirements

- **Linux**: Primary target platform
- **WireGuard**: Must be installed and configured
- **Permissions**: Requires root/sudo to manage interfaces

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Inspired by [htop](https://htop.dev/)
