package ui

import (
	"fmt"
	"strings"
	"time"

	"wireguard-tui/internal/wg"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type dataMsg struct {
	interfaces []wg.Interface
	peers      map[string][]wg.Peer
}

type Model struct {
	client     wg.Client
	interfaces []wg.Interface
	peers      map[string][]wg.Peer
	cursor     int
	width      int
	height     int
	err        error
	tick       time.Duration
	themeIndex int
	showHelp   bool
	showFilter bool
	filterText string
}

func NewModel(client wg.Client) Model {
	return Model{
		client:     client,
		tick:       time.Second,
		themeIndex: 0,
		peers:      make(map[string][]wg.Peer),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.refreshData, m.tickCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showFilter {
			switch msg.String() {
			case "esc", "enter":
				m.showFilter = false
			case "backspace":
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.filterText += msg.String()
				}
			}
			return m, nil
		}

		if m.showHelp {
			if msg.String() != "" {
				m.showHelp = false
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "f10":
			return m, tea.Quit
		case "f1", "?":
			m.showHelp = !m.showHelp
		case "f2":
			m.themeIndex = (m.themeIndex + 1) % len(Themes)
		case "f5", "r":
			m.err = nil
			return m, m.refreshData
		case "f6", "/":
			m.showFilter = true
			m.filterText = ""
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.getFilteredCount()-1 {
				m.cursor++
			}
		case " ":
			filtered := m.getFilteredInterfaces()
			if m.cursor < len(filtered) {
				iface := filtered[m.cursor]
				newState := iface.Status == wg.InterfaceDown
				return m, m.toggleCmd(iface.Name, newState)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		return m, tea.Batch(m.refreshData, m.tickCmd())
	case dataMsg:
		m.interfaces = msg.interfaces
		m.peers = msg.peers
		if m.cursor >= len(m.interfaces) {
			m.cursor = len(m.interfaces) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	case error:
		m.err = msg
	}
	return m, nil
}

func (m Model) View() string {
	theme := Themes[m.themeIndex]
	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 24
	}

	// Styles
	sHeader := lipgloss.NewStyle().Foreground(theme.HeaderFg).Background(theme.HeaderBg).Bold(true)
	sColHdr := lipgloss.NewStyle().Foreground(theme.ColumnHeaderFg).Background(theme.ColumnHeaderBg).Bold(true)
	sSel := lipgloss.NewStyle().Foreground(theme.SelectedFg).Background(theme.SelectedBg)
	sNorm := lipgloss.NewStyle().Foreground(theme.NormalFg)
	sKey := lipgloss.NewStyle().Foreground(theme.KeyFg).Background(theme.KeyBg).Bold(true).Padding(0, 1)
	sDesc := lipgloss.NewStyle().Foreground(theme.DescFg).Background(theme.DescBg).Padding(0, 0)
	sError := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	sDim := lipgloss.NewStyle().Foreground(theme.DimFg)

	// Item Styles for Robust Alignment
	wName := 12
	wStatus := 8
	wPort := 7
	wPeers := 10
	wTransfer := 22
	wActive := width - wName - wStatus - wPort - wPeers - wTransfer - 2
	if wActive < 8 {
		wActive = 8
	}

	stName := lipgloss.NewStyle().Width(wName)
	stStatus := lipgloss.NewStyle().Width(wStatus).PaddingRight(1)
	stPort := lipgloss.NewStyle().Width(wPort)
	stPeers := lipgloss.NewStyle().Width(wPeers)
	stTrans := lipgloss.NewStyle().Width(wTransfer)
	stActive := lipgloss.NewStyle().Width(wActive)

	// 1. Header
	headerText := fmt.Sprintf(" WireGuard TUI (%s) ", theme.Name)
	clock := time.Now().Format("15:04:05")
	padLen := width - lipgloss.Width(headerText) - len(clock)
	if padLen < 0 {
		padLen = 0
	}
	header := sHeader.Render(fmt.Sprintf("%s%*s%s", headerText, padLen, "", clock))

	// Error Line (if any)
	errorLine := ""
	if m.err != nil {
		errorLine = sError.Render(fmt.Sprintf(" Error: %v", m.err))
		if lipgloss.Width(errorLine) < width {
			errorLine += strings.Repeat(" ", width-lipgloss.Width(errorLine))
		}
		errorLine = sHeader.Background(lipgloss.Color("0")).Render(errorLine) + "\n"
	}

	// 2. Column Headers
	colHeader := sColHdr.Render(lipgloss.JoinHorizontal(lipgloss.Top,
		stName.Render("Interface"),
		stStatus.Render("Status"),
		stPort.Render("Port"),
		stPeers.Render("Peers"),
		stTrans.Render("Transfer (Total)"),
		stActive.Render("Active (Latest)"),
	))
	if wh := lipgloss.Width(colHeader); wh < width {
		colHeader += sColHdr.Render(strings.Repeat(" ", width-wh))
	}

	// 3. Interface list (Window 1)
	detailsHeight := height / 2
	if detailsHeight < 10 {
		detailsHeight = 10
	}
	listHeight := height - 3 - detailsHeight
	if m.err != nil {
		listHeight--
	}
	if listHeight < 3 {
		listHeight = 3
	}

	filtered := m.getFilteredInterfaces()
	startRow := 0
	if m.cursor >= startRow+listHeight {
		startRow = m.cursor - listHeight + 1
	}
	endRow := startRow + listHeight
	if endRow > len(filtered) {
		endRow = len(filtered)
	}

	onSty := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	offSty := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)

	var bodyRows []string
	for i := startRow; i < endRow; i++ {
		iface := filtered[i]
		statusStr := offSty.Render("[OFF]")
		if iface.Status == wg.InterfaceUp {
			statusStr = onSty.Render("[ON]")
		}

		var totalRx, totalTx int64
		var latestHS time.Time
		for _, p := range m.peers[iface.Name] {
			totalRx += p.TransferRx
			totalTx += p.TransferTx
			if p.LatestHandshake.After(latestHS) {
				latestHS = p.LatestHandshake
			}
		}

		transferStr := "-"
		activeStr := "-"
		if iface.Status == wg.InterfaceUp {
			if totalRx > 0 || totalTx > 0 {
				transferStr = fmt.Sprintf("Rx:%s Tx:%s", formatBytes(totalRx), formatBytes(totalTx))
			}
			if !latestHS.IsZero() {
				activeStr = fmtDur(time.Since(latestHS))
			}
		}

		pc := len(m.peers[iface.Name])
		peersStr := "-"
		if iface.Status == wg.InterfaceUp && pc > 0 {
			peersStr = fmt.Sprintf("%d peers", pc)
		}

		portStr := "-"
		if iface.ListenPort > 0 {
			portStr = fmt.Sprintf("%d", iface.ListenPort)
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			stName.Render(truncate(iface.Name, wName-1)),
			stStatus.Render(statusStr),
			stPort.Render(truncate(portStr, wPort-1)),
			stPeers.Render(truncate(peersStr, wPeers-1)),
			stTrans.Render(truncate(transferStr, wTransfer-1)),
			stActive.Render(truncate(activeStr, wActive-1)),
		)

		if i == m.cursor {
			rowWidth := lipgloss.Width(row)
			if rowWidth < width {
				row += strings.Repeat(" ", width-rowWidth)
			}
			bodyRows = append(bodyRows, sSel.Render(row))
		} else {
			bodyRows = append(bodyRows, sNorm.Width(width).Render(row))
		}
	}
	for len(bodyRows) < listHeight {
		bodyRows = append(bodyRows, sNorm.Width(width).Render(""))
	}

	// 4. Details Panel (Window 2)
	// We pass the filtered interface if selected
	details := ""
	if len(filtered) > 0 && m.cursor < len(filtered) {
		details = m.renderDetailsPanelFor(filtered[m.cursor], width, detailsHeight, theme)
	} else {
		details = m.renderDetailsPanel(width, detailsHeight, theme)
	}

	mainView := header + "\n" + errorLine + colHeader + "\n" + strings.Join(bodyRows, "\n") + "\n" + details

	// 5. Footer / Filter Bar
	footerView := ""
	if m.showFilter {
		fBar := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("0")).Bold(true)
		prompt := " Filter: "
		footerView = fBar.Render(prompt + m.filterText + strings.Repeat(" ", width-lipgloss.Width(prompt+m.filterText)))
	} else {
		footerItems := []string{
			sKey.Render("F1") + sDesc.Render("Help"),
			sKey.Render("F2") + sDesc.Render("Theme"),
			sKey.Render("F5") + sDesc.Render("Refresh"),
			sKey.Render("F6") + sDesc.Render("Filter"),
			sKey.Render("Space") + sDesc.Render("Toggle"),
			sKey.Render("F10") + sDesc.Render("Quit"),
		}
		fc := strings.Join(footerItems, " ")
		vl := lipgloss.Width(fc)
		if vl > width {
			fc = strings.Join(footerItems, "")
			vl = lipgloss.Width(fc)
		}
		ff := fc
		if pad := width - vl; pad > 0 {
			ff += strings.Repeat(" ", pad)
		}
		footerView = sDesc.Background(theme.DescBg).Render(ff)
	}

	s := mainView + "\n" + footerView

	// 6. Help Overlay
	if m.showHelp {
		helpBox := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(theme.KeyBg).
			Padding(1, 2).
			Background(theme.ColumnHeaderBg).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					sKey.Render("F1 / ?")+" Show this help",
					sKey.Render("F2")+" Cycle color themes",
					sKey.Render("F5 / R")+" Refresh interface status",
					sKey.Render("F6 / /")+" Search / Filter interfaces",
					sKey.Render("Space")+" Toggle Interface (UP/DOWN)",
					sKey.Render("Arrows / J,K")+" Navigate list",
					sKey.Render("F10 / Q")+" Quit Application",
					"",
					sDim.Render(" Produced by lakecass and Gemini"),
					"",
					sDesc.Render("Press any key to close"),
				),
			)
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, helpBox)
	}

	return s
}

