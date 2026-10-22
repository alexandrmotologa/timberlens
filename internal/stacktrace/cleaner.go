package stacktrace

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

// DetectAndClean parses multi-line text to detect stack traces and collapse vendor frames.
func DetectAndClean(text string) *model.StackTraceInfo {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return nil
	}

	frames := make([]model.StackFrame, 0, len(lines))
	lang := ""
	rootCause := ""
	hiddenCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// 1. Try Node.js frame
		if matches := nodeFrameRegex.FindStringSubmatch(line); len(matches) > 0 {
			lang = "nodejs"
			fn := strings.TrimSpace(matches[1])
			file := matches[2]
			lineNoStr := matches[3]
			if file == "" {
				file = matches[5]
				lineNoStr = matches[6]
			}
			lineNo, _ := strconv.Atoi(lineNoStr)
			isVendor := IsVendorPath(file)

			if isVendor {
				hiddenCount++
			} else if rootCause == "" {
				rootCause = fmt.Sprintf("%s:%d (%s)", file, lineNo, fn)
			}

			frames = append(frames, model.StackFrame{
				File:     file,
				Line:     lineNo,
				Function: fn,
				IsVendor: isVendor,
				Raw:      line,
			})
			continue
		}

		// 2. Try Python frame
		if matches := pythonFrameRegex.FindStringSubmatch(line); len(matches) > 0 {
			lang = "python"
			file := matches[1]
			lineNo, _ := strconv.Atoi(matches[2])
			fn := matches[3]
			isVendor := IsVendorPath(file)

			if isVendor {
				hiddenCount++
			} else if rootCause == "" {
				rootCause = fmt.Sprintf("%s:%d in %s", file, lineNo, fn)
			}

			frames = append(frames, model.StackFrame{
				File:     file,
				Line:     lineNo,
				Function: fn,
				IsVendor: isVendor,
				Raw:      line,
			})
			continue
		}

		// 3. Try Java frame
		if matches := javaFrameRegex.FindStringSubmatch(line); len(matches) > 0 {
			lang = "java"
			fn := matches[1]
			file := matches[2]
			lineNo, _ := strconv.Atoi(matches[3])
			isVendor := IsVendorPath(fn) || IsVendorPath(file)

			if isVendor {
				hiddenCount++
			} else if rootCause == "" {
				rootCause = fmt.Sprintf("%s(%s:%d)", fn, file, lineNo)
			}

			frames = append(frames, model.StackFrame{
				File:     file,
				Line:     lineNo,
				Function: fn,
				IsVendor: isVendor,
				Raw:      line,
			})
			continue
		}

		// 4. Try Go frame
		if strings.HasPrefix(line, "\t") && strings.Contains(line, ".go:") {
			if matches := goFrameRegex.FindStringSubmatch(line); len(matches) > 0 {
				lang = "go"
				file := matches[1]
				lineNo, _ := strconv.Atoi(matches[2])
				isVendor := IsVendorPath(file)

				if isVendor {
					hiddenCount++
				} else if rootCause == "" {
					rootCause = fmt.Sprintf("%s:%d", file, lineNo)
				}

				frames = append(frames, model.StackFrame{
					File:     file,
					Line:     lineNo,
					IsVendor: isVendor,
					Raw:      line,
				})
				continue
			}
		}
	}

	if len(frames) == 0 {
		return nil
	}

	if rootCause == "" && len(frames) > 0 {
		rootCause = fmt.Sprintf("%s:%d", frames[0].File, frames[0].Line)
	}

	return &model.StackTraceInfo{
		Raw:          text,
		Language:     lang,
		RootCause:    rootCause,
		TotalFrames:  len(frames),
		HiddenFrames: hiddenCount,
		Frames:       frames,
	}
}

// FormatFolded returns a human-readable folded representation of the stack trace.
func FormatFolded(info *model.StackTraceInfo) string {
	if info == nil || len(info.Frames) == 0 {
		return ""
	}

	var sb strings.Builder
	vendorRun := 0

	for _, frame := range info.Frames {
		if frame.IsVendor {
			vendorRun++
			continue
		}

		if vendorRun > 0 {
			sb.WriteString(fmt.Sprintf("    \033[90m[+ %d library frames hidden]\033[0m\n", vendorRun))
			vendorRun = 0
		}

		// Highlight user frame
		sb.WriteString(fmt.Sprintf("    \033[36m--> %s:%d\033[0m \033[33m(%s)\033[0m\n", frame.File, frame.Line, frame.Function))
	}

	if vendorRun > 0 {
		sb.WriteString(fmt.Sprintf("    \033[90m[+ %d library frames hidden]\033[0m\n", vendorRun))
	}

	return sb.String()
}
