package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// FormatJSON converts data to pretty-printed JSON
func FormatJSON(data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FormatTable formats data as a table
func FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Print headers
	for i, h := range headers {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(fmt.Sprintf("%-*s", widths[i], h))
	}
	sb.WriteString("\n")

	// Print separator
	for i, w := range widths {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(strings.Repeat("-", w))
	}
	sb.WriteString("\n")

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				sb.WriteString("  ")
			}
			if i < len(widths) {
				sb.WriteString(fmt.Sprintf("%-*s", widths[i], cell))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatColoredTable formats data as a table with color support
func FormatColoredTable(headers []string, rows [][]string, enableColors bool) string {
	if len(headers) == 0 {
		return ""
	}

	// Disable colors if requested
	if !enableColors {
		color.NoColor = true
	} else {
		color.NoColor = false
	}

	// Calculate column widths (using plain text length)
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Print headers
	for i, h := range headers {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(fmt.Sprintf("%-*s", widths[i], h))
	}
	sb.WriteString("\n")

	// Print separator
	for i, w := range widths {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(strings.Repeat("-", w))
	}
	sb.WriteString("\n")

	// Print rows with colors
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				sb.WriteString("  ")
			}
			if i < len(widths) {
				// Apply color based on cell content
				coloredCell := colorizeCell(cell)
				// Calculate padding based on original cell length
				padding := widths[i] - len(cell)
				sb.WriteString(coloredCell)
				if padding > 0 {
					sb.WriteString(strings.Repeat(" ", padding))
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// colorizeCell applies color to a cell based on its content
func colorizeCell(cell string) string {
	lower := strings.ToLower(cell)

	// Green for healthy/active status
	if lower == "active" || lower == "healthy" || lower == "true" {
		return color.GreenString(cell)
	}

	// Red for errors/failures
	if lower == "error" || lower == "failed" || lower == "false" || lower == "invalid" {
		return color.RedString(cell)
	}

	// Yellow for warnings/pending
	if lower == "warning" || lower == "pending" || lower == "degraded" {
		return color.YellowString(cell)
	}

	// Default: no color
	return cell
}

// FormatJSONL converts data to JSON Lines format (one JSON object per line)
func FormatJSONL(data interface{}) (string, error) {
	// Check if data is a slice
	switch v := data.(type) {
	case []interface{}:
		var sb strings.Builder
		for _, item := range v {
			b, err := json.Marshal(item)
			if err != nil {
				return "", err
			}
			sb.Write(b)
			sb.WriteString("\n")
		}
		return sb.String(), nil
	default:
		// For single objects, just marshal once
		b, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(b) + "\n", nil
	}
}

// FormatCSV formats data as CSV
func FormatCSV(headers []string, rows [][]string, includeHeader bool) string {
	if len(headers) == 0 {
		return ""
	}

	var sb strings.Builder

	// Print headers if requested
	if includeHeader {
		for i, h := range headers {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(escapeCSV(h))
		}
		sb.WriteString("\n")
	}

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				sb.WriteString(",")
			}
			if i < len(headers) {
				sb.WriteString(escapeCSV(cell))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// escapeCSV escapes a CSV field value
func escapeCSV(s string) string {
	// If the string contains comma, quote, or newline, wrap in quotes and escape quotes
	if strings.ContainsAny(s, ",\"\n") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}

// FilterFields filters a slice of maps to only include specified fields
func FilterFields(data []map[string]interface{}, fields []string) []map[string]interface{} {
	if len(fields) == 0 {
		return data
	}

	// Create a set of requested fields for fast lookup
	fieldSet := make(map[string]bool)
	for _, f := range fields {
		fieldSet[f] = true
	}

	filtered := make([]map[string]interface{}, len(data))
	for i, item := range data {
		filtered[i] = make(map[string]interface{})
		for key, value := range item {
			if fieldSet[key] {
				filtered[i][key] = value
			}
		}
	}

	return filtered
}

// FilterTableFields filters headers and rows to only include specified field indices
func FilterTableFields(headers []string, rows [][]string, fields []string) ([]string, [][]string) {
	if len(fields) == 0 {
		return headers, rows
	}

	// Create a map of header name to index
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(h)] = i
	}

	// Determine which indices to keep
	indices := []int{}
	filteredHeaders := []string{}
	for _, field := range fields {
		if idx, ok := headerMap[strings.ToLower(field)]; ok {
			indices = append(indices, idx)
			filteredHeaders = append(filteredHeaders, headers[idx])
		}
	}

	// Filter rows
	filteredRows := make([][]string, len(rows))
	for i, row := range rows {
		filteredRow := make([]string, len(indices))
		for j, idx := range indices {
			if idx < len(row) {
				filteredRow[j] = row[idx]
			}
		}
		filteredRows[i] = filteredRow
	}

	return filteredHeaders, filteredRows
}