func (m Model) getFilteredInterfaces() []wg.Interface {
	if m.filterText == "" {
		return m.interfaces
	}
	var filtered []wg.Interface
	for _, iface := range m.interfaces {
		if strings.Contains(strings.ToLower(iface.Name), strings.ToLower(m.filterText)) {
			filtered = append(filtered, iface)
		}
	}
	return filtered
}

func (m Model) getFilteredCount() int {
	return len(m.getFilteredInterfaces())
}

func (m Model) renderDetailsPanelFor(iface wg.Interface, width, height int, theme Theme) string {
	sPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColumnHeaderFg).
		Padding(0, 1).
		Width(width - 2).
		Height(height - 2)

	// Almost identical to renderDetailsPanel but uses passed iface
	sLabel := lipgloss.NewStyle().Foreground(theme.ColumnHeaderFg).Bold(true)
	sValue := lipgloss.NewStyle().Foreground(theme.NormalFg)
	sDim := lipgloss.NewStyle().Foreground(theme.DimFg)
	sAccent := lipgloss.NewStyle().Foreground(theme.HeaderBg).Bold(true)

	var b strings.Builder
	status := "UP"
	if iface.Status == wg.InterfaceDown {
		status = "DOWN"
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		sLabel.Render("Interface: "), sAccent.Render(iface.Name),
		strings.Repeat(" ", 4),
		sLabel.Render("Status: "), sValue.Render(status),
	) + "\n")

	pk := iface.PublicKey
	if pk == "" {
		pk = "N/A"
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		sLabel.Render("Public Key: "), sValue.Render(truncate(pk, 16)),
		strings.Repeat(" ", 2),
		sLabel.Render("Port: "), sValue.Render(fmt.Sprintf("%d", iface.ListenPort)),
		strings.Repeat(" ", 2),
		sLabel.Render("FwMark: "), sValue.Render(fmt.Sprintf("%d", iface.FirewallMark)),
	) + "\n")

	peers := m.peers[iface.Name]
	if len(peers) == 0 {
		if iface.Status == wg.InterfaceDown {
			b.WriteString(sDim.Render("\nInterface is DOWN — no peer data"))
		} else {
			b.WriteString(sDim.Render("\nNo peers configured"))
		}
	} else {
		b.WriteString(fmt.Sprintf("\n%s (%d):\n", sLabel.Render("Peers"), len(peers)))
		iw := width - 4
		pK, pE, pI, pT := 11, 20, 14, 21
		pH := iw - pK - pE - pI - pT
		stK, stE, stI, stT, stH := lipgloss.NewStyle().Width(pK), lipgloss.NewStyle().Width(pE), lipgloss.NewStyle().Width(pI), lipgloss.NewStyle().Width(pT), lipgloss.NewStyle().Width(pH)

		hdr := lipgloss.JoinHorizontal(lipgloss.Top, stK.Render("Key"), stE.Render("Endpoint"), stI.Render("Allowed IPs"), stT.Render("Transfer"), stH.Render("Handshake"))
		b.WriteString(sDim.Render(truncate(hdr, iw)) + "\n")

		for _, p := range peers {
			tx := fmt.Sprintf("Rx:%s Tx:%s", formatBytes(p.TransferRx), formatBytes(p.TransferTx))
			hs := "Never"
			if !p.LatestHandshake.IsZero() {
				hs = fmtDur(time.Since(p.LatestHandshake))
			}
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				stK.Render(truncate(p.PublicKey, pK-2)), stE.Render(truncate(p.Endpoint, pE-1)), stI.Render(truncate(strings.Join(p.AllowedIPs, ","), pI-1)), stT.Render(truncate(tx, pT-1)), stH.Render(truncate(hs, pH-1)))
			b.WriteString(row + "\n")
		}
	}

	anyUp := false
	for _, ifc := range m.interfaces {
		if ifc.Status == wg.InterfaceUp {
			anyUp = true
			break
		}
	}
	frame := GetMascotFrame(int(time.Now().UnixMilli()/200), anyUp)
	mscSty := lipgloss.NewStyle().Width(width - 4).Align(lipgloss.Right).PaddingTop(1)
	b.WriteString(mscSty.Render(frame))

	return sPanel.Render(b.String())
}

