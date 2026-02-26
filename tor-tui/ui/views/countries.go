package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/amirk1998/tor-manager/core/tor"
	"github.com/amirk1998/tor-manager/tor-tui/ui/common"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CountriesView struct {
	width  int
	height int

	cursor    int
	scrollOff int
	selected  []string
	strict    bool

	toast    string
	toastErr bool

	ctrl *tor.Controller
}

type countryEntry struct {
	Code string
	Name string
	Flag string
}

var countryList = []countryEntry{
	{"", "Any (Random)", "🌐"},
	{"US", "United States", "🇺🇸"},
	{"DE", "Germany", "🇩🇪"},
	{"NL", "Netherlands", "🇳🇱"},
	{"FR", "France", "🇫🇷"},
	{"SE", "Sweden", "🇸🇪"},
	{"CH", "Switzerland", "🇨🇭"},
	{"NO", "Norway", "🇳🇴"},
	{"IS", "Iceland", "🇮🇸"},
	{"GB", "United Kingdom", "🇬🇧"},
	{"CA", "Canada", "🇨🇦"},
	{"JP", "Japan", "🇯🇵"},
	{"SG", "Singapore", "🇸🇬"},
	{"AU", "Australia", "🇦🇺"},
	{"AT", "Austria", "🇦🇹"},
	{"BE", "Belgium", "🇧🇪"},
	{"FI", "Finland", "🇫🇮"},
	{"CZ", "Czech Republic", "🇨🇿"},
	{"RO", "Romania", "🇷🇴"},
	{"LU", "Luxembourg", "🇱🇺"},
}

func NewCountriesView(ctrl *tor.Controller) *CountriesView {
	return &CountriesView{ctrl: ctrl}
}

func (c *CountriesView) SetSize(w, h int) {
	c.width = w
	c.height = h
}

func (c *CountriesView) Init() tea.Cmd { return nil }

func (c *CountriesView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case common.ExitCountryAppliedMsg:
		if msg.Err != nil {
			c.toast = "Apply failed: " + msg.Err.Error()
			c.toastErr = true
		} else {
			c.toast = "Exit country applied!"
			c.toastErr = false
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if c.cursor > 0 {
				c.cursor--
				c.clampScroll()
			}
		case "down", "j":
			if c.cursor < len(countryList)-1 {
				c.cursor++
				c.clampScroll()
			}
		case "enter", " ":
			c.toggleCountry()
		case "s":
			c.strict = !c.strict
		case "a":
			return c, c.doApply()
		case "x":
			c.selected = nil
			c.toast = "Country filter cleared"
			c.toastErr = false
		}
	}
	return c, nil
}

func (c *CountriesView) View() string {
	if c.width == 0 || c.height == 0 {
		return ""
	}

	// ─── layout constants ───────────────────────────────────────────────

	footerH := 1
	if c.toast != "" {
		footerH = 2
	}

	actionH := footerH + 1

	rightFixedLines := 12

	panelOverhead := 4

	leftFixedLines := 4
	listRows := c.height - actionH - panelOverhead - leftFixedLines
	if listRows < 3 {
		listRows = 3
	}

	rightContentH := rightFixedLines
	if rightContentH < listRows+leftFixedLines {
		rightContentH = listRows + leftFixedLines
	}
	_ = rightContentH

	totalW := c.width - 2
	if totalW > 80 {
		totalW = 80
	}

	leftW := totalW * 55 / 100
	rightW := totalW - leftW - 2
	if leftW < 24 {
		leftW = 24
	}
	if rightW < 22 {
		rightW = 22
	}

	left := c.renderList(leftW, listRows)
	right := c.renderInfo(rightW)
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	footer := c.renderActions()

	return lipgloss.JoinVertical(lipgloss.Left, row, footer)
}

func (c *CountriesView) renderList(w, visRows int) string {
	// clamp scrollOff
	c.clampScrollWithRows(visRows)

	var sb strings.Builder
	sb.WriteString(common.StyleBold.Render("  Exit Country") + "\n")
	sb.WriteString(common.StyleDim.Render("  ↑↓ navigate  space/enter select") + "\n")
	sb.WriteString("\n")

	end := c.scrollOff + visRows
	if end > len(countryList) {
		end = len(countryList)
	}

	for i := c.scrollOff; i < end; i++ {
		entry := countryList[i]

		cur := "  "
		if i == c.cursor {
			cur = common.StyleCursor.Render(" ▶")
		}

		chk := "  "
		if c.isSelected(entry.Code) {
			chk = common.StyleSuccessBold.Render(" ✓")
		}

		nameStyle := common.StyleRowNormal
		if i == c.cursor {
			nameStyle = common.StyleRowSelected
		}

		maxName := w - 14
		if maxName < 8 {
			maxName = 8
		}
		name := truncateName(entry.Name, maxName)
		line := fmt.Sprintf("%s%s %s %s", cur, chk, entry.Flag, nameStyle.Render(name))
		sb.WriteString(line + "\n")
	}

	// scroll indicator — tek satır
	if len(countryList) > visRows {
		total := len(countryList) - 1
		pct := 0
		if total > 0 {
			pct = c.cursor * 100 / total
		}
		sb.WriteString(common.StyleDim.Render(
			fmt.Sprintf("  ─── %d/%d  %d%% ───", c.cursor+1, len(countryList), pct),
		))
	}

	return common.StylePanel.Width(w).Render(sb.String())
}

