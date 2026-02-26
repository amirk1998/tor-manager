package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Palette ────────────────────────────────────────────────────────────────

const (
	ColorPurple    = lipgloss.Color("#7C3AED")
	ColorPurpleDim = lipgloss.Color("#4C1D95")
	ColorPurpleSub = lipgloss.Color("#A78BFA")
	ColorGreen     = lipgloss.Color("#10B981")
	ColorRed       = lipgloss.Color("#EF4444")
	ColorYellow    = lipgloss.Color("#F59E0B")
	ColorBlue      = lipgloss.Color("#60A5FA")
	ColorDim       = lipgloss.Color("#6B7280")
	ColorSubtle    = lipgloss.Color("#374151")
	ColorWhite     = lipgloss.Color("#F9FAFB")
	ColorBg        = lipgloss.Color("#111827")
	ColorBgPanel   = lipgloss.Color("#1F2937")
	ColorBorder    = lipgloss.Color("#4B5563")
)

// ── Base Text ──────────────────────────────────────────────────────────────

var (
	StyleNormal  = lipgloss.NewStyle().Foreground(ColorWhite)
	StyleDim     = lipgloss.NewStyle().Foreground(ColorDim)
	StyleBold    = lipgloss.NewStyle().Foreground(ColorWhite).Bold(true)
	StylePrimary = lipgloss.NewStyle().Foreground(ColorPurple)
	StyleAccent  = lipgloss.NewStyle().Foreground(ColorPurpleSub)

	StyleSuccess     = lipgloss.NewStyle().Foreground(ColorGreen)
	StyleError       = lipgloss.NewStyle().Foreground(ColorRed)
	StyleWarning     = lipgloss.NewStyle().Foreground(ColorYellow)
	StyleInfo        = lipgloss.NewStyle().Foreground(ColorBlue)
	StyleSuccessBold = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	StyleErrorBold   = lipgloss.NewStyle().Foreground(ColorRed).Bold(true)
	StyleWarningBold = lipgloss.NewStyle().Foreground(ColorYellow).Bold(true)
)

// ── Panels ─────────────────────────────────────────────────────────────────

var (
	StylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	StylePanelActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPurple).
				Padding(0, 1)

	StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 2)
)

// ── Tabs ───────────────────────────────────────────────────────────────────

var (
	StyleTabActive = lipgloss.NewStyle().
			Foreground(ColorBg).
			Background(ColorPurple).
			Bold(true).
			Padding(0, 2)

	StyleTabInactive = lipgloss.NewStyle().
				Foreground(ColorDim).
				Padding(0, 2)

	StyleTabBar = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorSubtle)
)

// ── Table / List ──────────────────────────────────────────────────────────

var (
	StyleTableHeader = lipgloss.NewStyle().
				Foreground(ColorDim).
				Bold(true)

	StyleRowSelected = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorPurpleDim).
				Bold(true)

	StyleRowNormal = lipgloss.NewStyle().Foreground(ColorWhite)
	StyleRowDim    = lipgloss.NewStyle().Foreground(ColorDim)
)

// ── Footer ────────────────────────────────────────────────────────────────

var StyleFooter = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder(), true, false, false, false).
	BorderForeground(ColorSubtle).
	Foreground(ColorDim)

// ── Helpers ───────────────────────────────────────────────────────────────

// StatusDot renders a coloured connection status indicator.
func StatusDot(connected, bootstrapping bool) string {
	switch {
	case bootstrapping:
		return StyleWarningBold.Render("◉ Bootstrapping")
	case connected:
		return StyleSuccessBold.Render("● Connected")
	default:
		return StyleErrorBold.Render("○ Disconnected")
	}
}

// RenderProgressBar renders a compact ASCII progress bar.
func RenderProgressBar(percent, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := width * percent / 100
	empty := width - filled

	bar := lipgloss.NewStyle().Foreground(ColorPurple).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(ColorSubtle).Render(strings.Repeat("░", empty))

	label := lipgloss.NewStyle().Foreground(ColorPurpleSub).
		Render(fmt.Sprintf("%3d%%", percent))

	return bar + " " + label
}

// CensorshipBadge returns a coloured censorship-resistance badge.
func CensorshipBadge(level int) string {
	base := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	switch {
	case level >= 5:
		return base.Foreground(ColorBg).Background(ColorGreen).Render("MAX")
	case level >= 4:
		return base.Foreground(ColorBg).Background(ColorBlue).Render("HIGH")
	case level >= 3:
		return base.Foreground(ColorBg).Background(ColorYellow).Render("MED")
	default:
		return base.Foreground(ColorBg).Background(ColorRed).Render("LOW")
	}
}

// SpeedBadge returns a coloured speed-impact badge.
func SpeedBadge(impact int) string {
	base := lipgloss.NewStyle().Padding(0, 1)
	switch {
	case impact <= 2:
		return base.Foreground(ColorBg).Background(ColorGreen).Render("FAST")
	case impact == 3:
		return base.Foreground(ColorBg).Background(ColorYellow).Render("MED")
	default:
		return base.Foreground(ColorBg).Background(ColorRed).Render("SLOW")
	}
}

// KeyHint renders a key + description hint, e.g. "[n]  New identity".
func KeyHint(key, desc string) string {
	k := lipgloss.NewStyle().Foreground(ColorPurpleSub).Bold(true).Render("[" + key + "]")
	d := StyleDim.Render(" " + desc)
	return k + d
}

// Divider renders a horizontal rule of the given width.
func Divider(width int) string {
	return StyleDim.Render(strings.Repeat("─", width))
}

// SectionTitle renders a prominent section heading.
func SectionTitle(title string) string {
	return lipgloss.NewStyle().
		Foreground(ColorPurpleSub).
		Bold(true).
		Render("▸ " + title)
}
