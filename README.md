# WireGuard TUI (Htop Classic)

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/lakecass/wireguard-tui)](https://goreportcard.com/report/github.com/lakecass/wireguard-tui)

**WireGuard TUI** is a modern, terminal-based dashboard for managing and monitoring WireGuard interfaces. Inspired by the legendary `htop`, it provides high-density real-time analytics, aggregated traffic statistics, and intuitive management tools in an aesthetically pleasing package.

![Main Dashboard](assets/dashboard_main.png)

## ✨ Features

- **📊 High-Density Dashboard**: View all your interfaces at a glance with 6-column analytics (Interface, Status, Port, Peers, Total Transfer, and Activity).
- **📉 Real-time Data Aggregation**: Automatically merges traffic (Rx/Tx) and handshake data from all peers to show interface-level performance.
- **🎨 Multi-Theme Support**: Includes premium themes like **Dracula**, **Nord**, **Tokyo Night**, and **Solarized Light**.
- **🔍 Advanced Filtering**: Lightning-fast live search for managing dozens of tunnels.
- **⌨️ Intuitive Keybindings**: Control your entire network without leaving the keyboard.
- **🛡️ Robust Error Handling**: Non-intrusive status reporting for backend issues (permissions, missing tools, etc.).
- **⚡ Built with Go**: Blazing-fast performance with zero external dependencies (aside from `wireguard-tools`).

## 🖼️ Screenshots

````carousel
![Tokyo Night Theme](assets/theme_tokyo_night.png)
<!-- slide -->
![Solarized Light](assets/dashboard_solarized.png)
<!-- slide -->
![Help Overlay](assets/help_menu.png)
<!-- slide -->
![Alternative View](assets/dashboard_alt.png)
````

## 🚀 Installation

### Debian / Ubuntu
Download the latest `.deb` from the [releases page](https://github.com/lakecass/wireguard-tui/releases) and install via dpkg:
```bash
sudo dpkg -i wireguard-tui_0.1.0_amd64.deb
```

### Arch Linux
You can build from the provided `PKGBUILD` in the `packaging/` directory:
```bash
cd packaging/arch
makepkg -si
```

### From Source
```bash
make build
sudo cp wireguard-tui /usr/bin/
```

## 🎮 Usage

Simply run the command with `sudo` (required for `wg show` interactions):
```bash
sudo wireguard-tui
```

### Keyboard Shortcuts
| Key | Action |
| --- | --- |
| `F1` / `?` | Show Help & Credits |
| `F2` | Cycle Color Themes |
| `F5` / `R` | Manual Data Refresh |
| `F6` / `/` | Search & Filter Interfaces |
| `Space` | Toggle Interface (UP/DOWN) |
| `Arrows` / `J,K` | Navigate List |
| `F10` / `Q` | Exit Application |

## 🛠️ Requirements
- Linux Kernel with WireGuard support
- `wireguard-tools` (provides the `wg` command)

## 🤝 Produced by
Produced with love by **lakecass** and **Gemini**.

## 📄 License
This project is licensed under the **GPL-3.0 License**. See the [LICENSE](LICENSE) file for details.
