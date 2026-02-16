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
			m.toggleInterface()
			return m, m.refreshData
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

	// Create a footer base style that fills the background
	styleBar := lipgloss.NewStyle().Background(theme.DescBg)

	// Combine content
	fullFooter := footerContent

	// Calculate padding
	visibleLen := lipgloss.Width(footerContent)
	padding := width - visibleLen
	if padding > 0 {
		fullFooter += strings.Repeat(" ", padding)
	}

	// Render the whole block with the base style
	s += "\n" + styleBar.Render(fullFooter)

	return s

}

// Overlay string on the right side
func overlayRight(line, overlay string, width int) string {
	// This is tricky with ANSI codes in `line`.
	// For now, let's assume `line` has padding spaces at the end if we used .Width(width).
	// But styling adds ANSI codes at the end (reset).

	// Simpler approach:
	// Construct the line as:  [Content .....   Mascot]
	// If the content overlaps, Mascot takes precedence?

	// Since we are using full-width backgrounds, `line` is full of ANSI.
	// Let's use Lipgloss to render the mascot with a transparent background (or matching background)
	// and use `lipgloss.Place` concept or just manual spacing.

	// Manual:
	// 1. Remove last `len(overlay)` distinct characters (ignoring ANSI) ?? Hard.

	// 2. Just print mascot on a new layer? Bubbletea doesn't support layers easily in string view.

	// 3. Re-render the line.
	// If this is an empty filler line, it is easy.
	// If it has content, we might hide content.

	// Let's just overlay on empty lines for now (Mascot will sit in empty space).
	// If line is empty (just spaces), we replace end.

	// Quick hack:
	// Force the line to be Width - MascotWidth, then append Mascot.
	// But styles...

	// Let's just return line + "\n" + overlay? No.

	// Let's try to assume the line is padded with spaces and ANSI reset is at end.
	// Only apply mascot if line represents "Empty" space (filler).
	if strings.Contains(line, "          ") { // heuristics
		// It's likely empty-ish.
		// Use lipgloss to place.
		return lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(overlay)
	}

	return line
}

// Logic helpers

func (m Model) toggleInterface() {
	if m.cursor >= len(m.rows) {
		return
	}

	currentRow := m.rows[m.cursor]
	targetName := currentRow.InterfaceName

	// We need to find the current status of this interface.
	// Since we are in the UI model, the efficient way is to find the Interface Row in m.rows
	// (It should be there if we are viewing a peer of it).

	var iface wg.Interface
	found := false

	// If current row IS the interface, use it
	if currentRow.Type == RowInterface {
		iface = currentRow.Interface
		found = true
	} else {
		// Search for parent interface row
		for _, r := range m.rows {
			if r.Type == RowInterface && r.InterfaceName == targetName {
				iface = r.Interface
				found = true
				break
			}
		}
	}

	if found {
		newState := iface.Status == wg.InterfaceDown
		go func() {
			m.client.ToggleInterface(iface.Name, newState)
		}()
	}
}

