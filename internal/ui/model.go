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

// RowType distinguishes between Interface rows and Peer rows
type RowType int

const (
	RowInterface RowType = iota
	RowPeer
)

type Row struct {
	Type          RowType
	InterfaceName string
	Peer          wg.Peer
	// For interface rows
	Interface wg.Interface
	Expanded  bool
}

type Model struct {
	client     wg.Client
	rows       []Row
	cursor     int
	width      int
	height     int
	err        error
	tick       time.Duration
	themeIndex int
}

func NewModel(client wg.Client) Model {
	return Model{
		client:     client,
		tick:       time.Second,
		themeIndex: 0, // Default to first theme
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshData,
		m.tickCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "f10":
			return m, tea.Quit
		case "f2":
			// Cycle theme
			m.themeIndex = (m.themeIndex + 1) % len(Themes)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		case "space":
			// Toggle interface UP/DOWN
			return m, m.toggleInterface()
		case "enter":
			// Toggle expansion for interfaces
			if m.cursor < len(m.rows) && m.rows[m.cursor].Type == RowInterface {
				m.rows[m.cursor].Expanded = !m.rows[m.cursor].Expanded
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		return m, tea.Batch(
			m.refreshData,
			m.tickCmd(),
		)
	case []Row:
		newRows := msg
		// Map old state (expansion)
		expansionState := make(map[string]bool)
		for _, r := range m.rows {
			if r.Type == RowInterface {
				expansionState[r.InterfaceName] = r.Expanded
			}
		}

		// Apply state
		for i := range newRows {
			if newRows[i].Type == RowInterface {
				if expanded, ok := expansionState[newRows[i].InterfaceName]; ok {
					newRows[i].Expanded = expanded
				} else {
					// Default expanded for active interfaces?
					if newRows[i].Interface.Status == wg.InterfaceUp {
						newRows[i].Expanded = true
					}
				}
			}
		}
		m.rows = m.flattenRows(newRows)

		// Adjust cursor
		if m.cursor >= len(m.rows) {
			m.cursor = len(m.rows) - 1
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
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	theme := Themes[m.themeIndex]
	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 24
	}

	// Calculate column widths based on available width
	// Total = 100%
	// Name: 30%, Endpoint: 35%, Transfer: 20%, Handshake: 15% (Approx)

	// Minimum widths
	minName := 20
	minEnd := 20
	minTran := 15
	minHand := 12

	// Remaining width distributed
	avail := width - minName - minEnd - minTran - minHand
	if avail < 0 {
		avail = 0
	}

	// Dynamic allocation
	wName := minName + int(float64(avail)*0.3)
	wEnd := minEnd + int(float64(avail)*0.4)
	wTran := minTran + int(float64(avail)*0.2)
	wHand := width - wName - wEnd - wTran // Rest

	// Styles
	styleHeader := lipgloss.NewStyle().
		Foreground(theme.HeaderFg).
		Background(theme.HeaderBg).
		Bold(true)

	styleColHeader := lipgloss.NewStyle().
		Foreground(theme.ColumnHeaderFg).
		Background(theme.ColumnHeaderBg).
		Bold(true)

	styleSelected := lipgloss.NewStyle().
		Foreground(theme.SelectedFg).
		Background(theme.SelectedBg)

	styleNormal := lipgloss.NewStyle().
		Foreground(theme.NormalFg)

	styleKey := lipgloss.NewStyle().
		Foreground(theme.KeyFg).
		Background(theme.KeyBg).
		Bold(true).
		Padding(0, 1)

	styleDesc := lipgloss.NewStyle().
		Foreground(theme.DescFg).
		Background(theme.DescBg).
		Padding(0, 1)

	// 1. Header
	headerText := fmt.Sprintf(" WireGuard TUI (%s) ", theme.Name)
	clock := time.Now().Format("15:04:05")
	padLen := width - lipgloss.Width(headerText) - len(clock)
	if padLen < 0 {
		padLen = 0
	}
	// Manual padding instead of .Width() to ensure visibility
	headerStr := fmt.Sprintf("%s%*s%s", headerText, padLen, "", clock)
	header := styleHeader.Render(headerStr)

	// 2. Column Headers
	colRow := fmt.Sprintf("%-*s%-*s%-*s%-*s",
		wName, "Name/Key",
		wEnd, "Endpoint",
		wTran, "Transfer",
		wHand, "Handshake")

	// Ensure background fills full width
	colHeader := styleColHeader.Render(colRow)

	// 3. Body (Split View)
	// Top: List
	// Bottom: Details (8 lines)

	// Calculate heights
	detailsHeight := 10                      // Fixed height for details panel
	listHeight := height - 3 - detailsHeight // Header(1) + ColHeader(1) + Footer(1)
	if listHeight < 4 {
		listHeight = 4
	} // Minimum list height

	// Scroll offset
	rowsToShow := listHeight
	startRow := 0
	if m.cursor >= rowsToShow {
		startRow = m.cursor - rowsToShow + 1
	}
	endRow := startRow + rowsToShow
	if endRow > len(m.rows) {
		endRow = len(m.rows)
	}

	var bodyRows []string

	for i := startRow; i < endRow; i++ {
		row := m.rows[i]
		var nameStr, endStr, tranStr, handStr string

		if row.Type == RowInterface {
			// Interface Row with Switch
			statusSymbol := "[-]"
			if !row.Expanded {
				statusSymbol = "[+]"
			}

			// Visual Switch with colors
			switchStr := "[ON ]"
			// Create styles per row to avoid race/global issues
			onStyle := styleNormal.Copy().Foreground(lipgloss.Color("2")).Bold(true)  // Green
			offStyle := styleNormal.Copy().Foreground(lipgloss.Color("1")).Bold(true) // Red

			statusColored := onStyle.Render(switchStr)
			if row.Interface.Status == wg.InterfaceDown {
				switchStr = "[OFF]"
				statusColored = offStyle.Render(switchStr)
				statusSymbol = "[x]"
			}

			// Name with switch
			// We need to calc length without ansi for alignment manually if we use Sprintf
			// Or just simple concat
			nameStr = fmt.Sprintf("%s %s %s", statusSymbol, row.InterfaceName, statusColored)
			endStr = row.Interface.PublicKey[:8] + "..."
		} else {
			treePrefix := "  |-"
			pubKey := row.Peer.PublicKey
			if len(pubKey) > 10 {
				pubKey = pubKey[:10] + "..."
			}
			nameStr = treePrefix + " " + pubKey
			endStr = row.Peer.Endpoint
			tranStr = fmt.Sprintf("%s/%s", formatBytes(row.Peer.TransferRx), formatBytes(row.Peer.TransferTx))
			if !row.Peer.LatestHandshake.IsZero() {
				handStr = time.Since(row.Peer.LatestHandshake).Round(time.Second).String()
			} else {
				handStr = "Never"
			}
		}

		// Truncation needs to happen on raw strings before color, or handle ansi width
		// Name column is tricky due to color.

		// Use lipgloss to measure visible width of nameStr
		visibleNameLen := lipgloss.Width(nameStr)
		paddingName := wName - visibleNameLen
		if paddingName < 0 {
			paddingName = 0
		}

		// Construct the name column with correct visual width
		// We append spaces manually instead of using %-*s with incorrect length
		nameCol := nameStr + strings.Repeat(" ", paddingName)

		line := fmt.Sprintf("%s%-*s%-*s%-*s",
			nameCol,
			wEnd, truncate(endStr, wEnd-1),
			wTran, truncate(tranStr, wTran-1),
			wHand, truncate(handStr, wHand-1),
		)

		// Correct approach for ANSI width: Use lipgloss to force width?
		// But we have mixed styles in one cell.
		// Let's rely on standard formatting, but remove the "truncate" call for nameStr containing ANSI.
		// We trust the interface name is short enough.

		if i == m.cursor {
			bodyRows = append(bodyRows, styleSelected.Width(width).Render(line))
		} else {
			if row.Type == RowInterface {
				// Use ColumnHeaderBg for Interface rows to separate them visually
				// Also Bold
				interfaceStyle := styleNormal.Copy().
					Bold(true).
					Background(theme.ColumnHeaderBg).
					Foreground(theme.ColumnHeaderFg).
					Width(width)

				bodyRows = append(bodyRows, interfaceStyle.Render(line))
			} else {
				bodyRows = append(bodyRows, styleNormal.Width(width).Render(line))
			}
		}
	}

	// Fill list empty rows
	for len(bodyRows) < rowsToShow {
		bodyRows = append(bodyRows, styleNormal.Width(width).Render(""))
	}

	// 4. Details Panel (Bottom)
	details := m.renderDetailsPanel(width, detailsHeight, theme)

	// Join all parts
	// Header + ColHeader + Body + Details + Footer
	// Note: Footer is redundant if we have Details Panel?
	// The user said "F10 ... added features".
	// Let's keep footer below Details? Or integrated?
	// User said "fill the bottom".

	s := header + "\n" + colHeader + "\n" + strings.Join(bodyRows, "\n") + "\n" + details

	// 5. Footer (Overlay / Append)
	footerItems := []string{
		styleKey.Render("F1") + styleDesc.Render("Help"),
		styleKey.Render("F2") + styleDesc.Render("Theme"),
		styleKey.Render("Space") + styleDesc.Render("Toggle"),
		styleKey.Render("Enter") + styleDesc.Render("Expand"),
		styleKey.Render("F10") + styleDesc.Render("Quit"),
	}
	footerContent := strings.Join(footerItems, "")

	// Create a footer style that fills the background
	styleFooter := lipgloss.NewStyle().
		Background(theme.DescBg) // Match Description background

	// Pad footer to fill width
	footerLen := lipgloss.Width(footerContent)
	footerPad := width - footerLen
	if footerPad > 0 {
		footerContent += strings.Repeat(" ", footerPad)
	}

	s += "\n" + styleFooter.Render(footerContent)

	return s

}


// Logic helpers

func (m Model) toggleInterface() tea.Cmd {
	if m.cursor < len(m.rows) && m.rows[m.cursor].Type == RowInterface {
		iface := m.rows[m.cursor].Interface
		newState := iface.Status == wg.InterfaceDown

		return func() tea.Msg {
			// Perform toggle synchronously in the Cmd
			_ = m.client.ToggleInterface(iface.Name, newState)
			// Return a message to trigger refresh?
			// Or just call refreshData logic here?
			// refreshData returns tea.Msg (which is []Row or error).
			// So we can just call it (since it's a method on Model that returns Msg, but wait, refreshData receives (m Model)... pure function?)

			// m.refreshData() returns tea.Msg
			// So we can just return that.
			return m.refreshData()
		}
	}
	return nil
}

func (m Model) flattenRows(baseRows []Row) []Row {
	var flat []Row
	for _, r := range baseRows {
		flat = append(flat, r)
	}
	// ... (rest logic same, just truncating for context)
	return filterCollapsed(flat)
}

func filterCollapsed(rows []Row) []Row {
	var filtered []Row
	var skipping bool
	for _, r := range rows {
		if r.Type == RowInterface {
			filtered = append(filtered, r)
			skipping = !r.Expanded
		} else {
			if !skipping {
				filtered = append(filtered, r)
			}
		}
	}
	return filtered
}

// Commands

func (m Model) refreshData() tea.Msg {
	ifaces, err := m.client.GetInterfaces()
	if err != nil {
		return err
	}

	var rows []Row
	for _, iface := range ifaces {
		r := Row{
			Type:          RowInterface,
			InterfaceName: iface.Name,
			Interface:     iface,
			Expanded:      iface.Status == wg.InterfaceUp,
		}
		rows = append(rows, r)

		if iface.Status == wg.InterfaceUp { // Or always if we want to show cached?
			peers, _ := m.client.GetPeers(iface.Name)
			for _, p := range peers {
				rows = append(rows, Row{
					Type:          RowPeer,
					InterfaceName: iface.Name,
					Peer:          p,
				})
			}
		}
	}
	return rows
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) renderDetailsPanel(width, height int, theme Theme) string {
	// Calculate dimensions
	panelWidth := width - 2
	panelHeight := height - 2

	// Style for panel
	// Use DescBg to unify with footer as requested
	stylePanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColumnHeaderFg).
		Background(theme.DescBg).
		Width(panelWidth).
		Height(panelHeight)

	if m.cursor >= len(m.rows) {
		return stylePanel.Render("No selection")
	}

	row := m.rows[m.cursor]
	contentLines := []string{}

	// Label/Value styles - adjust foreground to be visible on DescBg
	// Use DescFg (which is designed for DescBg)
	styleLabel := lipgloss.NewStyle().Foreground(theme.ColumnHeaderFg).Bold(true).Background(theme.DescBg)
	styleValue := lipgloss.NewStyle().Foreground(theme.DescFg).Background(theme.DescBg)

	if row.Type == RowInterface {
		iface := row.Interface
		status := "UP"
		if iface.Status == wg.InterfaceDown {
			status = "DOWN"
		}

		contentLines = append(contentLines, fmt.Sprintf("Interface: %s", styleLabel.Render(iface.Name)))
		contentLines = append(contentLines, fmt.Sprintf("Status:    %s", styleValue.Render(status)))
		contentLines = append(contentLines, fmt.Sprintf("Public Key: %s", styleValue.Render(iface.PublicKey)))
		contentLines = append(contentLines, fmt.Sprintf("Port:      %d", iface.ListenPort)) // Raw int, minimal style
		contentLines = append(contentLines, fmt.Sprintf("FwMark:    %d", iface.FirewallMark))
	} else {
		peer := row.Peer
		contentLines = append(contentLines, fmt.Sprintf("Peer:      %s", styleLabel.Render(peer.PublicKey)))
		contentLines = append(contentLines, fmt.Sprintf("Endpoint:  %s", styleValue.Render(peer.Endpoint)))
		contentLines = append(contentLines, fmt.Sprintf("AllowedIPs: %v", peer.AllowedIPs))
		contentLines = append(contentLines, fmt.Sprintf("Transfer:  Rx: %s / Tx: %s", formatBytes(peer.TransferRx), formatBytes(peer.TransferTx)))
		handshake := "Never"
		if !peer.LatestHandshake.IsZero() {
			handshake = time.Since(peer.LatestHandshake).String() + " ago"
		}
		contentLines = append(contentLines, fmt.Sprintf("Handshake: %s", styleValue.Render(handshake)))
	}

	// Calculate mascot frame
	anyUp := false
	for _, r := range m.rows {
		if r.Type == RowInterface && r.Interface.Status == wg.InterfaceUp {
			anyUp = true
			break
		}
	}
	frame := GetMascotFrame(int(time.Now().UnixMilli()/200), anyUp)

	// Combine content and mascot
	// We manually pad content lines to width, and on the last few lines we insert mascot

	// Create final content string
	var finalContent string

	// Helper to pad line
	padLine := func(s string, w int) string {
		// ANSI aware length?
		// Since we are inside the panel, lipgloss handles the panel width.
		// Use lipgloss Place?
		// Simple approach: Render content left, Mascot right.
		return lipgloss.NewStyle().Width(w).Background(theme.DescBg).Render(s)
	}

	// We simply join lines?
	// But we want mascot at bottom right.
	// Inject mascot into empty lines at the bottom?

	// Fill remaining lines with empty strings
	for len(contentLines) < panelHeight {
		contentLines = append(contentLines, "")
	}

	// Inject mascot into the last line (or last N lines if multi-line mascot)
	// Frame is single line string?
	lastIdx := panelHeight - 1
	if lastIdx >= 0 && lastIdx < len(contentLines) {
		// Construct the last line: Content (Left) + Space + Mascot (Right)

		// Use lipgloss to layout
		// PlaceHorizontal
		combined := lipgloss.NewStyle().Width(panelWidth).Background(theme.DescBg).
			Render(
				lipgloss.PlaceHorizontal(panelWidth, lipgloss.Right, frame,
					lipgloss.WithWhitespaceChars(" "),
					lipgloss.WithWhitespaceForeground(theme.DescFg), // matching fg?
				),
			)

		// Wait, PlaceHorizontal REPLACES content? No, it places content in width.
		// Currently `leftContent` is essentially empty string for the last line usually.
		// If it's not empty, we might overwrite.
		// Assuming last line is empty for now.
		contentLines[lastIdx] = combined
	}

	// Apply padding to all other lines to ensure background fill
	for i := 0; i < len(contentLines); i++ {
		if i != lastIdx { // lastIdx already handled
			contentLines[i] = padLine(contentLines[i], panelWidth)
		}
	}

	finalContent = strings.Join(contentLines, "\n")
	return stylePanel.Render(finalContent)
}

func truncate(s string, maxLen int) string {
	if maxLen <= 3 {
		if len(s) > maxLen {
			if maxLen > 0 {
				return s[:maxLen]
			}
			return ""
		}
		return s
	}
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// Utils
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %c", float64(bytes)/float64(div), "KMGTPE"[exp])
}
