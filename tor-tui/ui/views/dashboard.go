package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/amirk1998/tor-manager/core/proxy"
	"github.com/amirk1998/tor-manager/core/tor"
	"github.com/amirk1998/tor-manager/tor-tui/ui/common"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DashboardView is the main connection status screen.
type DashboardView struct {
	width  int
	height int

	// State
	ipResult    *proxy.CheckResult
	bootstrap   *tor.BootstrapEvent
	torVersion  string
	connected   bool
	loading     bool
	lastRefresh time.Time
	cooldown    int // seconds remaining before NEWNYM is allowed

	// Spinner for loading state
	spinner spinner.Model

	// Refs
	checker *proxy.Checker
	ctrl    *tor.Controller

	// Toast
	toast    string
	toastErr bool
}

func NewDashboardView(checker *proxy.Checker, ctrl *tor.Controller) *DashboardView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = common.StylePrimary

	return &DashboardView{
		spinner: sp,
		checker: checker,
		ctrl:    ctrl,
	}
}

func (d *DashboardView) SetSize(w, h int) {
	d.width = w
	d.height = h
}

func (d *DashboardView) SetBootstrap(evt tor.BootstrapEvent) {
	d.bootstrap = &evt
	if evt.Progress == 100 {
		d.connected = true
	}
}

func (d *DashboardView) SetTorVersion(v string) { d.torVersion = v }

// Init starts the spinner and triggers the initial IP check.
func (d *DashboardView) Init() tea.Cmd {
	return tea.Batch(
		d.spinner.Tick,
		d.doCheckIP(),
	)
}

func (d *DashboardView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case common.TickMsg:
		cmds = append(cmds, d.doCheckIP())
		if d.cooldown > 0 {
			d.cooldown--
		}

	case common.IPCheckResultMsg:
		d.loading = false
		d.lastRefresh = time.Now()
		res := msg.Result
		d.ipResult = &res
		if res.Error == nil && res.IPInfo != nil {
			d.connected = true
		}

	case common.NewNymResultMsg:
		if msg.Err != nil {
			d.toast = "New identity failed: " + msg.Err.Error()
			d.toastErr = true
		} else {
			d.toast = "New identity requested — fetching new IP..."
			d.toastErr = false
			d.cooldown = 10
			cmds = append(cmds, d.doCheckIP())
		}

	case common.BootstrapEventMsg:
		d.SetBootstrap(msg.Event)

	case tea.KeyMsg:
		switch {
		case msg.String() == "r":
			d.loading = true
			d.toast = ""
			cmds = append(cmds, d.doCheckIP())
		case msg.String() == "n":
			d.toast = ""
			cmds = append(cmds, d.doNewNym())
		case msg.String() == "c":
			cmds = append(cmds, func() tea.Msg {
				return common.ClipboardMsg{Text: fmt.Sprintf("socks5://127.0.0.1:%d", 9050)}
			})
		}
	}

	return d, tea.Batch(cmds...)
}

