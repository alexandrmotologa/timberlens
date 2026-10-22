package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/timberlens/internal/model"
	"github.com/alexandrmotologa/timberlens/internal/stacktrace"
)

// RenderDetailsModal produces a formatted popup inspecting a single LogRecord.
func RenderDetailsModal(rec *model.LogRecord, width, height int, expandStack bool) string {
	if rec == nil {
		return ""
	}

	modalW := width - 10
	if modalW < 50 {
		modalW = 50
	}
	if modalW > 120 {
		modalW = 120
	}

	modalH := height - 6
	if modalH < 15 {
		modalH = 15
	}

	var sb strings.Builder

	// Title / Summary bar
	sb.WriteString(modalTitleStyle.Render("Log Event Inspector (ID: "+fmt.Sprint(rec.ID)+")") + "\n\n")

	sb.WriteString(fmt.Sprintf("\033[90mTimestamp:\033[0m  %s\n", rec.Timestamp.Format("2006-01-02 15:04:05.000000")))
	sb.WriteString(fmt.Sprintf("\033[90mSeverity:\033[0m   %s\n", LevelStyle(rec.Level).Render(rec.LevelStr)))
	sb.WriteString(fmt.Sprintf("\033[90mService:\033[0m    %s\n", ServiceStyle(rec.Service).Render(rec.Service)))

	if rec.TraceID != "" {
		sb.WriteString(fmt.Sprintf("\033[90mTrace ID:\033[0m   \033[35m%s\033[0m\n", rec.TraceID))
	}
	if rec.SpanID != "" {
		sb.WriteString(fmt.Sprintf("\033[90mSpan ID:\033[0m    \033[34m%s\033[0m\n", rec.SpanID))
	}
	sb.WriteString(fmt.Sprintf("\033[90mFormat:\033[0m     %s\n", rec.FormatStr))

	sb.WriteString("\n\033[1mMessage:\033[0m\n")
	sb.WriteString("  " + rec.Message + "\n")

	// Stack trace section
	if rec.StackTrace != nil {
		sb.WriteString("\n\033[1;31mStack Trace:\033[0m (Press 'e' to toggle full expansion)\n")
		if expandStack {
			sb.WriteString(rec.StackTrace.Raw + "\n")
		} else {
			sb.WriteString(stacktrace.FormatFolded(rec.StackTrace))
		}
	}

	// Structured Fields
	if len(rec.Fields) > 0 {
		sb.WriteString("\n\033[1mStructured Fields:\033[0m\n")
		prettyJSON, err := json.MarshalIndent(rec.Fields, "  ", "  ")
		if err == nil {
			sb.WriteString("  " + string(prettyJSON) + "\n")
		}
	}

	sb.WriteString("\n\033[90m[Esc/Enter] Close Inspector    [e] Toggle Stack Trace\033[0m")

	content := sb.String()
	lines := strings.Split(content, "\n")
	if len(lines) > modalH {
		content = strings.Join(lines[:modalH], "\n")
	}

	return modalStyle.Width(modalW).Render(content)
}
