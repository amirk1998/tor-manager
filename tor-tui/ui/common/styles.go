package common

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors
const (
	ColorBg       = "#1a1a2e"
	ColorBgPanel  = "#16213e"
	ColorFg       = "#eaeaea"
	ColorWhite    = "#ffffff"
	ColorGray     = "#6272a4"
	ColorPurple   = "#bd93f9"
	ColorPurpleSub = "#9d7bcf"
	ColorGreen    = "#50fa7b"
	ColorYellow   = "#f1fa8c"
	ColorRed      = "#ff5555"
	ColorOrange   = "#ffb86c"
	ColorBlue     = "#8be9fd"
)

// Styles
var (
	StylePrimary   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple))
	StyleAccent    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurpleSub))
	StyleDim       = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray))
	StyleSuccess   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen))
	StyleWarning   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow))
	StyleError     = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed))
	StyleInfo      = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue))

	StyleSuccessBold = StyleSuccess.Bold(true)
	StyleWarningBold = StyleWarning.Bold(true)
	StyleErrorBold   = StyleError.Bold(true)
	StyleBold        = lipgloss.NewStyle().Bold(true)

	StylePanel       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorBgPanel)).Padding(1, 2)
	StylePanelActive = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorPurple)).Padding(1, 2)

	StyleTabActive   = lipgloss.NewStyle().Background(lipgloss.Color(ColorPurple)).Foreground(lipgloss.Color(ColorWhite)).Padding(0, 1)
	StyleTabInactive = lipgloss.NewStyle().Background(lipgloss.Color(ColorBgPanel)).Foreground(lipgloss.Color(ColorGray)).Padding(0, 1)

	StyleRowNormal   = lipgloss.NewStyle()
	StyleRowSelected = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple))

	StyleInputFocused  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(ColorPurple)).Padding(0, 1)
	StyleInputBlurred  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(ColorGray)).Padding(0, 1)

	StyleInputLabel = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurpleSub)).Bold(true)
	StyleTableHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurpleSub)).Bold(true)

	StyleCursor = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple)).Bold(true)
)

// SectionTitle renders a section header.
func SectionTitle(title string) string {
	return StyleBold.Render("  " + title)
}

// KeyHint renders a keyboard shortcut hint.
func KeyHint(key, desc string) string {
	return StyleDim.Render("[") + StyleAccent.Render(key) + StyleDim.Render("] ")+desc
}

// Divider renders a horizontal divider line.
func Divider(width int) string {
	return StyleDim.Render(strings.Repeat("─", width))
}

// StatusDot renders a status indicator.
func StatusDot(connected, bootstrapping bool) string {
	if connected {
		return StyleSuccessBold.Render("●  Connected")
	}
	if bootstrapping {
		return StyleWarning.Render("◐  Bootstrapping...")
	}
	return StyleError.Render("○  Disconnected")
}

// RenderProgressBar renders a progress bar.
func RenderProgressBar(progress, width int) string {
	filled := progress * width / 100
	empty := width - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return StyleDim.Render("[") + StyleSuccess.Render(bar) + StyleDim.Render(fmt.Sprintf(" %d%%", progress))+"]"
}

// CensorshipBadge renders a censorship resistance badge.
func CensorshipBadge(level int) string {
	bg := lipgloss.Color(ColorGray)
	label := "LOW"
	switch {
	case level >= 5:
		bg = lipgloss.Color(ColorGreen)
		label = "MAX"
	case level >= 4:
		bg = lipgloss.Color(ColorBlue)
		label = "HIGH"
	case level >= 3:
		bg = lipgloss.Color(ColorYellow)
		label = "MED"
	default:
		bg = lipgloss.Color(ColorRed)
		label = "LOW"
	}
	return lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color(ColorWhite)).Padding(0, 1).Render(label)
}

// SpeedBadge renders a speed impact badge.
func SpeedBadge(impact int) string {
	bg := lipgloss.Color(ColorGray)
	label := "MED"
	switch {
	case impact <= 2:
		bg = lipgloss.Color(ColorGreen)
		label = "FAST"
	case impact == 3:
		bg = lipgloss.Color(ColorYellow)
		label = "MED"
	default:
		bg = lipgloss.Color(ColorRed)
		label = "SLOW"
	}
	return lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color(ColorWhite)).Padding(0, 1).Render(label)
}
