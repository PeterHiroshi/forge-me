package tail

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type OutputFormat string

const (
	FormatPretty  OutputFormat = "pretty"
	FormatJSON    OutputFormat = "json"
	FormatCompact OutputFormat = "compact"
)

type Formatter struct {
	format            OutputFormat
	noColor           bool
	IncludeLogs       bool
	IncludeExceptions bool
}

func NewFormatter(format string, noColor bool) *Formatter {
	if noColor {
		color.NoColor = true
	} else {
		color.NoColor = false
	}
	return &Formatter{
		format:            OutputFormat(format),
		noColor:           noColor,
		IncludeLogs:       true,
		IncludeExceptions: true,
	}
}

func (f *Formatter) Format(event TailEvent) string {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(event)
	case FormatCompact:
		return f.formatCompact(event)
	default:
		return f.formatPretty(event)
	}
}

func (f *Formatter) formatJSON(event TailEvent) string {
	jsonBytes, _ := json.Marshal(event)
	return string(jsonBytes)
}

func (f *Formatter) colorize(fn func(format string, a ...interface{}) string, format string, a ...interface{}) string {
	if f.noColor {
		return fmt.Sprintf(format, a...)
	}
	return fn(format, a...)
}

func (f *Formatter) formatCompact(event TailEvent) string {
	var status string
	if event.Outcome == "ok" {
		status = f.colorize(color.GreenString, "OK")
	} else {
		status = f.colorize(color.RedString, "ERROR")
	}

	ts := event.Time().Format("15:04:05")

	return fmt.Sprintf("[%s] %s %s %s %d",
		f.colorize(color.BlueString, "%s", ts),
		status,
		f.colorize(color.YellowString, "%s", event.Event.Request.Method),
		f.colorize(color.CyanString, "%s", event.Event.Request.URL),
		event.Event.Response.Status,
	)
}

func (f *Formatter) formatPretty(event TailEvent) string {
	var builder strings.Builder

	// Header
	var status string
	if event.Outcome == "ok" {
		status = f.colorize(color.GreenString, "✓ OK")
	} else {
		status = f.colorize(color.RedString, "✗ ERROR")
	}

	builder.WriteString(fmt.Sprintf("%s %s %s\n",
		f.colorize(color.CyanString, "➤"),
		f.colorize(color.HiWhiteString, "%s", event.ScriptName),
		status,
	))

	// Request details
	builder.WriteString(fmt.Sprintf("  %s %s [%d]\n",
		f.colorize(color.YellowString, "%s", event.Event.Request.Method),
		f.colorize(color.BlueString, "%s", event.Event.Request.URL),
		event.Event.Response.Status,
	))

	// Logs
	if f.IncludeLogs && len(event.Logs) > 0 {
		builder.WriteString("  Logs:\n")
		for _, log := range event.Logs {
			builder.WriteString(fmt.Sprintf("    • %s\n", strings.Join(log.Message, " ")))
		}
	}

	// Exceptions
	if f.IncludeExceptions && len(event.Exceptions) > 0 {
		builder.WriteString("  Exceptions:\n")
		for _, ex := range event.Exceptions {
			builder.WriteString(fmt.Sprintf("    ✗ %s: %s\n",
				f.colorize(color.RedString, "%s", ex.Name),
				f.colorize(color.RedString, "%s", ex.Message),
			))
		}
	}

	return builder.String()
}
