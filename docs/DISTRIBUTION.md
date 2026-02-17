# WireGuard TUI 手动分发指南

既然你手头有一台树莓派（DietPi / Debian 系）和可能的 Arch Linux 环境，下面是手动为其打包和安装的详细步骤。

---

## 1. Debian / 树莓派 (ARM64 & x86-64)

在你的树莓派 (192.168.124.102) 上，你可以通过以下两种方式安装。

### 方案 A：快速“一键”安装（推荐）
如果你只是想直接用起来，可以直接利用我在 `dist/` 目录下为你编译好的二进制文件：

1. **从 Mac 发送文件到树莓派**:
   ```bash
   scp dist/wireguard-tui-linux-arm64 dietpi@192.168.124.102:/tmp/
   ```
2. **在树莓派上移动到系统目录**:
   ```bash
   ssh dietpi@192.168.124.102
   sudo mv /tmp/wireguard-tui-linux-arm64 /usr/bin/wireguard-tui
   sudo chmod +x /usr/bin/wireguard-tui
   ```
3. **运行**:
   ```bash
   sudo wireguard-tui
   ```

### 方案 B：手动制作 `.deb` 安装包
如果你想给别人发一安装包，或者想用 `apt` 管理：

1. **在树莓派上克隆项目**:
   ```bash
   git clone https://github.com/lakecass/wireguard-tui.git
   cd wireguard-tui
   ```
2. **运行我为你准备的脚本**:
   ```bash
   # 首先确保你有 arm64 的二进制文件在 dist 目录下
   # 或者直接在树莓派上编译：
   make build
   mkdir -p dist/
   cp wireguard-tui dist/wireguard-tui-linux-arm64
   
   # 执行打包脚本
   chmod +x scripts/package-deb.sh
   ./scripts/package-deb.sh
   ```
3. **安装生成的包**:
   ```bash
   sudo dpkg -i dist/wireguard-tui_0.1.0_arm64.deb
   ```

---

## 2. Arch Linux (x86-64)

Arch Linux 使用 `PKGBUILD` 进行打包，这是最标准的方式。

1. **进入打包目录**:
   ```bash
   cd packaging/arch
   ```
2. **执行构建命令**:
   `makepkg` 会自动读取 `PKGBUILD`，下载源码、编译并打包。
   ```bash
   makepkg -si
   ```
   - `-s`: 自动安装缺失的依赖（如 `go`）。
   - `-i`: 打包完成后自动安装到系统。

---

## 3. 常见问题 (FAQ)

- **为什么需要 `sudo`？**
  WireGuard 的核心信息（密钥、流量等）存储在内核中，只有 root 权限才能通过 `wg show` 命令读取并展示到 TUI 上。
- **依赖工具**:
  请确保你的系统已经安装了 `wireguard-tools`。
  - Debian: `sudo apt install wireguard-tools`
  - Arch: `sudo pacman -S wireguard-tools`

---
**祝你在树莓派上使用愉快！** 🚀
**Produced by lakecass and Gemini**
