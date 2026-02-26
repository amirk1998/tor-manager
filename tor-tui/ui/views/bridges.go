package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/you/tor-manager/core/tor"
	"github.com/you/tor-tui/ui"
)

type bridgeMode int

const (
	modeBridgeList  bridgeMode = iota // browsing built-in list
	modeCustomInput                   // typing a custom bridge line
	modeFetching                      // waiting for BridgeDB
)

// BridgesView manages bridge selection and configuration.
type BridgesView struct {
	width  int
	height int

	mode     bridgeMode
	cursor   int
	selected int // index in activeTransport's built-in list (-1 = custom)

	activeTransport tor.TransportType
	customBridges   []tor.Bridge // user-added custom bridges
	customInput     textinput.Model
	customError     string

	fetchedBridges []tor.Bridge
	fetching       bool
	spinner        spinner.Model

	toast    string
	toastErr bool

	ctrl *tor.Controller
}

// transportOrder is the display order of transports in the selector.
var transportOrder = []tor.TransportType{
	tor.TransportObfs4,
	tor.TransportSnowflake,
	tor.TransportWebTunnel,
	tor.TransportMeekAzure,
	tor.TransportDirect,
}

func NewBridgesView(ctrl *tor.Controller) *BridgesView {
	ti := textinput.New()
	ti.Placeholder = "obfs4 1.2.3.4:1234 FINGERPRINT cert=... iat-mode=0"
	ti.CharLimit = 512
	ti.Width = 60

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = ui.StylePrimary

	return &BridgesView{
		ctrl:            ctrl,
		activeTransport: tor.TransportObfs4,
		customInput:     ti,
		spinner:         sp,
		selected:        -1,
	}
}

func (b *BridgesView) SetSize(w, h int) {
	b.width = w
	b.height = h
}

func (b *BridgesView) Init() tea.Cmd {
	return b.spinner.Tick
}

