package tail

import (
	"encoding/json"
	"testing"
)

func TestTailEvent_UnmarshalJSON(t *testing.T) {
	raw := `{
		"outcome": "ok",
		"scriptName": "my-worker",
		"eventTimestamp": 1709123456789,
		"event": {
			"request": {
				"url": "https://example.com/api/users",
				"method": "GET",
				"headers": {"user-agent": "Mozilla/5.0"}
			},
			"response": {
				"status": 200
			}
		},
		"logs": [
			{"level": "log", "message": ["hello world"], "timestamp": 1709123456789}
		],
		"exceptions": [],
		"diagnosticsChannelEvents": []
	}`

	var event TailEvent
	err := json.Unmarshal([]byte(raw), &event)
	if err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	if event.Outcome != "ok" {
		t.Errorf("Outcome = %q, want %q", event.Outcome, "ok")
	}
	if event.ScriptName != "my-worker" {
		t.Errorf("ScriptName = %q, want %q", event.ScriptName, "my-worker")
	}
	if event.EventTimestamp != 1709123456789 {
		t.Errorf("EventTimestamp = %d, want 1709123456789", event.EventTimestamp)
	}
	if event.Event.Request.URL != "https://example.com/api/users" {
		t.Errorf("Request.URL = %q", event.Event.Request.URL)
	}
	if event.Event.Request.Method != "GET" {
		t.Errorf("Request.Method = %q, want GET", event.Event.Request.Method)
	}
	if event.Event.Response.Status != 200 {
		t.Errorf("Response.Status = %d, want 200", event.Event.Response.Status)
	}
	if len(event.Logs) != 1 {
		t.Fatalf("Logs length = %d, want 1", len(event.Logs))
	}
	if event.Logs[0].Level != "log" {
		t.Errorf("Logs[0].Level = %q, want log", event.Logs[0].Level)
	}
	if len(event.Logs[0].Message) != 1 || event.Logs[0].Message[0] != "hello world" {
		t.Errorf("Logs[0].Message = %v, want [hello world]", event.Logs[0].Message)
	}
}

func TestTailEvent_WithExceptions(t *testing.T) {
	raw := `{
		"outcome": "exception",
		"scriptName": "broken-worker",
		"eventTimestamp": 1709123456789,
		"event": {
			"request": {"url": "https://example.com/fail", "method": "POST"},
			"response": {"status": 500}
		},
		"logs": [],
		"exceptions": [
			{"name": "TypeError", "message": "Cannot read property 'x' of undefined", "timestamp": 1709123456789}
		],
		"diagnosticsChannelEvents": []
	}`

	var event TailEvent
	err := json.Unmarshal([]byte(raw), &event)
	if err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	if event.Outcome != "exception" {
		t.Errorf("Outcome = %q, want exception", event.Outcome)
	}
	if len(event.Exceptions) != 1 {
		t.Fatalf("Exceptions length = %d, want 1", len(event.Exceptions))
	}
	if event.Exceptions[0].Name != "TypeError" {
		t.Errorf("Exceptions[0].Name = %q, want TypeError", event.Exceptions[0].Name)
	}
	if event.Exceptions[0].Message != "Cannot read property 'x' of undefined" {
		t.Errorf("Exceptions[0].Message = %q", event.Exceptions[0].Message)
	}
}

func TestTailEvent_Timestamp(t *testing.T) {
	event := TailEvent{EventTimestamp: 1709123456789}
	ts := event.Time()

	if ts.Year() != 2024 {
		t.Errorf("Year = %d, want 2024", ts.Year())
	}
	if ts.IsZero() {
		t.Error("Time() returned zero time")
	}
}

func TestTailEvent_LogMessages(t *testing.T) {
	raw := `{
		"outcome": "ok",
		"scriptName": "test",
		"eventTimestamp": 1709123456789,
		"event": {"request": {"url": "https://example.com", "method": "GET"}, "response": {"status": 200}},
		"logs": [
			{"level": "log", "message": ["msg1", "msg2"], "timestamp": 1709123456789},
			{"level": "error", "message": ["err msg"], "timestamp": 1709123456790},
			{"level": "warn", "message": ["warn msg"], "timestamp": 1709123456791}
		],
		"exceptions": [],
		"diagnosticsChannelEvents": []
	}`

	var event TailEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	if len(event.Logs) != 3 {
		t.Fatalf("Logs length = %d, want 3", len(event.Logs))
	}
	if event.Logs[1].Level != "error" {
		t.Errorf("Logs[1].Level = %q, want error", event.Logs[1].Level)
	}
}
