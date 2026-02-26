package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/amirk1998/tor-manager/core/config"
	"github.com/amirk1998/tor-manager/core/proxy"
	"github.com/amirk1998/tor-manager/core/tor"
	"github.com/amirk1998/tor-manager/tor-tui/ui/views"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Tab int

const (
	TabDashboard Tab = iota
	TabBridges
	TabCountries
	TabLogs
	TabSettings
	tabCount
)

var tabLabels = [tabCount]string{"Dashboard", "Bridges", "Countries", "Logs", "Settings"}
var tabIcons = [tabCount]string{"◈", "⬡", "◎", "≡", "⚙"}

type controllerBundle struct {
	ctrl    *tor.Controller
	checker *proxy.Checker
	err     error
}

type App struct {
	width, height int
	activeTab     Tab
	cfg           *config.AppConfig
	ctrl          *tor.Controller
	checker       *proxy.Checker
	dashboard     *views.DashboardView
	bridges       *views.BridgesView
	countries     *views.CountriesView
	logs          *views.LogsView
	settings      *views.SettingsView
	booting       bool
	bootSpinner   spinner.Model
	bootMsg       string
	bootErr       string
	toast         string
	toastIsErr    bool
	toastExpiry   time.Time
	showHelp      bool
}

var (
	StyleHeaderBar = lipgloss.NewStyle().Background(ColorBgPanel).Foreground(ColorWhite).Padding(0, 2)
	StyleLogo      = lipgloss.NewStyle().Foreground(ColorPurpleSub).Bold(true)
)

func NewApp(cfg *config.AppConfig) *App {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = StylePrimary
	return &App{cfg: cfg, booting: true, bootMsg: "Connecting to Tor control port...", bootSpinner: sp}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.bootSpinner.Tick,
		a.doConnect(),
		tickCmd(time.Duration(a.cfg.AutoRefreshSecs)*time.Second),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.propagateSize()
	case TickMsg:
		cmds = append(cmds, tickCmd(time.Duration(a.cfg.AutoRefreshSecs)*time.Second))
		if !a.toastExpiry.IsZero() && time.Now().After(a.toastExpiry) {
			a.toast = ""
		}
		cmds = append(cmds, a.forwardToAll(msg)...)
	case spinner.TickMsg:
		if a.booting {
			var cmd tea.Cmd
			a.bootSpinner, cmd = a.bootSpinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, a.forwardToActive(msg)...)
	case controllerBundle:
		a.booting = false
		if msg.err != nil {
			a.bootErr = msg.err.Error()
		} else {
			a.ctrl, a.checker = msg.ctrl, msg.checker
			a.initViews()
			cmds = append(cmds, a.initViewCmds()...)
		}
	case BootstrapEventMsg:
		cmds = append(cmds, a.forwardToAll(msg)...)
	case LogLineMsg:
		if a.logs != nil {
			m, cmd := a.logs.Update(msg)
			a.logs = m.(*views.LogsView)
			cmds = append(cmds, cmd)
		}
	case ClipboardMsg:
		a.showToast("Copied: "+msg.Text, false)
	case ToastMsg:
		a.showToast(msg.Message, msg.IsError)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
		if msg.String() == "q" && !a.inputFocused() && !a.booting {
			return a, tea.Quit
		}
		if !a.booting && a.bootErr == "" {
			switch msg.String() {
			case "?":
				a.showHelp = !a.showHelp
				return a, nil
			case "esc":
				if a.showHelp {
					a.showHelp = false
					return a, nil
				}
			case "tab":
				if !a.showHelp {
					a.activeTab = (a.activeTab + 1) % tabCount
					return a, nil
				}
			case "shift+tab":
				if !a.showHelp {
					a.activeTab = (a.activeTab - 1 + tabCount) % tabCount
					return a, nil
				}
			case "1", "2", "3", "4", "5":
				if !a.showHelp {
					a.activeTab = Tab(msg.String()[0] - '1')
					return a, nil
				}
			default:
				if !a.showHelp {
					cmds = append(cmds, a.forwardToActive(msg)...)
				}
			}
		}
	default:
		cmds = append(cmds, a.forwardToActive(msg)...)
	}
	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if a.width == 0 {
		return "\n  Loading..."
	}
	if a.booting {
		return a.bootingView()
	}
	if a.bootErr != "" {
		return a.errorView()
	}
	if a.showHelp {
		return a.helpView()
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		a.renderHeader(), a.renderTabs(), a.renderContent(), a.renderFooter())
}