func (m Model) flattenRows(baseRows []Row) []Row {
	var flat []Row
	for _, r := range baseRows {
		flat = append(flat, r)
		// Assuming baseRows has interfaces, and refreshData attaches peers??
		// Actually, refreshData implementation below attaches peers properly.
		// Wait, refreshData does NOT attach peers to `r.Children`.
		// It returns a flat list already?
		// CHECK: refreshData implementation below.
	}
	// My previous implementation of refreshData returned a flat list.
	// So `flattenRows` might be redundant or just a pass-through if we don't change the structure.
	// But `Update` calls `flattenRows`.
	// Let's ensure `refreshData` returns what we expect.
	// If refreshData returns flat list including expanded peers, then we just use it.
	// BUT, Update Logic for expansion re-flattens.
	// We need to re-implement `refreshData` to honor expansion or just return raw data and let Update flatten?
	// Best approach: `refreshData` returns only Interfaces (and maybe their peers in a struct), and we flatten in Update/View?
	// Or `refreshData` honors `Expanded` state?
	// The problem is `refreshData` runs periodically.

	// Let's stick to: Update receives []Row (new data).
	// We re-apply expansion state.
	// Then we filter out peers if interface is collapsed?

	// Let's refine `refreshData` to return everything, and `flattenRows` (or `filterRows`) hides collapsed peers.
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
	// Style for panel
	stylePanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColumnHeaderFg).
		Width(width - 2). // Account for border
		Height(height - 2)

	if m.cursor >= len(m.rows) {
		return stylePanel.Render("No selection")
	}

	row := m.rows[m.cursor]
	content := ""

	styleLabel := lipgloss.NewStyle().Foreground(theme.ColumnHeaderFg).Bold(true)
	styleValue := lipgloss.NewStyle().Foreground(theme.NormalFg)

	if row.Type == RowInterface {
		iface := row.Interface
		status := "UP"
		if iface.Status == wg.InterfaceDown {
			status = "DOWN"
		}

		content += fmt.Sprintf("Interface: %s\n", styleLabel.Render(iface.Name))
		content += fmt.Sprintf("Status:    %s\n", styleValue.Render(status))
		content += fmt.Sprintf("Public Key: %s\n", styleValue.Render(iface.PublicKey))
		content += fmt.Sprintf("Port:      %d\n", iface.ListenPort)
		content += fmt.Sprintf("FwMark:    %d\n", iface.FirewallMark)
	} else {
		peer := row.Peer
		content += fmt.Sprintf("Peer:      %s\n", styleLabel.Render(peer.PublicKey))
		content += fmt.Sprintf("Endpoint:  %s\n", styleValue.Render(peer.Endpoint))
		content += fmt.Sprintf("AllowedIPs: %v\n", peer.AllowedIPs)
		content += fmt.Sprintf("Transfer:  Rx: %s / Tx: %s\n", formatBytes(peer.TransferRx), formatBytes(peer.TransferTx))
		handshake := "Never"
		if !peer.LatestHandshake.IsZero() {
			handshake = time.Since(peer.LatestHandshake).String() + " ago"
		}
		content += fmt.Sprintf("Handshake: %s\n", styleValue.Render(handshake))
	}

	// Mascot Overlay in Details Panel?
	// The user asked for "fill the bottom" and mascot to be bottom right.
	// We can put the mascot inside the details panel on the right side.

	// Calculate mascot frame
	anyUp := false
	for _, r := range m.rows {
		if r.Type == RowInterface && r.Interface.Status == wg.InterfaceUp {
			anyUp = true
			break
		}
	}
	frame := GetMascotFrame(int(time.Now().UnixMilli()/200), anyUp)

	// Render Panel
	panelStr := stylePanel.Render(content)

	// Overlay Mascot on top of Panel string (bottom right corner)
	// Lipgloss border makes this tricky once rendered.
	// Easier to append mascot to content before render?
	// Or use Place.

	// Let's use `lipgloss.Place` to position mascot in the panel area.
	// Actually simple string overlay on the last line of content might be cleaner if we have space.

	// Let's stick to the "Global" footer area for mascot if Panel is full?
	// But the panel IS the footer area now.

	// Let's render the mascot as a separate block aligned right, and JoinHorizontal with content?
	// Content (Left) + Spacer + Mascot (Right)

	mascotStyle := lipgloss.NewStyle().Align(lipgloss.Right).Width(width - 4 - lipgloss.Width(content)) // rough calc
	_ = mascotStyle
	// Just place it manually in the string if possible.

	// For now, let's just return the panel. The mascot was "bottom right" of the screen.
	// The panel is at the bottom.
	// We can inject the mascot into the panel text.

	lines := strings.Split(panelStr, "\n")
	if len(lines) > 2 {
		// Inject into bottom-most content line (inside border)
		targetLine := len(lines) - 2 // -1 is bottom border
		lines[targetLine] = overlayRight(lines[targetLine], frame, width-2)
	}

	return strings.Join(lines, "\n")
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
