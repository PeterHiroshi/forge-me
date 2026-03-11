package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTail_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		expectedPath := "/accounts/acc-123/workers/scripts/my-worker/tails"
		if r.URL.Path != expectedPath {
			t.Errorf("Path = %s, want %s", r.URL.Path, expectedPath)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		resp := map[string]interface{}{
			"success": true,
			"result": map[string]interface{}{
				"id":         "tail-abc",
				"url":        "wss://tail.developers.workers.dev/abc",
				"expires_at": "2026-03-11T08:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	filter := TailFilter{SamplingRate: 1.0}
	session, err := client.CreateTail("acc-123", "my-worker", filter)
	if err != nil {
		t.Fatalf("CreateTail() error = %v", err)
	}
	if session.ID != "tail-abc" {
		t.Errorf("ID = %q, want %q", session.ID, "tail-abc")
	}
	if session.URL != "wss://tail.developers.workers.dev/abc" {
		t.Errorf("URL = %q, want %q", session.URL, "wss://tail.developers.workers.dev/abc")
	}
}

func TestCreateTail_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "errors": [{"message": "script not found"}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	_, err := client.CreateTail("acc-123", "missing-worker", TailFilter{})
	if err == nil {
		t.Fatal("CreateTail() error = nil, want error")
	}
}

func TestDeleteTail_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %s, want DELETE", r.Method)
		}
		expectedPath := "/accounts/acc-123/workers/scripts/my-worker/tails/tail-abc"
		if r.URL.Path != expectedPath {
			t.Errorf("Path = %s, want %s", r.URL.Path, expectedPath)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	err := client.DeleteTail("acc-123", "my-worker", "tail-abc")
	if err != nil {
		t.Fatalf("DeleteTail() error = %v", err)
	}
}

func TestDeleteTail_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	err := client.DeleteTail("acc-123", "my-worker", "tail-not-found")
	if err == nil {
		t.Fatal("DeleteTail() error = nil, want error")
	}
}

func TestTailFilter_ToRequest(t *testing.T) {
	filter := TailFilter{
		Status:       []string{"ok", "error"},
		Method:       []string{"GET", "POST"},
		SamplingRate: 0.5,
		ClientIP:     []string{"1.2.3.4"},
		Headers:      map[string]string{"X-Custom": "value"},
	}
	req := filter.toRequestBody()

	if len(req.Filters) != 1 {
		t.Fatalf("Filters length = %d, want 1", len(req.Filters))
	}
	f := req.Filters[0]
	if len(f.Status) != 2 || f.Status[0] != "ok" {
		t.Errorf("Status = %v, want [ok error]", f.Status)
	}
	if len(f.Method) != 2 || f.Method[0] != "GET" {
		t.Errorf("Method = %v, want [GET POST]", f.Method)
	}
	if f.SamplingRate != 0.5 {
		t.Errorf("SamplingRate = %f, want 0.5", f.SamplingRate)
	}
}

func TestTailFilter_Empty(t *testing.T) {
	filter := TailFilter{}
	req := filter.toRequestBody()
	if len(req.Filters) != 1 {
		t.Fatalf("Filters length = %d, want 1", len(req.Filters))
	}
	f := req.Filters[0]
	if len(f.Status) != 0 {
		t.Errorf("Status = %v, want empty", f.Status)
	}
}