func (a *App) bootingView() string {
	logo := renderLogo()
	sp := a.bootSpinner.View() + " " + StyleDim.Render(a.bootMsg)
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, logo+"\n\n"+sp)
}

func (a *App) errorView() string {
	logo := renderLogo()
	e := StyleErrorBold.Render("✗  Cannot connect to Tor control port")
	d := StyleDim.Render(a.bootErr)
	h := "\n\n" + StyleDim.Render("Make sure Tor is running:  tor -f /etc/tor/torrc\n\n") + KeyHint("q", "quit")
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, logo+"\n\n"+e+"\n"+d+h)
}

func (a *App) renderHeader() string {
	logo := StyleLogo.Render("🧅 tor-manager")
	ver := StyleDim.Render("v0.1.0")
	gap := a.width - lipgloss.Width(logo) - lipgloss.Width(ver) - 4
	if gap < 1 {
		gap = 1
	}
	return StyleHeaderBar.Width(a.width).Render(logo + strings.Repeat(" ", gap) + ver)
}

func (a *App) renderTabs() string {
	var tabs []string
	for i := Tab(0); i < tabCount; i++ {
		label := " " + tabIcons[i] + " " + tabLabels[i] + " "
		if i == a.activeTab {
			tabs = append(tabs, StyleTabActive.Render(label))
		} else {
			tabs = append(tabs, StyleTabInactive.Render(label))
		}
	}
	return StyleTabBar.Width(a.width).Render(strings.Join(tabs, ""))
}

func (a *App) renderContent() string {
	v := a.activeView()
	if v == nil {
		return ""
	}
	content := v.View()
	lines := strings.Count(content, "\n") + 1
	avail := a.height - 5
	if lines < avail {
		content += strings.Repeat("\n", avail-lines)
	}
	return content
}

func (a *App) renderFooter() string {
	var line string
	if a.toast != "" {
		st := StyleSuccess
		if a.toastIsErr {
			st = StyleError
		}
		line = st.Render(a.toast)
	} else {
		line = strings.Join([]string{
			KeyHint("1-5", "tabs"), KeyHint("tab", "next"),
			KeyHint("?", "help"), KeyHint("q", "quit"),
		}, "  "+StyleDim.Render("·")+"  ")
	}
	return StyleFooter.Width(a.width).Render(line)
}

func (a *App) helpView() string {
	type item struct{ k, d string }
	secs := []struct {
		t string
		i []item
	}{
		{"Global", []item{{"1–5 / tab", "Switch tabs"}, {"q / ctrl+c", "Quit"}, {"?", "Close help"}}},
		{"Dashboard", []item{{"r", "Refresh IP"}, {"n", "New identity (10s cooldown)"}, {"c", "Copy proxy"}}},
		{"Bridges", []item{{"← →", "Change transport"}, {"↑ ↓ / enter", "Select"}, {"u", "Custom bridge"}, {"f", "Fetch from BridgeDB"}, {"a", "Apply"}, {"x", "Disable"}}},
		{"Countries", []item{{"↑ ↓", "Navigate"}, {"enter", "Toggle"}, {"s", "StrictNodes"}, {"a", "Apply"}, {"x", "Clear"}}},
		{"Logs", []item{{"↑ ↓", "Scroll"}, {"g / G", "Top/bottom"}, {"f", "Follow"}, {"c", "Clear"}}},
		{"Settings", []item{{"tab", "Next field"}, {"ctrl+s", "Save"}, {"esc", "Revert"}}},
	}
	var sb strings.Builder
	sb.WriteString(StyleBold.Render("  Keyboard Reference") + "\n\n")
	for _, sec := range secs {
		sb.WriteString(SectionTitle(sec.t) + "\n")
		for _, it := range sec.i {
			sb.WriteString(fmt.Sprintf("  %-26s%s\n", StyleAccent.Render(it.k), StyleDim.Render(it.d)))
		}
		sb.WriteByte('\n')
	}
	sb.WriteString(KeyHint("esc / ?", "close"))
	w := a.width - 8
	if w > 60 {
		w = 60
	}
	panel := StylePanelActive.Width(w).Render(sb.String())
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, panel)
}

