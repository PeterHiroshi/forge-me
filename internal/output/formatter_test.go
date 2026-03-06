package output

import (
	"encoding/json"
	"strings"
	"testing"
)

type testData struct {
	ID   string
	Name string
	CPU  int
}

func TestFormatJSON(t *testing.T) {
	data := []testData{
		{ID: "1", Name: "test1", CPU: 100},
		{ID: "2", Name: "test2", CPU: 200},
	}

	result, err := FormatJSON(data)
	if err != nil {
		t.Fatalf("FormatJSON() error = %v", err)
	}

	// Verify it's valid JSON
	var parsed []testData
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("parsed length = %d, want 2", len(parsed))
	}
}

func TestFormatTable(t *testing.T) {
	headers := []string{"ID", "Name", "CPU"}
	rows := [][]string{
		{"1", "test1", "100"},
		{"2", "test2", "200"},
	}

	result := FormatTable(headers, rows)

	// Verify headers are present
	if !strings.Contains(result, "ID") {
		t.Errorf("result missing ID header")
	}
	if !strings.Contains(result, "Name") {
		t.Errorf("result missing Name header")
	}
	if !strings.Contains(result, "CPU") {
		t.Errorf("result missing CPU header")
	}

	// Verify data rows are present
	if !strings.Contains(result, "test1") {
		t.Errorf("result missing test1 data")
	}
	if !strings.Contains(result, "test2") {
		t.Errorf("result missing test2 data")
	}
}

func TestFormatTable_Empty(t *testing.T) {
	headers := []string{"ID", "Name"}
	rows := [][]string{}

	result := FormatTable(headers, rows)

	// Should still show headers
	if !strings.Contains(result, "ID") {
		t.Errorf("result missing ID header")
	}
}

func TestFormatColoredTable_WithColors(t *testing.T) {
	headers := []string{"Status", "Name", "Value"}
	rows := [][]string{
		{"active", "Service 1", "100"},
		{"error", "Service 2", "200"},
		{"warning", "Service 3", "300"},
	}

	result := FormatColoredTable(headers, rows, true)

	// Verify ANSI escape codes are present (colors enabled)
	if !strings.Contains(result, "\x1b[") {
		t.Errorf("result should contain ANSI escape codes when colors enabled")
	}

	// Verify data is present
	if !strings.Contains(result, "Service 1") {
		t.Errorf("result missing Service 1 data")
	}
}

func TestFormatColoredTable_WithoutColors(t *testing.T) {
	headers := []string{"Status", "Name", "Value"}
	rows := [][]string{
		{"active", "Service 1", "100"},
		{"error", "Service 2", "200"},
	}

	result := FormatColoredTable(headers, rows, false)

	// Verify no ANSI escape codes are present (colors disabled)
	if strings.Contains(result, "\x1b[") {
		t.Errorf("result should not contain ANSI escape codes when colors disabled")
	}

	// Verify data is still present
	if !strings.Contains(result, "Service 1") {
		t.Errorf("result missing Service 1 data")
	}
	if !strings.Contains(result, "active") {
		t.Errorf("result missing status data")
	}
}

func TestFormatColoredTable_ColorMapping(t *testing.T) {
	headers := []string{"Status"}
	rows := [][]string{
		{"active"},
		{"healthy"},
		{"error"},
		{"failed"},
		{"warning"},
		{"pending"},
		{"normal"},
	}

	result := FormatColoredTable(headers, rows, true)

	// Just verify the function runs and produces output
	if len(result) == 0 {
		t.Errorf("result should not be empty")
	}

	// Verify all status values are present
	for _, row := range rows {
		if !strings.Contains(result, row[0]) {
			t.Errorf("result missing status: %s", row[0])
		}
	}
}

func TestFormatJSON_Error(t *testing.T) {
	// Test with un-marshalable data (channel)
	ch := make(chan int)
	defer close(ch)

	_, err := FormatJSON(ch)
	if err == nil {
		t.Fatal("FormatJSON() with channel: error = nil, want error")
	}
}

func TestFormatJSON_ErrorWithFunc(t *testing.T) {
	// Test with function (also un-marshalable)
	fn := func() {}

	_, err := FormatJSON(fn)
	if err == nil {
		t.Fatal("FormatJSON() with function: error = nil, want error")
	}
}

