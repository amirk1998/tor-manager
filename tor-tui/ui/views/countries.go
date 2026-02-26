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

// CountriesView lets the user pick an exit country (or leave it random).
type CountriesView struct {
	width  int
	height int

	cursor   int
	selected []string // selected country codes; empty = any

	strict bool // StrictNodes

	toast    string
	toastErr bool

	ctrl *tor.Controller
}

type countryEntry struct {
	Code string
	Name string
	Flag string
}

// countryList is an extended list of countries with flag emojis.
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
			}
		case "down", "j":
			if c.cursor < len(countryList)-1 {
				c.cursor++
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
	if c.width == 0 {
		return ""
	}
	w := c.width - 4
	if w > 78 {
		w = 78
	}

	left := c.renderCountryList(w/2 - 1)
	right := c.renderSelectionPanel(w/2 - 1)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	actions := c.renderActions(w)

	return lipgloss.JoinVertical(lipgloss.Left, columns, actions)
}

func (c *CountriesView) renderCountryList(w int) string {
	var rows []string
	rows = append(rows, common.SectionTitle("Exit Country"))
	rows = append(rows, common.StyleDim.Render("  Space/enter to toggle"))
	rows = append(rows, "")

	for i, entry := range countryList {
		cursor := "  "
		nameStyle := common.StyleRowNormal
		checkmark := "  "

		if i == c.cursor {
			cursor = common.StyleCursor.Render(" ▶")
			nameStyle = common.StyleRowSelected
		}
		if c.isSelected(entry.Code) {
			checkmark = common.StyleSuccessBold.Render(" ✓")
		}

		flag := entry.Flag
		name := truncate(entry.Name, w-12)

		row := fmt.Sprintf("%s%s %s %s", cursor, checkmark, flag, nameStyle.Render(name))
		rows = append(rows, row)
	}

	content := strings.Join(rows, "\n")
	return common.StylePanel.Width(w).Render(content)
}

func (c *CountriesView) renderSelectionPanel(w int) string {
	var rows []string
	rows = append(rows, common.SectionTitle("Active Filter"))
	rows = append(rows, "")

	if len(c.selected) == 0 {
		rows = append(rows, "  "+common.StyleDim.Render("🌐 Any country (random)"))
		rows = append(rows, "")
		rows = append(rows,
			common.StyleDim.Render("  Tor will choose exit nodes\n  from any available country."),
		)
	} else {
		for _, code := range c.selected {
			entry := findCountry(code)
			rows = append(rows, "  "+
				common.StyleSuccessBold.Render("✓ ")+
				entry.Flag+" "+
				common.StyleAccent.Render(entry.Name),
			)
		}
	}

	rows = append(rows, "")
	rows = append(rows, common.Divider(w-4))
	rows = append(rows, "")

	// StrictNodes toggle
	strictLabel := common.StyleDim.Render("  StrictNodes:")
	strictVal := common.StyleError.Render("OFF")
	if c.strict {
		strictVal = common.StyleSuccess.Render("ON ")
	}
	rows = append(rows, strictLabel+" "+strictVal)
	rows = append(rows, common.StyleDim.Render("  Toggle with [s]"))
	rows = append(rows, "")
	rows = append(rows, common.StyleDim.Render("  StrictNodes ON means Tor will\n  only use nodes from selected\n  country. May reduce anonymity."))

	content := strings.Join(rows, "\n")
	return common.StylePanel.Width(w).Render(content)
}

func (c *CountriesView) renderActions(w int) string {
	hints := strings.Join([]string{
		common.KeyHint("↑↓", "navigate"),
		common.KeyHint("enter", "toggle"),
		common.KeyHint("s", "strict nodes"),
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

// ── Logic ─────────────────────────────────────────────────────────────────

func (c *CountriesView) toggleCountry() {
	entry := countryList[c.cursor]

	// "Any" clears selection
	if entry.Code == "" {
		c.selected = nil
		c.toast = "Country filter cleared"
		c.toastErr = false
		return
	}

	if c.isSelected(entry.Code) {
		// Deselect
		newSel := c.selected[:0]
		for _, code := range c.selected {
			if code != entry.Code {
				newSel = append(newSel, code)
			}
		}
		c.selected = newSel
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

// ── Commands ──────────────────────────────────────────────────────────────

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
			// Build {US},{DE} style
			codes := make([]string, len(selected))
			for i, code := range selected {
				codes[i] = "{" + code + "}"
			}
			nodes := strings.Join(codes, ",")
			err = ctrl.SetConf(ctx, "ExitNodes", nodes)
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
