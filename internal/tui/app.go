package tui

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/buffer"
	"github.com/alexandrmotologa/timberlens/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// AppModel represents the Bubble Tea TUI state.
type AppModel struct {
	ring        *buffer.RingBuffer
	query       *buffer.Query
	queryInput  string
	searching   bool
	following   bool
	paused      bool
	width       int
	height      int
	cursor      int
	showModal   bool
	expandStack bool
	selected    *model.LogRecord

	// Ingestion rate metrics
	lastCount uint64
	lastTime  time.Time
	rateLps   float64

	// Quick filter
	activeServices []string
	filteredSvc    string
}

// NewAppModel creates the TUI application model.
func NewAppModel(ring *buffer.RingBuffer) *AppModel {
	return &AppModel{
		ring:      ring,
		following: true,
		lastTime:  time.Now(),
	}
}

func (m *AppModel) Init() tea.Cmd {
	return tickEvery(100 * time.Millisecond)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		now := time.Time(msg)
		elapsed := now.Sub(m.lastTime).Seconds()
		if elapsed >= 0.5 {
			curCount := m.ring.Total()
			diff := curCount - m.lastCount
			m.rateLps = float64(diff) / elapsed
			m.lastCount = curCount
			m.lastTime = now

			// Update active services list
			svcMap := m.ring.ServiceCounts()
			services := make([]string, 0, len(svcMap))
			for s := range svcMap {
				services = append(services, s)
			}
			m.activeServices = services
		}
		return m, tickEvery(100 * time.Millisecond)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// 1. Modal View Active
	if m.showModal {
		switch key {
		case "esc", "enter", "q":
			m.showModal = false
			return m, nil
		case "e":
			m.expandStack = !m.expandStack
			return m, nil
		}
		return m, nil
	}

	// 2. Search / Query Input Mode
	if m.searching {
		switch key {
		case "enter":
			m.searching = false
			m.query = buffer.ParseQuery(m.queryInput)
			m.following = true
			return m, nil
		case "esc":
			m.searching = false
			return m, nil
		case "backspace":
			if len(m.queryInput) > 0 {
				m.queryInput = m.queryInput[:len(m.queryInput)-1]
			}
			return m, nil
		default:
			if len(key) == 1 {
				m.queryInput += key
			}
			return m, nil
		}
	}

	// 3. Normal Viewport Controls
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "/":
		m.searching = true
		return m, nil
	case "f":
		m.following = !m.following
		return m, nil
	case "p":
		m.paused = !m.paused
		return m, nil
	case "c":
		m.query = nil
		m.queryInput = ""
		m.filteredSvc = ""
		return m, nil
	case "e":
		m.expandStack = !m.expandStack
		return m, nil
	case "enter":
		records := m.getVisibleRecords()
		if len(records) > 0 && m.cursor >= 0 && m.cursor < len(records) {
			m.selected = records[m.cursor]
			m.showModal = true
		}
		return m, nil
	case "j", "down":
		records := m.getVisibleRecords()
		if m.cursor < len(records)-1 {
			m.cursor++
		}
		if m.cursor == len(records)-1 {
			m.following = true
		}
		return m, nil
	case "k", "up":
		m.following = false
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "g":
		m.following = false
		m.cursor = 0
		return m, nil
	case "G":
		m.following = true
		records := m.getVisibleRecords()
		if len(records) > 0 {
			m.cursor = len(records) - 1
		}
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx := int(key[0] - '1')
		if idx < len(m.activeServices) {
			svc := m.activeServices[idx]
			if m.filteredSvc == svc {
				m.filteredSvc = "" // Toggle off
				m.query = nil
			} else {
				m.filteredSvc = svc
				m.query = buffer.ParseQuery("service:" + svc)
			}
		}
		return m, nil
	}

	return m, nil
}

func (m *AppModel) getVisibleRecords() []*model.LogRecord {
	logHeight := m.height - 4
	if logHeight <= 0 {
		logHeight = 20
	}

	var recs []*model.LogRecord
	if m.query != nil && m.query.Raw != "" {
		recs = m.ring.Filter(m.query, logHeight*3)
	} else {
		recs = m.ring.Recent(logHeight * 3)
	}

	// Slice to viewport height
	if len(recs) > logHeight {
		if m.following {
			recs = recs[len(recs)-logHeight:]
		} else {
			recs = recs[:logHeight]
		}
	}

	return recs
}