func (b *BridgesView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		b.spinner, cmd = b.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case ui.BridgesFetchedMsg:
		b.fetching = false
		if msg.Err != nil {
			b.toast = "BridgeDB error: " + msg.Err.Error()
			b.toastErr = true
		} else {
			b.fetchedBridges = msg.Bridges
			b.toast = fmt.Sprintf("Fetched %d bridge(s) from BridgeDB", len(msg.Bridges))
			b.toastErr = false
		}

	case ui.BridgeAppliedMsg:
		if msg.Err != nil {
			b.toast = "Apply failed: " + msg.Err.Error()
			b.toastErr = true
		} else {
			b.toast = "Bridge configuration applied!"
			b.toastErr = false
		}

	case tea.KeyMsg:
		switch b.mode {

		case modeBridgeList:
			switch {
			case msg.String() == "up" || msg.String() == "k":
				b.moveCursor(-1)
			case msg.String() == "down" || msg.String() == "j":
				b.moveCursor(1)
			case msg.String() == "left" || msg.String() == "h":
				b.prevTransport()
			case msg.String() == "right" || msg.String() == "l":
				b.nextTransport()
			case msg.String() == "enter" || msg.String() == " ":
				b.selectCurrent()
			case msg.String() == "u":
				b.mode = modeCustomInput
				b.customInput.Focus()
				cmds = append(cmds, textinput.Blink)
			case msg.String() == "f":
				cmds = append(cmds, b.doFetchBridges())
			case msg.String() == "a":
				cmds = append(cmds, b.doApplyBridge())
			case msg.String() == "d":
				b.deleteSelected()
			case msg.String() == "x":
				b.clearBridges()
			}

		case modeCustomInput:
			switch msg.String() {
			case "esc":
				b.mode = modeBridgeList
				b.customInput.Blur()
				b.customError = ""
			case "enter":
				if err := b.submitCustomBridge(); err != nil {
					b.customError = err.Error()
				} else {
					b.mode = modeBridgeList
					b.customInput.Blur()
					b.customError = ""
				}
			default:
				var cmd tea.Cmd
				b.customInput, cmd = b.customInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return b, tea.Batch(cmds...)
}

func (b *BridgesView) View() string {
	if b.width == 0 {
		return ""
	}
	w := b.width - 4
	if w > 78 {
		w = 78
	}

	sections := []string{
		b.renderTransportSelector(w),
		b.renderBridgeList(w),
	}

	if b.mode == modeCustomInput {
		sections = append(sections, b.renderCustomInput(w))
	} else {
		sections = append(sections, b.renderFetchedBridges(w))
	}

	sections = append(sections, b.renderActions(w))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// ── Sub-renderers ─────────────────────────────────────────────────────────

func (b *BridgesView) renderTransportSelector(w int) string {
	var tabs []string
	for _, tt := range transportOrder {
		meta, ok := tor.Transports[tt]
		if !ok {
			continue
		}
		name := meta.Name
		if tt == tor.TransportDirect {
			name = "No Bridge"
		} else {
			// short names for tabs
			parts := strings.Fields(meta.Name)
			name = parts[0]
		}

		if tt == b.activeTransport {
			tabs = append(tabs, ui.StyleTabActive.Render(" "+name+" "))
		} else {
			tabs = append(tabs, ui.StyleTabInactive.Render(" "+name+" "))
		}
	}

	tabRow := strings.Join(tabs, ui.StyleDim.Render("│"))

	meta := tor.Transports[b.activeTransport]
	desc := ui.StyleDim.Render(meta.Description)

	badges := ""
	if b.activeTransport != tor.TransportDirect {
		badges = "  " + ui.CensorshipBadge(meta.CensorshipResistance) +
			"  " + ui.SpeedBadge(meta.SpeedImpact)
	}

	hint := ui.StyleDim.Render("  ← → to switch transport")

	content := tabRow + "\n" + desc + badges + "\n" + hint
	return ui.StylePanel.Width(w).Render(content)
}

func (b *BridgesView) renderBridgeList(w int) string {
	bridges := b.currentBuiltIns()
	if len(bridges) == 0 && b.activeTransport == tor.TransportDirect {
		content := ui.SectionTitle("Built-in Bridges") + "\n\n" +
			ui.StyleDim.Render("  Direct connection — no bridge in use.\n  Your IP is visible to your ISP.\n")
		return ui.StylePanel.Width(w).Render(content)
	}

	var rows []string
	rows = append(rows, ui.SectionTitle(fmt.Sprintf("Built-in Bridges (%d)", len(bridges))))
	rows = append(rows, "")

	colW := w - 10
	header := fmt.Sprintf("  %s  %-*s  %s",
		ui.StyleTableHeader.Render("   "),
		colW,
		ui.StyleTableHeader.Render("Bridge Address"),
		ui.StyleTableHeader.Render("Fingerprint"),
	)
	rows = append(rows, header)
	rows = append(rows, "  "+ui.Divider(w-6))

	for i, br := range bridges {
		cursor := "  "
		rowStyle := ui.StyleRowNormal
		if i == b.cursor {
			cursor = ui.StyleCursor.Render(" ▶")
			rowStyle = ui.StyleRowSelected
		}
		if i == b.selected {
			cursor = ui.StyleSuccessBold.Render(" ✓")
		}

		// Display truncated address + partial fingerprint
		addr := br.Address
		fp := ""
		if br.Fingerprint != "" {
			fp = br.Fingerprint[:8] + "..."
		}

		addr = truncate(addr, colW)
		row := fmt.Sprintf("%s %-*s  %s", cursor, colW, addr, ui.StyleDim.Render(fp))
		rows = append(rows, rowStyle.Render(row))
	}

	// Custom bridges section
	if len(b.customBridges) > 0 {
		rows = append(rows, "")
		rows = append(rows, ui.SectionTitle(fmt.Sprintf("Custom Bridges (%d)", len(b.customBridges))))
		rows = append(rows, "  "+ui.Divider(w-6))
		for _, br := range b.customBridges {
			rows = append(rows, ui.StyleAccent.Render("  + ")+ui.StyleDim.Render(truncate(br.String(), w-8)))
		}
	}

	content := strings.Join(rows, "\n")
	return ui.StylePanel.Width(w).Render(content)
}

func (b *BridgesView) renderCustomInput(w int) string {
	label := ui.StyleInputLabel.Render("Paste a bridge line:")
	input := b.customInput.View()

	errLine := ""
	if b.customError != "" {
		errLine = "\n" + ui.StyleError.Render("  ✗ "+b.customError)
	}

	hint := ui.StyleDim.Render("\n  [enter] add   [esc] cancel")

	content := label + "\n\n  " + input + errLine + hint
	return ui.StylePanelActive.Width(w).Render(content)
}

func (b *BridgesView) renderFetchedBridges(w int) string {
	var content string
	title := ui.SectionTitle("BridgeDB — Fresh Bridges")

	if b.fetching {
		content = title + "\n\n  " + b.spinner.View() + " " +
			ui.StyleDim.Render("Contacting bridges.torproject.org...")
	} else if len(b.fetchedBridges) == 0 {
		content = title + "\n\n  " +
			ui.StyleDim.Render("Press [f] to request fresh bridges for the selected transport.")
	} else {
		var rows []string
		rows = append(rows, title)
		rows = append(rows, "")
		for _, br := range b.fetchedBridges {
			rows = append(rows, ui.StyleAccent.Render("  + ")+
				ui.StyleDim.Render(truncate(br.String(), w-8)))
		}
		rows = append(rows, "")
		rows = append(rows, ui.StyleDim.Render("  Press [a] to apply these bridges."))
		content = strings.Join(rows, "\n")
	}

	return ui.StylePanel.Width(w).Render(content)
}

func (b *BridgesView) renderActions(w int) string {
	var hints []string

	if b.mode == modeBridgeList {
		hints = []string{
			ui.KeyHint("↑↓", "navigate"),
			ui.KeyHint("←→", "transport"),
			ui.KeyHint("enter", "select"),
			ui.KeyHint("u", "add custom"),
			ui.KeyHint("f", "fetch new"),
			ui.KeyHint("a", "apply"),
			ui.KeyHint("d", "remove custom"),
			ui.KeyHint("x", "disable bridges"),
		}
	}

	line := strings.Join(hints, "  ")

	if b.toast != "" {
		toastStyle := ui.StyleSuccess
		prefix := "✓ "
		if b.toastErr {
			toastStyle = ui.StyleError
			prefix = "✗ "
		}
		return line + "\n" + toastStyle.Render(prefix+b.toast)
	}
	return line
}

// ── Logic ─────────────────────────────────────────────────────────────────

func (b *BridgesView) currentBuiltIns() []tor.Bridge {
	return tor.BuiltInBridgeEntries(b.activeTransport)
}

func (b *BridgesView) moveCursor(delta int) {
	bridges := b.currentBuiltIns()
	if len(bridges) == 0 {
		return
	}
	b.cursor += delta
	if b.cursor < 0 {
		b.cursor = len(bridges) - 1
	}
	if b.cursor >= len(bridges) {
		b.cursor = 0
	}
}

func (b *BridgesView) selectCurrent() {
	bridges := b.currentBuiltIns()
	if b.cursor >= 0 && b.cursor < len(bridges) {
		b.selected = b.cursor
		b.toast = "Selected: " + bridges[b.cursor].Address
		b.toastErr = false
	}
}

func (b *BridgesView) nextTransport() {
	for i, tt := range transportOrder {
		if tt == b.activeTransport {
			b.activeTransport = transportOrder[(i+1)%len(transportOrder)]
			b.cursor = 0
			b.selected = -1
			return
		}
	}
}

func (b *BridgesView) prevTransport() {
	for i, tt := range transportOrder {
		if tt == b.activeTransport {
			idx := (i - 1 + len(transportOrder)) % len(transportOrder)
			b.activeTransport = transportOrder[idx]
			b.cursor = 0
			b.selected = -1
			return
		}
	}
}

func (b *BridgesView) submitCustomBridge() error {
	raw := strings.TrimSpace(b.customInput.Value())
	if raw == "" {
		return fmt.Errorf("bridge line is empty")
	}
	br, err := tor.ParseBridgeLine(raw)
	if err != nil {
		return err
	}
	b.customBridges = append(b.customBridges, br)
	b.customInput.Reset()
	b.toast = "Custom bridge added"
	b.toastErr = false
	return nil
}

func (b *BridgesView) deleteSelected() {
	if len(b.customBridges) == 0 {
		return
	}
	if len(b.customBridges) > 0 {
		b.customBridges = b.customBridges[:len(b.customBridges)-1]
		b.toast = "Last custom bridge removed"
		b.toastErr = false
	}
}

func (b *BridgesView) clearBridges() {
	b.selected = -1
	b.customBridges = nil
	b.fetchedBridges = nil
	b.activeTransport = tor.TransportDirect
	b.toast = "Bridges disabled"
	b.toastErr = false
}

func (b *BridgesView) buildBridgeConfig() tor.BridgeConfig {
	if b.activeTransport == tor.TransportDirect {
		return tor.BridgeConfig{Enabled: false}
	}

	var bridges []tor.Bridge
	if len(b.fetchedBridges) > 0 {
		bridges = b.fetchedBridges
	} else if len(b.customBridges) > 0 {
		bridges = b.customBridges
	} else {
		builtIns := tor.BuiltInBridgeEntries(b.activeTransport)
		if b.selected >= 0 && b.selected < len(builtIns) {
			bridges = []tor.Bridge{builtIns[b.selected]}
		} else if len(builtIns) > 0 {
			bridges = builtIns // all built-ins
		}
	}

	return tor.BridgeConfig{
		Enabled:   len(bridges) > 0,
		Bridges:   bridges,
		Transport: b.activeTransport,
	}
}

// ── Commands ──────────────────────────────────────────────────────────────

func (b *BridgesView) doFetchBridges() tea.Cmd {
	b.fetching = true
	transport := b.activeTransport
	if transport == tor.TransportDirect {
		transport = tor.TransportObfs4
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		bridges, err := tor.RequestBridgesFromBridgeDB(ctx, tor.BridgeDBOptions{
			Transport: transport,
			SocksPort: 9050,
			Count:     3,
		})
		return ui.BridgesFetchedMsg{Bridges: bridges, Err: err}
	}
}

func (b *BridgesView) doApplyBridge() tea.Cmd {
	ctrl := b.ctrl
	if ctrl == nil || !ctrl.IsConnected() {
		return func() tea.Msg {
			return ui.BridgeAppliedMsg{Err: fmt.Errorf("not connected to control port")}
		}
	}
	cfg := b.buildBridgeConfig()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := cfg.Apply(ctx, ctrl)
		return ui.BridgeAppliedMsg{Err: err}
	}
}