func TestFormatTable_NilHeaders(t *testing.T) {
	rows := [][]string{
		{"value1", "value2"},
	}

	result := FormatTable(nil, rows)

	// Should return empty string with nil headers
	if result != "" {
		t.Errorf("FormatTable() with nil headers = %q, want empty string", result)
	}
}

func TestFormatTable_EmptyHeaders(t *testing.T) {
	headers := []string{}
	rows := [][]string{
		{"value1", "value2"},
	}

	result := FormatTable(headers, rows)

	// Should return empty string with empty headers
	if result != "" {
		t.Errorf("FormatTable() with empty headers = %q, want empty string", result)
	}
}

func TestFormatTable_MismatchedRowLengths(t *testing.T) {
	headers := []string{"Col1", "Col2", "Col3"}
	rows := [][]string{
		{"A", "B", "C"},        // matches header count
		{"X", "Y"},             // fewer columns
		{"1", "2", "3", "4"},   // more columns
	}

	result := FormatTable(headers, rows)

	// Should not crash and should contain all headers
	if !strings.Contains(result, "Col1") {
		t.Errorf("result missing Col1 header")
	}
	if !strings.Contains(result, "Col2") {
		t.Errorf("result missing Col2 header")
	}
	if !strings.Contains(result, "Col3") {
		t.Errorf("result missing Col3 header")
	}

	// Should contain data from rows
	if !strings.Contains(result, "A") {
		t.Errorf("result missing data from first row")
	}
	if !strings.Contains(result, "X") {
		t.Errorf("result missing data from second row")
	}
}

func TestFormatColoredTable_EmptyRows(t *testing.T) {
	headers := []string{"Header1", "Header2"}
	rows := [][]string{}

	result := FormatColoredTable(headers, rows, true)

	// Should show headers even with no rows
	if !strings.Contains(result, "Header1") {
		t.Errorf("result missing Header1")
	}
	if !strings.Contains(result, "Header2") {
		t.Errorf("result missing Header2")
	}

	// Should contain separator line
	if !strings.Contains(result, "-") {
		t.Errorf("result missing separator line")
	}
}

func TestFormatColoredTable_NilHeaders(t *testing.T) {
	rows := [][]string{
		{"value1", "value2"},
	}

	result := FormatColoredTable(nil, rows, true)

	// Should return empty string with nil headers
	if result != "" {
		t.Errorf("FormatColoredTable() with nil headers = %q, want empty string", result)
	}
}

func TestFormatColoredTable_MismatchedRowLengths(t *testing.T) {
	headers := []string{"Col1", "Col2", "Col3"}
	rows := [][]string{
		{"A", "B", "C"},
		{"X"},                // much shorter
		{"1", "2", "3", "4", "5"},  // much longer
	}

	result := FormatColoredTable(headers, rows, false)

	// Should not crash
	if len(result) == 0 {
		t.Error("result is empty, want non-empty")
	}

	// Should contain headers
	if !strings.Contains(result, "Col1") {
		t.Errorf("result missing Col1 header")
	}
}

func TestColorizeCell_CaseInsensitive(t *testing.T) {
	// Test that colorization is case-insensitive
	tests := []struct {
		input string
		desc  string
	}{
		{"ACTIVE", "uppercase active"},
		{"Active", "mixed case active"},
		{"ERROR", "uppercase error"},
		{"Error", "mixed case error"},
		{"WARNING", "uppercase warning"},
		{"Warning", "mixed case warning"},
	}

	headers := []string{"Status"}
	for _, tt := range tests {
		rows := [][]string{{tt.input}}
		result := FormatColoredTable(headers, rows, true)

		// Should contain the input (possibly with ANSI codes)
		if !strings.Contains(result, tt.input) {
			t.Errorf("result missing %s: %s", tt.desc, tt.input)
		}
	}
}