func (d *DashboardView) View() string {
	if d.width == 0 {
		return ""
	}

	panelW := d.width - 4
	if panelW > 72 {
		panelW = 72
	}

	sections := []string{
		d.renderConnectionCard(panelW),
		d.renderIPCard(panelW),
		d.renderProxyCard(panelW),
		d.renderActions(panelW),
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (d *DashboardView) renderConnectionCard(w int) string {
	var rows []string
	rows = append(rows, common.SectionTitle("Connection Status"))
	rows = append(rows, "")

	// Status line
	statusStr := common.StatusDot(d.connected, d.isBootstrapping())
	rows = append(rows, fmt.Sprintf("  %-20s %s", "Status", statusStr))

	// Bootstrap progress
	if d.isBootstrapping() && d.bootstrap != nil {
		rows = append(rows, fmt.Sprintf("  %-20s %s",
			"Bootstrap",
			common.RenderProgressBar(d.bootstrap.Progress, 20),
		))
		if d.bootstrap.Summary != "" {
			rows = append(rows, fmt.Sprintf("  %-20s %s",
				"",
				common.StyleDim.Render(d.bootstrap.Summary),
			))
		}
	}

	if d.torVersion != "" {
		rows = append(rows, fmt.Sprintf("  %-20s %s",
			"Tor Version",
			common.StyleAccent.Render(d.torVersion),
		))
	}

	if !d.lastRefresh.IsZero() {
		rows = append(rows, fmt.Sprintf("  %-20s %s",
			"Last Checked",
			common.StyleDim.Render(d.lastRefresh.Format("15:04:05")),
		))
	}

	content := strings.Join(rows, "\n")
	return common.StylePanel.Width(w).Render(content)
}

func (d *DashboardView) renderIPCard(w int) string {
	var rows []string
	rows = append(rows, common.SectionTitle("Exit Node Info"))
	rows = append(rows, "")

	if d.loading {
		rows = append(rows, "  "+d.spinner.View()+" "+common.StyleDim.Render("Fetching IP through Tor..."))
	} else if d.ipResult == nil {
		rows = append(rows, "  "+common.StyleDim.Render("Not checked yet  —  press [r] to refresh"))
	} else if d.ipResult.Error != nil {
		rows = append(rows, "  "+common.StyleErrorBold.Render("✗ Cannot reach Tor exit node"))
		rows = append(rows, "  "+common.StyleDim.Render(d.ipResult.Error.Error()))
	} else if d.ipResult.IPInfo != nil {
		info := d.ipResult.IPInfo

		torBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color(common.ColorBg)).Background(lipgloss.Color(common.ColorGreen)).Bold(true).
			Padding(0, 1).Render("✓ TOR EXIT")
		if !info.IsTor {
			torBadge = lipgloss.NewStyle().
				Foreground(lipgloss.Color(common.ColorBg)).Background(lipgloss.Color(common.ColorRed)).Bold(true).
				Padding(0, 1).Render("✗ NOT TOR")
		}

		rows = append(rows,
			fmt.Sprintf("  %-20s %s %s",
				"Exit IP",
				common.StyleBold.Render(info.IP),
				torBadge,
			),
		)

		if info.City != "" || info.Country != "" {
			loc := info.City
			if loc != "" && info.Country != "" {
				loc += ", "
			}
			loc += info.Country
			rows = append(rows, fmt.Sprintf("  %-20s %s", "Location", common.StyleAccent.Render(loc)))
		}
		if info.Org != "" {
			rows = append(rows, fmt.Sprintf("  %-20s %s",
				"Network",
				common.StyleDim.Render(truncate(info.Org, 40)),
			))
		}
		rows = append(rows, fmt.Sprintf("  %-20s %s",
			"Latency",
			common.StyleDim.Render(d.ipResult.Latency.Round(time.Millisecond).String()),
		))
	}

	content := strings.Join(rows, "\n")
	return common.StylePanel.Width(w).Render(content)
}

func (d *DashboardView) renderProxyCard(w int) string {
	addr := lipgloss.NewStyle().
		Foreground(lipgloss.Color(common.ColorPurpleSub)).
		Bold(true).
		Render("socks5://127.0.0.1:9050")

	dns := common.StyleDim.Render("dns://127.0.0.1:9053 (if enabled)")

	content := common.SectionTitle("Proxy Addresses") + "\n\n" +
		fmt.Sprintf("  %-20s %s\n", "SOCKS5", addr) +
		fmt.Sprintf("  %-20s %s", "DNS-over-Tor", dns)

	return common.StylePanel.Width(w).Render(content)
}

func (d *DashboardView) renderActions(w int) string {
	hints := []string{
		common.KeyHint("r", "Refresh IP"),
	}
	if d.cooldown > 0 {
		hints = append(hints,
			common.StyleDim.Render(fmt.Sprintf("[n] New Identity (%ds)", d.cooldown)),
		)
	} else {
		hints = append(hints, common.KeyHint("n", "New Identity"))
	}
	hints = append(hints, common.KeyHint("c", "Copy Proxy"))

	line := strings.Join(hints, "  "+common.StyleDim.Render("·")+"  ")

	if d.toast != "" {
		toastStyle := common.StyleSuccess
		prefix := "✓ "
		if d.toastErr {
			toastStyle = common.StyleError
			prefix = "✗ "
		}
		return line + "\n" + toastStyle.Render(prefix+d.toast)
	}
	return line
}

// ── Commands ──────────────────────────────────────────────────────────────

func (d *DashboardView) doCheckIP() tea.Cmd {
	checker := d.checker
	if checker == nil {
		return nil
	}
	d.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		return common.IPCheckResultMsg{Result: checker.CheckIP(ctx)}
	}
}

func (d *DashboardView) doNewNym() tea.Cmd {
	ctrl := d.ctrl
	if ctrl == nil || !ctrl.IsConnected() {
		return func() tea.Msg {
			return common.NewNymResultMsg{Err: fmt.Errorf("not connected to control port")}
		}
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := ctrl.NewNym(ctx)
		return common.NewNymResultMsg{Err: err}
	}
}

func (d *DashboardView) isBootstrapping() bool {
	return d.bootstrap != nil && d.bootstrap.Progress < 100
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