func (c *CountriesView) renderInfo(w int) string {
	var sb strings.Builder

	sb.WriteString(common.StyleBold.Render("  Active Filter") + "\n\n")

	if len(c.selected) == 0 {
		sb.WriteString("  " + common.StyleDim.Render("🌐 Any (random)") + "\n")
		sb.WriteString("\n")
		sb.WriteString(common.StyleDim.Render("  Tor picks exits\n  from any country.") + "\n")
	} else {
		for _, code := range c.selected {
			e := findCountry(code)
			sb.WriteString(fmt.Sprintf("  %s %s %s\n",
				common.StyleSuccessBold.Render("✓"),
				e.Flag,
				common.StyleAccent.Render(truncateName(e.Name, w-10)),
			))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(common.Divider(w-4) + "\n\n")

	// StrictNodes
	strictVal := common.StyleError.Render("OFF")
	if c.strict {
		strictVal = common.StyleSuccess.Render("ON ")
	}
	sb.WriteString(common.StyleDim.Render("  StrictNodes: ") + strictVal + "\n")
	sb.WriteString(common.StyleDim.Render("  [s] to toggle") + "\n\n")
	sb.WriteString(common.StyleDim.Render("  ON = force exit\n  country. Reduces\n  anonymity."))

	return common.StylePanel.Width(w).Render(sb.String())
}

func (c *CountriesView) renderActions() string {
	hints := strings.Join([]string{
		common.KeyHint("↑↓", "navigate"),
		common.KeyHint("enter", "toggle"),
		common.KeyHint("s", "strict"),
		common.KeyHint("a", "apply"),
		common.KeyHint("x", "clear"),
	}, "  ")

	if c.toast != "" {
		toastStyle := common.StyleSuccess
		prefix := "✓ "
		if c.toastErr {
			toastStyle = common.StyleError
			prefix = "✗ "
		}
		return hints + "\n" + toastStyle.Render(prefix+c.toast)
	}
	return hints
}

// ─── helpers ────────────────────────────────────────────────────────────

func (c *CountriesView) clampScroll() {
	vis := c.height - 8
	if vis < 3 {
		vis = 3
	}
	c.clampScrollWithRows(vis)
}

func (c *CountriesView) clampScrollWithRows(vis int) {
	if c.cursor < c.scrollOff {
		c.scrollOff = c.cursor
	}
	if c.cursor >= c.scrollOff+vis {
		c.scrollOff = c.cursor - vis + 1
	}
	if c.scrollOff < 0 {
		c.scrollOff = 0
	}
}

func (c *CountriesView) toggleCountry() {
	entry := countryList[c.cursor]
	if entry.Code == "" {
		c.selected = nil
		c.toast = "Country filter cleared"
		c.toastErr = false
		return
	}
	if c.isSelected(entry.Code) {
		var ns []string
		for _, code := range c.selected {
			if code != entry.Code {
				ns = append(ns, code)
			}
		}
		c.selected = ns
	} else {
		c.selected = append(c.selected, entry.Code)
	}
}

func (c *CountriesView) isSelected(code string) bool {
	if code == "" {
		return len(c.selected) == 0
	}
	for _, s := range c.selected {
		if s == code {
			return true
		}
	}
	return false
}

func findCountry(code string) countryEntry {
	for _, e := range countryList {
		if e.Code == code {
			return e
		}
	}
	return countryEntry{Code: code, Name: code, Flag: "🏳️"}
}

func truncateName(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

// ─── command ────────────────────────────────────────────────────────────

func (c *CountriesView) doApply() tea.Cmd {
	ctrl := c.ctrl
	if ctrl == nil || !ctrl.IsConnected() {
		return func() tea.Msg {
			return common.ExitCountryAppliedMsg{Err: fmt.Errorf("not connected to control port")}
		}
	}
	selected := c.selected
	strict := c.strict
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var err error
		if len(selected) == 0 {
			err = ctrl.ResetConf(ctx, "ExitNodes")
			if err == nil {
				err = ctrl.ResetConf(ctx, "StrictNodes")
			}
		} else {
			codes := make([]string, len(selected))
			for i, code := range selected {
				codes[i] = "{" + code + "}"
			}
			err = ctrl.SetConf(ctx, "ExitNodes", strings.Join(codes, ","))
			if err == nil {
				strictVal := "0"
				if strict {
					strictVal = "1"
				}
				err = ctrl.SetConf(ctx, "StrictNodes", strictVal)
			}
		}
		return common.ExitCountryAppliedMsg{Err: err}
	}
}