func TestFormatJSONL(t *testing.T) {
	data := []interface{}{
		testData{ID: "1", Name: "test1", CPU: 100},
		testData{ID: "2", Name: "test2", CPU: 200},
	}

	result, err := FormatJSONL(data)
	if err != nil {
		t.Fatalf("FormatJSONL() error = %v", err)
	}

	// Split into lines
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 2 {
		t.Errorf("number of lines = %d, want 2", len(lines))
	}

	// Each line should be valid JSON
	for i, line := range lines {
		var parsed testData
		if err := json.Unmarshal([]byte(line), &parsed); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestFormatJSONL_SingleObject(t *testing.T) {
	data := testData{ID: "1", Name: "test1", CPU: 100}

	result, err := FormatJSONL(data)
	if err != nil {
		t.Fatalf("FormatJSONL() error = %v", err)
	}

	// Should have one line
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 1 {
		t.Errorf("number of lines = %d, want 1", len(lines))
	}

	// Should be valid JSON
	var parsed testData
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Errorf("result is not valid JSON: %v", err)
	}

	if parsed.ID != "1" || parsed.Name != "test1" || parsed.CPU != 100 {
		t.Errorf("parsed data = %+v, want {ID:1 Name:test1 CPU:100}", parsed)
	}
}

func TestFormatCSV_WithHeader(t *testing.T) {
	headers := []string{"ID", "Name", "CPU"}
	rows := [][]string{
		{"1", "test1", "100"},
		{"2", "test2", "200"},
	}

	result := FormatCSV(headers, rows, true)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 3 {
		t.Errorf("number of lines = %d, want 3 (header + 2 rows)", len(lines))
	}

	// First line should be headers
	if lines[0] != "ID,Name,CPU" {
		t.Errorf("header line = %q, want %q", lines[0], "ID,Name,CPU")
	}

	// Data rows
	if lines[1] != "1,test1,100" {
		t.Errorf("data line 1 = %q, want %q", lines[1], "1,test1,100")
	}
}

func TestFormatCSV_WithoutHeader(t *testing.T) {
	headers := []string{"ID", "Name", "CPU"}
	rows := [][]string{
		{"1", "test1", "100"},
		{"2", "test2", "200"},
	}

	result := FormatCSV(headers, rows, false)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 2 {
		t.Errorf("number of lines = %d, want 2 (no header)", len(lines))
	}

	// Should not contain header
	if strings.Contains(result, "ID,Name,CPU") {
		t.Errorf("result should not contain header when includeHeader=false")
	}
}

func TestFormatCSV_Escaping(t *testing.T) {
	headers := []string{"Name", "Description"}
	rows := [][]string{
		{"test,with,commas", "normal"},
		{"test\"with\"quotes", "normal"},
		{"test\nwith\nnewlines", "normal"},
	}

	result := FormatCSV(headers, rows, true)

	// Verify commas are escaped
	if !strings.Contains(result, "\"test,with,commas\"") {
		t.Errorf("result should escape commas with quotes")
	}

	// Verify quotes are escaped
	if !strings.Contains(result, "\"test\"\"with\"\"quotes\"") {
		t.Errorf("result should escape quotes by doubling them")
	}

	// Verify newlines are escaped
	if !strings.Contains(result, "\"test\nwith\nnewlines\"") {
		t.Errorf("result should escape newlines with quotes")
	}
}

func TestFilterFields(t *testing.T) {
	data := []map[string]interface{}{
		{"id": "1", "name": "test1", "cpu": 100, "memory": 256},
		{"id": "2", "name": "test2", "cpu": 200, "memory": 512},
	}

	fields := []string{"id", "name"}
	filtered := FilterFields(data, fields)

	if len(filtered) != 2 {
		t.Errorf("filtered length = %d, want 2", len(filtered))
	}

	// Each item should only have id and name
	for i, item := range filtered {
		if len(item) != 2 {
			t.Errorf("filtered[%d] has %d fields, want 2", i, len(item))
		}
		if _, ok := item["id"]; !ok {
			t.Errorf("filtered[%d] missing id field", i)
		}
		if _, ok := item["name"]; !ok {
			t.Errorf("filtered[%d] missing name field", i)
		}
		if _, ok := item["cpu"]; ok {
			t.Errorf("filtered[%d] should not have cpu field", i)
		}
	}
}

func TestFilterFields_EmptyFields(t *testing.T) {
	data := []map[string]interface{}{
		{"id": "1", "name": "test1", "cpu": 100},
	}

	filtered := FilterFields(data, []string{})

	// Should return original data when no fields specified
	if len(filtered) != len(data) {
		t.Errorf("filtered length = %d, want %d", len(filtered), len(data))
	}
	if len(filtered[0]) != len(data[0]) {
		t.Errorf("filtered[0] fields = %d, want %d", len(filtered[0]), len(data[0]))
	}
}

func TestFilterTableFields(t *testing.T) {
	headers := []string{"ID", "Name", "CPU", "Memory"}
	rows := [][]string{
		{"1", "test1", "100", "256"},
		{"2", "test2", "200", "512"},
	}

	fields := []string{"name", "cpu"}
	filteredHeaders, filteredRows := FilterTableFields(headers, rows, fields)

	// Should only have 2 headers
	if len(filteredHeaders) != 2 {
		t.Errorf("filtered headers length = %d, want 2", len(filteredHeaders))
	}

	// Headers should be Name and CPU
	if filteredHeaders[0] != "Name" {
		t.Errorf("filteredHeaders[0] = %q, want Name", filteredHeaders[0])
	}
	if filteredHeaders[1] != "CPU" {
		t.Errorf("filteredHeaders[1] = %q, want CPU", filteredHeaders[1])
	}

	// Each row should have 2 fields
	for i, row := range filteredRows {
		if len(row) != 2 {
			t.Errorf("filteredRows[%d] length = %d, want 2", i, len(row))
		}
	}

	// First row should have test1 and 100
	if filteredRows[0][0] != "test1" {
		t.Errorf("filteredRows[0][0] = %q, want test1", filteredRows[0][0])
	}
	if filteredRows[0][1] != "100" {
		t.Errorf("filteredRows[0][1] = %q, want 100", filteredRows[0][1])
	}
}

func TestFilterTableFields_EmptyFields(t *testing.T) {
	headers := []string{"ID", "Name", "CPU"}
	rows := [][]string{
		{"1", "test1", "100"},
	}

	filteredHeaders, filteredRows := FilterTableFields(headers, rows, []string{})

	// Should return original data when no fields specified
	if len(filteredHeaders) != len(headers) {
		t.Errorf("filtered headers length = %d, want %d", len(filteredHeaders), len(headers))
	}
	if len(filteredRows) != len(rows) {
		t.Errorf("filtered rows length = %d, want %d", len(filteredRows), len(rows))
	}
}

func TestFilterTableFields_CaseInsensitive(t *testing.T) {
	headers := []string{"ID", "Name", "CPU"}
	rows := [][]string{
		{"1", "test1", "100"},
	}

	// Use lowercase field names
	fields := []string{"id", "cpu"}
	filteredHeaders, filteredRows := FilterTableFields(headers, rows, fields)

	// Should still match despite case difference
	if len(filteredHeaders) != 2 {
		t.Errorf("filtered headers length = %d, want 2 (case insensitive matching)", len(filteredHeaders))
	}

	// Should have ID and CPU (original case preserved)
	if filteredHeaders[0] != "ID" {
		t.Errorf("filteredHeaders[0] = %q, want ID", filteredHeaders[0])
	}
	if filteredHeaders[1] != "CPU" {
		t.Errorf("filteredHeaders[1] = %q, want CPU", filteredHeaders[1])
	}

	// Check rows are filtered correctly
	if len(filteredRows) != 1 {
		t.Errorf("filtered rows length = %d, want 1", len(filteredRows))
	}
	if len(filteredRows[0]) != 2 {
		t.Errorf("filteredRows[0] length = %d, want 2", len(filteredRows[0]))
	}
}

func TestEscapeCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"simple", "simple", "simple string"},
		{"with,comma", "\"with,comma\"", "string with comma"},
		{"with\"quote", "\"with\"\"quote\"", "string with quote"},
		{"with\nnewline", "\"with\nnewline\"", "string with newline"},
		{"with,comma\"and\"quote", "\"with,comma\"\"and\"\"quote\"", "string with comma and quote"},
	}

	for _, tt := range tests {
		result := escapeCSV(tt.input)
		if result != tt.expected {
			t.Errorf("%s: escapeCSV(%q) = %q, want %q", tt.desc, tt.input, result, tt.expected)
		}
	}
}
