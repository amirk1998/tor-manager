package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/amirk1998/tor-manager/tor-tui/ui/common"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxLogLines = 500

// LogEntry is a single timestamped log line.
type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

// LogsView is a scrollable viewport of Tor daemon log output.
type LogsView struct {
	width  int
	height int

	entries  []LogEntry
	viewport viewport.Model
	follow   bool
	filter   string
}

func NewLogsView() *LogsView {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(common.ColorWhite))
	return &LogsView{
		viewport: vp,
		follow:   true,
	}
}

func (l *LogsView) SetSize(w, h int) {
	l.width = w
	l.height = h

	// پنل header: 1 خط محتوا + 2 padding + 2 border = ~5 خط
	// پنل viewport: 2 padding + 2 border = 4 خط overhead
	// footer (hints): 1 خط
	// مجموع overhead: 5 + 4 + 1 = 10 خط
	l.viewport.Width = w - 6 // border(1)*2 + padding(2)*2 = 6
	l.viewport.Height = h - 10
	if l.viewport.Height < 5 {
		l.viewport.Height = 5
	}
	l.rebuildViewport()
}

func (l *LogsView) Init() tea.Cmd { return nil }

func (l *LogsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case common.LogLineMsg:
		l.appendLine(msg.Line)

	case common.BootstrapEventMsg:
		entry := LogEntry{
			Time:    time.Now(),
			Level:   "notice",
			Message: fmt.Sprintf("[bootstrap %d%%] %s", msg.Event.Progress, msg.Event.Summary),
		}
		l.addEntry(entry)

	case tea.KeyMsg:
		switch msg.String() {
		case "f":
			l.follow = !l.follow
		case "c":
			l.entries = nil
			l.rebuildViewport()
		case "up", "k":
			l.follow = false
			l.viewport.LineUp(1)
		case "down", "j":
			l.viewport.LineDown(1)
		case "ctrl+u", "pgup":
			l.follow = false
			l.viewport.HalfViewUp()
		case "ctrl+d", "pgdown":
			l.viewport.HalfViewDown()
		case "g":
			l.follow = false
			l.viewport.GotoTop()
		case "G":
			l.follow = true
			l.viewport.GotoBottom()
		}
	}

	var cmd tea.Cmd
	l.viewport, cmd = l.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return l, tea.Batch(cmds...)
}

func (l *LogsView) View() string {
	if l.width == 0 {
		return ""
	}

	w := l.width - 4
	if w > 100 {
		w = 100
	}

	header := l.renderHeader(w)

	// ❌ قبلاً: .Height(l.viewport.Height + 2) — باعث clip شدن و overflow میشد
	// ✅ الان: بدون Height ثابت — viewport خودش ارتفاع رو تعیین میکنه
	vpContent := common.StylePanel.
		Width(w).
		Render(l.viewport.View())

	footer := l.renderFooter(w)

	return lipgloss.JoinVertical(lipgloss.Left, header, vpContent, footer)
}

func (l *LogsView) renderHeader(w int) string {
	count := common.StyleDim.Render(fmt.Sprintf("%d lines", len(l.entries)))

	followStatus := common.StyleDim.Render(" ○ paused")
	if l.follow {
		followStatus = common.StyleSuccess.Render(" ● FOLLOW")
	}

	scrollPct := ""
	if l.viewport.TotalLineCount() > 0 {
		pct := l.viewport.ScrollPercent() * 100
		scrollPct = common.StyleDim.Render(fmt.Sprintf(" %3.0f%%", pct))
	}

	title := common.SectionTitle("Tor Daemon Logs")
	right := count + followStatus + scrollPct

	titleW := lipgloss.Width(title)
	rightW := lipgloss.Width(right)
	gap := w - titleW - rightW - 4
	if gap < 1 {
		gap = 1
	}

	return common.StylePanel.Width(w).Render(
		title + strings.Repeat(" ", gap) + right,
	)
}

func (l *LogsView) renderFooter(w int) string {
	return strings.Join([]string{
		common.KeyHint("↑↓", "scroll"),
		common.KeyHint("pgup/dn", "page"),
		common.KeyHint("g/G", "top/bottom"),
		common.KeyHint("f", "follow"),
		common.KeyHint("c", "clear"),
	}, "  ")
}

func (l *LogsView) appendLine(raw string) {
	l.addEntry(parseLogLine(raw))
}

func (l *LogsView) addEntry(entry LogEntry) {
	l.entries = append(l.entries, entry)
	if len(l.entries) > maxLogLines {
		l.entries = l.entries[len(l.entries)-maxLogLines:]
	}
	l.rebuildViewport()
	if l.follow {
		l.viewport.GotoBottom()
	}
}

func (l *LogsView) rebuildViewport() {
	var sb strings.Builder
	for _, e := range l.entries {
		sb.WriteString(l.formatEntry(e))
		sb.WriteByte('\n')
	}
	l.viewport.SetContent(sb.String())
}

func (l *LogsView) formatEntry(e LogEntry) string {
	ts := common.StyleDim.Render(e.Time.Format("15:04:05"))

	var levelStr string
	switch e.Level {
	case "err", "error":
		levelStr = common.StyleErrorBold.Render("[ERR]   ")
	case "warn", "warning":
		levelStr = common.StyleWarningBold.Render("[WARN]  ")
	case "notice":
		levelStr = common.StyleInfo.Render("[notice]")
	case "debug":
		levelStr = common.StyleDim.Render("[debug] ")
	default:
		levelStr = common.StyleDim.Render("[info]  ")
	}

	return fmt.Sprintf("%s %s %s", ts, levelStr, colorizeLogMessage(e.Message))
}

func colorizeLogMessage(msg string) string {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "bootstrap") && strings.Contains(lower, "100%"):
		return common.StyleSuccessBold.Render(msg)
	case strings.Contains(lower, "bootstrap"):
		return common.StyleWarning.Render(msg)
	case strings.Contains(lower, "error") || strings.Contains(lower, "failed"):
		return common.StyleError.Render(msg)
	case strings.Contains(lower, "warn"):
		return common.StyleWarning.Render(msg)
	case strings.Contains(lower, "circuit") && strings.Contains(lower, "built"):
		return common.StyleSuccess.Render(msg)
	case strings.Contains(lower, "new identity"):
		return common.StyleAccent.Render(msg)
	default:
		return common.StyleDim.Render(msg)
	}
}

func parseLogLine(raw string) LogEntry {
	entry := LogEntry{Time: time.Now(), Level: "notice", Message: raw}
	if start := strings.Index(raw, "["); start >= 0 {
		if end := strings.Index(raw[start:], "]"); end >= 0 {
			entry.Level = raw[start+1 : start+end]
			entry.Message = strings.TrimSpace(raw[start+end+2:])
		}
	}
	return entry
}