func (m Model) renderDetailsPanel(width, height int, theme Theme) string {
	if len(m.interfaces) > 0 && m.cursor < len(m.interfaces) {
		return m.renderDetailsPanelFor(m.interfaces[m.cursor], width, height, theme)
	}
	sPanel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.ColumnHeaderFg).Padding(0, 1).Width(width - 2).Height(height - 2)
	return sPanel.Render("No interface selected")
}

func (m Model) toggleCmd(name string, up bool) tea.Cmd {
	return func() tea.Msg {
		err := m.client.ToggleInterface(name, up)
		if err != nil {
			return err
		}
		return m.refreshData()
	}
}

func (m Model) refreshData() tea.Msg {
	ifaces, err := m.client.GetInterfaces()
	if err != nil {
		return err
	}
	peers := make(map[string][]wg.Peer)
	for _, iface := range ifaces {
		if iface.Status == wg.InterfaceUp {
			p, _ := m.client.GetPeers(iface.Name)
			peers[iface.Name] = p
		}
	}
	return dataMsg{interfaces: ifaces, peers: peers}
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func truncate(s string, maxLen int) string {
	if maxLen <= 2 {
		if len(s) > maxLen && maxLen > 0 {
			return s[:maxLen]
		}
		return s
	}
	if len(s) > maxLen {
		return s[:maxLen-2] + ".."
	}
	return s
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func fmtDur(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
