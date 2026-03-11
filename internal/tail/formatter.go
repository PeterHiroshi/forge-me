package tail

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type OutputFormat string

const (
	FormatPretty  OutputFormat = "pretty"
	FormatJSON    OutputFormat = "json"
	FormatCompact OutputFormat = "compact"
)

type Formatter struct {
	format OutputFormat
}

func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{format: format}
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

func (f *Formatter) formatCompact(event TailEvent) string {
	status := color.GreenString("OK")
	if event.Outcome != "ok" {
		status = color.RedString("ERROR")
	}

	return fmt.Sprintf("[%s] %s %s %s %s",
		color.BlueString(time.Unix(0, event.EventTimestamp*int64(time.Millisecond)).Format(time.RFC3339)),
		status,
		color.YellowString(event.Event.Request.Method),
		color.CyanString(event.Event.Request.URL),
		color.MagentaString(fmt.Sprintf("%d", event.EventTimestamp)),
	)
}

func (f *Formatter) formatPretty(event TailEvent) string {
	var builder strings.Builder

	// Header
	status := color.GreenString("✓ OK")
	if event.Outcome != "ok" {
		status = color.RedString("✗ ERROR")
	}

	builder.WriteString(fmt.Sprintf("%s %s %s\n",
		color.CyanString("➤"),
		color.HiWhiteString(event.ScriptName),
		status,
	))

	// Request details
	builder.WriteString(fmt.Sprintf("  %s %s\n",
		color.YellowString(event.Event.Request.Method),
		color.BlueString(event.Event.Request.URL),
	))

	// Logs
	if len(event.Logs) > 0 {
		builder.WriteString("  Logs:\n")
		for _, log := range event.Logs {
			builder.WriteString(fmt.Sprintf("    • %s\n", color.WhiteString(strings.Join(log.Message, " "))))
		}
	}

	// Exceptions
	if len(event.Exceptions) > 0 {
		builder.WriteString("  Exceptions:\n")
		for _, ex := range event.Exceptions {
			builder.WriteString(fmt.Sprintf("    ✗ %s: %s\n", 
				color.RedString(ex.Name), 
				color.RedString(ex.Message),
			))
		}
	}

	return builder.String()
}
