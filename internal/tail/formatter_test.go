package tail

import (
	"strings"
	"testing"
)

func newTestEvent() TailEvent {
	return TailEvent{
		Outcome:        "ok",
		ScriptName:     "my-worker",
		EventTimestamp: 1709123456789,
		Event: TailEventDetail{
			Request:  TailRequest{URL: "https://example.com/api/users", Method: "GET"},
			Response: TailResponse{Status: 200},
		},
		Logs: []TailLog{
			{Level: "log", Message: []string{"hello world"}, Timestamp: 1709123456789},
		},
		Exceptions: nil,
	}
}

func TestFormatEvent_JSON(t *testing.T) {
	event := newTestEvent()
	f := NewFormatter("json", false)
	result := f.Format(event)

	if !strings.Contains(result, `"outcome":"ok"`) && !strings.Contains(result, `"outcome": "ok"`) {
		t.Errorf("JSON output should contain outcome, got: %s", result)
	}
	if !strings.Contains(result, `"scriptName"`) {
		t.Errorf("JSON output should contain scriptName, got: %s", result)
	}
}

func TestFormatEvent_Compact(t *testing.T) {
	event := newTestEvent()
	f := NewFormatter("compact", true) // no color
	result := f.Format(event)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 1 {
		t.Errorf("Compact format should be 1 line, got %d: %q", len(lines), result)
	}
	if !strings.Contains(result, "200") {
		t.Errorf("Should contain status 200, got: %s", result)
	}
	if !strings.Contains(result, "GET") {
		t.Errorf("Should contain method GET, got: %s", result)
	}
	if !strings.Contains(result, "example.com/api/users") {
		t.Errorf("Should contain URL, got: %s", result)
	}
}

func TestFormatEvent_Pretty(t *testing.T) {
	event := newTestEvent()
	f := NewFormatter("pretty", true) // no color
	result := f.Format(event)

	if !strings.Contains(result, "GET") {
		t.Errorf("Pretty output should contain method, got: %s", result)
	}
	if !strings.Contains(result, "200") {
		t.Errorf("Pretty output should contain status, got: %s", result)
	}
	if !strings.Contains(result, "https://example.com/api/users") {
		t.Errorf("Pretty output should contain URL, got: %s", result)
	}
	if !strings.Contains(result, "hello world") {
		t.Errorf("Pretty output should contain log message, got: %s", result)
	}
}

func TestFormatEvent_Pretty_WithExceptions(t *testing.T) {
	event := newTestEvent()
	event.Outcome = "exception"
	event.Exceptions = []TailException{
		{Name: "TypeError", Message: "undefined is not a function", Timestamp: 1709123456789},
	}
	f := NewFormatter("pretty", true)
	result := f.Format(event)

	if !strings.Contains(result, "TypeError") {
		t.Errorf("Should contain exception name, got: %s", result)
	}
	if !strings.Contains(result, "undefined is not a function") {
		t.Errorf("Should contain exception message, got: %s", result)
	}
}

func TestFormatEvent_Pretty_ErrorStatus(t *testing.T) {
	event := newTestEvent()
	event.Event.Response.Status = 500
	event.Outcome = "exception"
	f := NewFormatter("pretty", true)
	result := f.Format(event)

	if !strings.Contains(result, "500") {
		t.Errorf("Should contain status 500, got: %s", result)
	}
}

func TestFormatEvent_Pretty_WithColor(t *testing.T) {
	event := newTestEvent()
	f := NewFormatter("pretty", false) // color enabled
	result := f.Format(event)

	if !strings.Contains(result, "\033[") {
		t.Errorf("Colored output should contain ANSI codes, got: %s", result)
	}
}

func TestFormatEvent_Compact_WithErrorStatus(t *testing.T) {
	event := newTestEvent()
	event.Event.Response.Status = 500
	event.Outcome = "exception"
	f := NewFormatter("compact", true)
	result := f.Format(event)

	if !strings.Contains(result, "500") {
		t.Errorf("Should contain status 500, got: %s", result)
	}
}

func TestFormatEvent_Pretty_HideLogs(t *testing.T) {
	event := newTestEvent()
	f := NewFormatter("pretty", true)
	f.IncludeLogs = false
	result := f.Format(event)

	if strings.Contains(result, "hello world") {
		t.Errorf("Should not contain log message when IncludeLogs=false, got: %s", result)
	}
}

func TestFormatEvent_Pretty_HideExceptions(t *testing.T) {
	event := newTestEvent()
	event.Exceptions = []TailException{
		{Name: "Error", Message: "test error", Timestamp: 1709123456789},
	}
	f := NewFormatter("pretty", true)
	f.IncludeExceptions = false
	result := f.Format(event)

	if strings.Contains(result, "test error") {
		t.Errorf("Should not contain exception when IncludeExceptions=false, got: %s", result)
	}
}
