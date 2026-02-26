package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/amirk1998/tor-manager/core/config"
	"github.com/amirk1998/tor-manager/core/proxy"
	"github.com/amirk1998/tor-manager/core/tor"
	"github.com/amirk1998/tor-manager/tor-tui/ui/common"
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
	case common.TickMsg:
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
			cmds = append(cmds, a.doBootstrapCheck())
		}
	case common.BootstrapEventMsg:
		cmds = append(cmds, a.forwardToAll(msg)...)
	case common.LogLineMsg:
		if a.logs != nil {
			m, cmd := a.logs.Update(msg)
			a.logs = m.(*views.LogsView)
			cmds = append(cmds, cmd)
		}
	case common.ClipboardMsg:
		a.showToast("Copied: "+msg.Text, false)
	case common.ToastMsg:
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

	header := a.renderHeader()
	tabs := a.renderTabs()
	footer := a.renderFooter()

	usedH := lipgloss.Height(header) + lipgloss.Height(tabs) + lipgloss.Height(footer)
	contentH := a.height - usedH
	if contentH < 1 {
		contentH = 1
	}

	content := a.renderContent(contentH)

	return lipgloss.JoinVertical(lipgloss.Left, header, tabs, content, footer)
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

func (a *App) renderContent(avail int) string {
	v := a.activeView()
	if v == nil {
		return strings.Repeat("\n", avail-1)
	}

	content := v.View()
	actual := lipgloss.Height(content)

	if actual > avail {
		// ✅ Clip — محتوای اضافه رو قطع میکنیم تا header/footer جابجا نشه
		lines := strings.Split(content, "\n")
		if len(lines) > avail {
			lines = lines[:avail]
		}
		return strings.Join(lines, "\n")
	}
	if actual < avail {
		// padding پایین برای پر کردن فضا
		content += strings.Repeat("\n", avail-actual)
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
	contentH := a.height - 4
	if contentH < 5 {
		contentH = 5
	}
	if a.dashboard != nil {
		a.dashboard.SetSize(a.width, contentH)
	}
	if a.bridges != nil {
		a.bridges.SetSize(a.width, contentH)
	}
	if a.countries != nil {
		a.countries.SetSize(a.width, contentH)
	}
	if a.logs != nil {
		a.logs.SetSize(a.width, contentH)
	}
	if a.settings != nil {
		a.settings.SetSize(a.width, contentH)
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
		if a.dashboard == nil {
			return nil
		}
		return a.dashboard
	case TabBridges:
		if a.bridges == nil {
			return nil
		}
		return a.bridges
	case TabCountries:
		if a.countries == nil {
			return nil
		}
		return a.countries
	case TabLogs:
		if a.logs == nil {
			return nil
		}
		return a.logs
	case TabSettings:
		if a.settings == nil {
			return nil
		}
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
	type namedView struct {
		m sizedModel
	}
	views := []sizedModel{}
	if a.dashboard != nil {
		views = append(views, a.dashboard)
	}
	if a.bridges != nil {
		views = append(views, a.bridges)
	}
	if a.countries != nil {
		views = append(views, a.countries)
	}
	if a.logs != nil {
		views = append(views, a.logs)
	}
	if a.settings != nil {
		views = append(views, a.settings)
	}
	for _, v := range views {
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

// doBootstrapCheck fetches the current bootstrap status from a running Tor daemon
// and emits a BootstrapEventMsg so views update their connected state immediately.
func (a *App) doBootstrapCheck() tea.Cmd {
	ctrl := a.ctrl
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done, statusStr, err := ctrl.IsBootstrapped(ctx)
		if err != nil {
			return nil // بی‌صدا fail میشه — IP check بعداً وضعیت رو آپدیت میکنه
		}

		progress := parseStatusProgress(statusStr)
		if done {
			progress = 100
		}

		// همچنین Tor version رو بگیر
		version, _ := ctrl.GetVersion(ctx)
		_ = version // در مرحله بعد از طریق toast نمایش میدیم

		return common.BootstrapEventMsg{
			Event: tor.BootstrapEvent{Progress: progress},
		}
	}
}

// parseStatusProgress extracts PROGRESS=N from a Tor bootstrap status string.
func parseStatusProgress(status string) int {
	const tag = "PROGRESS="
	idx := strings.Index(status, tag)
	if idx < 0 {
		return 0
	}
	var n int
	fmt.Sscanf(status[idx+len(tag):], "%d", &n)
	return n
}

func (a *App) showToast(msg string, isErr bool) {
	a.toast, a.toastIsErr = msg, isErr
	a.toastExpiry = time.Now().Add(4 * time.Second)
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return common.TickMsg(t) })
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
