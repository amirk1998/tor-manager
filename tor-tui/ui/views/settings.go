package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/you/tor-manager/core/config"
	"github.com/you/tor-tui/ui"
)

type settingField int

const (
	fieldSocksPort settingField = iota
	fieldControlPort
	fieldPassword
	fieldAutoRefresh
	fieldCount // sentinel
)

// SettingsView lets the user configure ports and auth.
type SettingsView struct {
	width  int
	height int

	cfg    *config.AppConfig
	inputs [fieldCount]textinput.Model
	focus  settingField

	dirty    bool
	toast    string
	toastErr bool
}

func NewSettingsView(cfg *config.AppConfig) *SettingsView {
	s := &SettingsView{cfg: cfg}
	s.initInputs()
	return s
}

func (s *SettingsView) initInputs() {
	newInput := func(value, placeholder string, charLimit int) textinput.Model {
		ti := textinput.New()
		ti.SetValue(value)
		ti.Placeholder = placeholder
		ti.CharLimit = charLimit
		ti.Width = 20
		return ti
	}

	s.inputs[fieldSocksPort] = newInput(
		fmt.Sprintf("%d", s.cfg.SocksPort),
		"9050", 5,
	)
	s.inputs[fieldControlPort] = newInput(
		fmt.Sprintf("%d", s.cfg.ControlPort),
		"9051", 5,
	)
	s.inputs[fieldPassword] = newInput(
		s.cfg.Password,
		"(leave blank for cookie auth)", 128,
	)
	s.inputs[fieldPassword].EchoMode = textinput.EchoPassword
	s.inputs[fieldPassword].EchoCharacter = '•'

	s.inputs[fieldAutoRefresh] = newInput(
		fmt.Sprintf("%d", s.cfg.AutoRefreshSecs),
		"30", 3,
	)

	s.inputs[s.focus].Focus()
}

func (s *SettingsView) SetSize(w, h int) { s.width = w; s.height = h }
func (s *SettingsView) Init() tea.Cmd    { return textinput.Blink }

func (s *SettingsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			s.nextField()
		case "shift+tab", "up":
			s.prevField()
		case "enter":
			s.nextField()
		case "ctrl+s":
			s.save()
		case "esc":
			s.revert()
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	s.inputs[s.focus], cmd = s.inputs[s.focus].Update(msg)
	cmds = append(cmds, cmd)
	s.dirty = true

	return s, tea.Batch(cmds...)
}

