package tui

import (
	"hash/fnv"

	"github.com/alexandrmotologa/timberlens/internal/model"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#2563EB")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Background(lipgloss.Color("#0F172A")).
			Padding(0, 1)

	statusActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#10B981"))

	statusPausedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#EF4444"))

	searchPromptStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#38BDF8"))

	cursorRowStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E293B"))

	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	traceIdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#818CF8"))

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#38BDF8")).
			Background(lipgloss.Color("#0F172A")).
			Padding(1, 2)

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#38BDF8")).
			MarginBottom(1)
)

// LevelStyle returns a Lip Gloss style for a severity level.
func LevelStyle(l model.SeverityLevel) lipgloss.Style {
	switch l {
	case model.LevelTrace:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	case model.LevelDebug:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4"))
	case model.LevelInfo:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	case model.LevelWarn:
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))
	case model.LevelError:
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))
	case model.LevelFatal:
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EC4899"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1"))
	}
}

// ServicePalette provides deterministic colors for microservices.
var serviceColors = []string{
	"#38BDF8", // Sky Blue
	"#A855F7", // Purple
	"#EC4899", // Pink
	"#10B981", // Emerald
	"#F97316", // Orange
	"#EAB308", // Amber
	"#6366F1", // Indigo
	"#14B8A6", // Teal
	"#F43F5E", // Rose
	"#84CC16", // Lime
}

// ServiceColor returns a deterministic color string for a service name.
func ServiceColor(service string) string {
	if service == "" {
		return "#94A3B8"
	}
	h := fnv.New32a()
	h.Write([]byte(service))
	idx := int(h.Sum32()) % len(serviceColors)
	return serviceColors[idx]
}

// ServiceStyle returns a Lip Gloss style for a service name.
func ServiceStyle(service string) lipgloss.Style {
	colorHex := ServiceColor(service)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorHex))
}