func (a *App) initViews() {
	a.dashboard = views.NewDashboardView(a.checker, a.ctrl)
	a.bridges = views.NewBridgesView(a.ctrl)
	a.countries = views.NewCountriesView(a.ctrl)
	a.logs = views.NewLogsView()
	a.settings = views.NewSettingsView(a.cfg)
	a.propagateSize()
}

func (a *App) initViewCmds() []tea.Cmd {
	var cmds []tea.Cmd
	for _, v := range []tea.Model{a.dashboard, a.bridges, a.logs} {
		if v != nil {
			cmds = append(cmds, v.Init())
		}
	}
	return cmds
}

func (a *App) propagateSize() {
	w, h := a.width, a.height-5
	if h < 5 {
		h = 5
	}
	for _, v := range []interface{ SetSize(int, int) }{
		a.dashboard, a.bridges, a.countries, a.logs, a.settings,
	} {
		if v != nil {
			v.SetSize(w, h)
		}
	}
}

type sizedModel interface {
	tea.Model
	View() string
	SetSize(int, int)
}

func (a *App) activeView() sizedModel {
	switch a.activeTab {
	case TabDashboard:
		return a.dashboard
	case TabBridges:
		return a.bridges
	case TabCountries:
		return a.countries
	case TabLogs:
		return a.logs
	case TabSettings:
		return a.settings
	}
	return nil
}

func (a *App) setViewByModel(m tea.Model) {
	switch v := m.(type) {
	case *views.DashboardView:
		a.dashboard = v
	case *views.BridgesView:
		a.bridges = v
	case *views.CountriesView:
		a.countries = v
	case *views.LogsView:
		a.logs = v
	case *views.SettingsView:
		a.settings = v
	}
}

func (a *App) forwardToActive(msg tea.Msg) []tea.Cmd {
	v := a.activeView()
	if v == nil {
		return nil
	}
	m, cmd := v.Update(msg)
	a.setViewByModel(m)
	if cmd != nil {
		return []tea.Cmd{cmd}
	}
	return nil
}

func (a *App) forwardToAll(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	for _, v := range []sizedModel{a.dashboard, a.bridges, a.countries, a.logs, a.settings} {
		if v == nil {
			continue
		}
		m, cmd := v.Update(msg)
		a.setViewByModel(m)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

func (a *App) inputFocused() bool { return a.activeTab == TabSettings }

func (a *App) doConnect() tea.Cmd {
	cfg := a.cfg
	return func() tea.Msg {
		cc := tor.DefaultControllerConfig()
		cc.Addr = fmt.Sprintf("127.0.0.1:%d", cfg.ControlPort)
		if cfg.ControlAuth == "password" {
			cc.AuthMethod = tor.AuthPassword
			cc.Password = cfg.Password
		} else {
			cc.AuthMethod = tor.AuthCookie
		}
		ctrl, err := tor.NewController(cc)
		if err != nil {
			return controllerBundle{err: err}
		}
		checker, err := proxy.NewChecker(cfg.SocksPort)
		if err != nil {
			ctrl.Close()
			return controllerBundle{err: err}
		}
		return controllerBundle{ctrl: ctrl, checker: checker}
	}
}

func (a *App) showToast(msg string, isErr bool) {
	a.toast, a.toastIsErr = msg, isErr
	a.toastExpiry = time.Now().Add(4 * time.Second)
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return TickMsg(t) })
}

func renderLogo() string {
	lines := []string{
		"  ████████╗ ██████╗ ██████╗ ",
		"     ██╔══╝██╔═══██╗██╔══██╗",
		"     ██║   ██║   ██║██████╔╝",
		"     ██║   ██║   ██║██╔══██╗",
		"     ██║   ╚██████╔╝██║  ██║",
		"     ╚═╝    ╚═════╝ ╚═╝  ╚═╝",
		"", "  🧅  M A N A G E R",
	}
	var sb strings.Builder
	for i, line := range lines {
		if i < 6 {
			sb.WriteString(StylePrimary.Render(line))
		} else {
			sb.WriteString(StyleAccent.Render(line))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// contextBg returns a background context (used by commands).
func contextBg() context.Context { return context.Background() }