func (m *AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing TimberLens..."
	}

	// 1. Header Bar
	header := m.renderHeader()

	// 2. Main Logs Viewport or Modal
	var content string
	if m.showModal && m.selected != nil {
		modal := RenderDetailsModal(m.selected, m.width, m.height, m.expandStack)
		content = lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, modal)
	} else {
		content = m.renderLogs()
	}

	// 3. Status Footer Bar
	footer := m.renderFooter()

	return fmt.Sprintf("%s\n%s\n%s", header, content, footer)
}

func (m *AppModel) renderHeader() string {
	title := titleStyle.Render("TimberLens")

	// Service quick pills
	var pills []string
	for i, s := range m.activeServices {
		if i >= 6 {
			break
		}
		style := ServiceStyle(s)
		if m.filteredSvc == s {
			pills = append(pills, fmt.Sprintf("\033[7m[%d:%s]\033[0m", i+1, s))
		} else {
			pills = append(pills, style.Render(fmt.Sprintf("[%d:%s]", i+1, s)))
		}
	}
	pillsStr := strings.Join(pills, " ")

	rateStr := fmt.Sprintf("%.0f lines/s", m.rateLps)

	left := fmt.Sprintf("%s  %s", title, pillsStr)
	right := fmt.Sprintf("\033[90m%s\033[0m", rateStr)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m *AppModel) renderLogs() string {
	records := m.getVisibleRecords()
	logHeight := m.height - 4
	if logHeight <= 0 {
		logHeight = 20
	}

	if len(records) == 0 {
		emptyMsg := "\033[90mWaiting for piped logs... (e.g. docker compose logs -f | timberlens)\033[0m"
		return lipgloss.Place(m.width, logHeight, lipgloss.Center, lipgloss.Center, emptyMsg)
	}

	// Auto-adjust cursor if out of bounds or following
	if m.following {
		m.cursor = len(records) - 1
	} else if m.cursor >= len(records) {
		m.cursor = len(records) - 1
	}

	lines := make([]string, 0, len(records))

	for idx, rec := range records {
		ts := timestampStyle.Render(rec.Timestamp.Format("15:04:05.000"))
		lvl := LevelStyle(rec.Level).Render(fmt.Sprintf("%-5s", rec.LevelStr))
		svc := ServiceStyle(rec.Service).Render(fmt.Sprintf("%-12s", truncate(rec.Service, 12)))

		trace := ""
		if rec.TraceID != "" {
			trace = traceIdStyle.Render("[" + truncate(rec.TraceID, 8) + "] ")
		}

		msg := rec.Message
		if msg == "" {
			msg = rec.Raw
		}

		row := fmt.Sprintf("%s %s %s %s%s", ts, lvl, svc, trace, msg)

		// Stack trace indicator icon
		if rec.StackTrace != nil {
			row += fmt.Sprintf(" \033[31m[stack: -%d frames]\033[0m", rec.StackTrace.HiddenFrames)
		}

		// Cursor highlight
		if idx == m.cursor {
			row = cursorRowStyle.Width(m.width).Render(row)
		}

		lines = append(lines, row)
	}

	// Pad with empty lines if needed
	for len(lines) < logHeight {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (m *AppModel) renderFooter() string {
	if m.searching {
		return searchPromptStyle.Render("Search: ") + m.queryInput + "\033[5m_\033[0m" +
			statusBarStyle.Render("  [Enter] Apply  [Esc] Cancel")
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	memMb := float64(mem.Alloc) / 1024 / 1024

	var followStatus string
	if m.following {
		followStatus = statusActiveStyle.Render("FOLLOW: ON")
	} else {
		followStatus = statusPausedStyle.Render("FOLLOW: OFF")
	}

	total := fmt.Sprintf("Total: %d", m.ring.Total())
	memory := fmt.Sprintf("RAM: %.1f MB", memMb)

	queryNotice := ""
	if m.query != nil && m.query.Raw != "" {
		queryNotice = fmt.Sprintf("Filter: %q  ", m.query.Raw)
	}

	left := statusBarStyle.Render(fmt.Sprintf("%s  %s  %s  %s", total, memory, followStatus, queryNotice))
	right := statusBarStyle.Render("[/] Search  [f] Follow  [p] Pause  [c] Clear  [Enter] Inspect  [q] Quit")

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