func (s *SettingsView) View() string {
	if s.width == 0 {
		return ""
	}
	w := s.width - 4
	if w > 72 {
		w = 72
	}

	sections := []string{
		s.renderPortsSection(w),
		s.renderAuthSection(w),
		s.renderPrefsSection(w),
		s.renderActions(w),
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (s *SettingsView) renderPortsSection(w int) string {
	rows := []string{
		ui.SectionTitle("Ports"),
		"",
		s.renderField(fieldSocksPort, "SOCKS5 Port",
			"Applications connect here to route traffic through Tor."),
		"",
		s.renderField(fieldControlPort, "Control Port",
			"Used internally for bridge/identity management."),
	}
	return ui.StylePanel.Width(w).Render(strings.Join(rows, "\n"))
}

func (s *SettingsView) renderAuthSection(w int) string {
	authMode := "Cookie  " + ui.StyleSuccess.Render("(recommended)")
	if s.cfg.ControlAuth == "password" {
		authMode = "Password"
	}

	rows := []string{
		ui.SectionTitle("Control Port Authentication"),
		"",
		fmt.Sprintf("  %-22s %s", "Current Mode", ui.StyleAccent.Render(authMode)),
		"",
		s.renderField(fieldPassword, "Password",
			"Leave blank to use cookie auth (more secure)."),
	}
	return ui.StylePanel.Width(w).Render(strings.Join(rows, "\n"))
}

func (s *SettingsView) renderPrefsSection(w int) string {
	rows := []string{
		ui.SectionTitle("Preferences"),
		"",
		s.renderField(fieldAutoRefresh, "Auto-refresh (sec)",
			"How often the IP check runs automatically."),
	}
	return ui.StylePanel.Width(w).Render(strings.Join(rows, "\n"))
}

func (s *SettingsView) renderField(f settingField, label, desc string) string {
	isFocused := s.focus == f
	inputStyle := ui.StyleInputBlurred
	if isFocused {
		inputStyle = ui.StyleInputFocused
	}

	labelStr := fmt.Sprintf("  %-22s", ui.StyleInputLabel.Render(label))
	inputStr := inputStyle.Render(s.inputs[f].View())
	descStr := "  " + ui.StyleDim.Render(desc)

	return labelStr + inputStr + "\n" + descStr
}

func (s *SettingsView) renderActions(w int) string {
	hints := strings.Join([]string{
		ui.KeyHint("tab/↓", "next field"),
		ui.KeyHint("ctrl+s", "save"),
		ui.KeyHint("esc", "revert"),
	}, "  ")

	dirtyStr := ""
	if s.dirty {
		dirtyStr = "  " + ui.StyleWarning.Render("● unsaved changes")
	}

	footer := hints + dirtyStr

	if s.toast != "" {
		style := ui.StyleSuccess
		prefix := "✓ "
		if s.toastErr {
			style = ui.StyleError
			prefix = "✗ "
		}
		return footer + "\n" + style.Render(prefix+s.toast)
	}
	return footer
}

// ── Logic ─────────────────────────────────────────────────────────────────

func (s *SettingsView) nextField() {
	s.inputs[s.focus].Blur()
	s.focus = (s.focus + 1) % fieldCount
	s.inputs[s.focus].Focus()
}

func (s *SettingsView) prevField() {
	s.inputs[s.focus].Blur()
	s.focus = (s.focus - 1 + fieldCount) % fieldCount
	s.inputs[s.focus].Focus()
}

func (s *SettingsView) save() {
	// Parse and validate
	var socksPort, controlPort, autoRefresh int
	if _, err := fmt.Sscanf(s.inputs[fieldSocksPort].Value(), "%d", &socksPort); err != nil || socksPort < 1 || socksPort > 65535 {
		s.toast = "Invalid SOCKS5 port (1–65535)"
		s.toastErr = true
		return
	}
	if _, err := fmt.Sscanf(s.inputs[fieldControlPort].Value(), "%d", &controlPort); err != nil || controlPort < 1 || controlPort > 65535 {
		s.toast = "Invalid control port (1–65535)"
		s.toastErr = true
		return
	}
	if socksPort == controlPort {
		s.toast = "SOCKS5 and control port must differ"
		s.toastErr = true
		return
	}
	if _, err := fmt.Sscanf(s.inputs[fieldAutoRefresh].Value(), "%d", &autoRefresh); err != nil || autoRefresh < 5 {
		s.toast = "Auto-refresh must be ≥ 5 seconds"
		s.toastErr = true
		return
	}

	s.cfg.SocksPort = socksPort
	s.cfg.ControlPort = controlPort
	s.cfg.AutoRefreshSecs = autoRefresh

	pw := s.inputs[fieldPassword].Value()
	if pw != "" {
		s.cfg.Password = pw
		s.cfg.ControlAuth = "password"
	} else {
		s.cfg.Password = ""
		s.cfg.ControlAuth = "cookie"
	}

	if err := s.cfg.Save(); err != nil {
		s.toast = "Save failed: " + err.Error()
		s.toastErr = true
		return
	}

	s.dirty = false
	s.toast = "Settings saved"
	s.toastErr = false
}

func (s *SettingsView) revert() {
	s.initInputs()
	s.dirty = false
	s.toast = "Changes reverted"
	s.toastErr = false
}
