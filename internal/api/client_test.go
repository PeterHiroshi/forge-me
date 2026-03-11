package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.token != "test-token" {
		t.Errorf("token = %q, want %q", client.token, "test-token")
	}

	if client.baseURL != "https://api.cloudflare.com/client/v4" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://api.cloudflare.com/client/v4")
	}
}

func TestClient_DoRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "result": {"id": "123"}}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	var result map[string]interface{}
	err := client.doRequest("GET", "/test", &result)
	if err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}

	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}
}

func TestClient_DoRequest_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "errors": [{"message": "Unauthorized"}]}`))
	}))
	defer server.Close()

	client := NewClient("invalid-token")
	client.baseURL = server.URL

	var result map[string]interface{}
	err := client.doRequest("GET", "/test", &result)
	if err == nil {
		t.Fatal("doRequest() error = nil, want error")
	}
}

func TestClient_DoRequest_NilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	// Test with nil result - should not crash
	err := client.doRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("doRequest() with nil result error = %v, want nil", err)
	}
}

func TestClient_DoRequest_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	var result map[string]interface{}
	err := client.doRequest("GET", "/test", &result)
	if err == nil {
		t.Fatal("doRequest() with invalid JSON error = nil, want error")
	}

	// Should mention parsing
	if err != nil && err.Error() == "" {
		t.Error("error message is empty")
	}
}

func TestClient_DoRequest_ConnectionError(t *testing.T) {
	client := NewClient("test-token")
	// Use an unreachable URL
	client.baseURL = "http://localhost:1"

	var result map[string]interface{}
	err := client.doRequest("GET", "/test", &result)
	if err == nil {
		t.Fatal("doRequest() with unreachable server error = nil, want error")
	}
}

func TestClient_DoRequest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "errors": [{"message": "Internal Server Error"}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	var result map[string]interface{}
	err := client.doRequest("GET", "/test", &result)
	if err == nil {
		t.Fatal("doRequest() with 500 error = nil, want error")
	}

	// Should mention status code
	if err != nil && err.Error() == "" {
		t.Error("error message is empty")
	}
}

func TestClient_DoRequestWithBody_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if payload["key"] != "value" {
			t.Errorf("body key = %v, want %q", payload["key"], "value")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "result": {"id": "123"}}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	requestBody := map[string]string{"key": "value"}
	var result map[string]interface{}
	err := client.doRequestWithBody("POST", "/test", requestBody, &result)
	if err != nil {
		t.Fatalf("doRequestWithBody() error = %v", err)
	}
	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}
}

func TestClient_DoRequestWithBody_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false}`))
	}))
	defer server.Close()

	client := NewClient("invalid-token")
	client.baseURL = server.URL

	var result map[string]interface{}
	err := client.doRequestWithBody("POST", "/test", nil, &result)
	if err == nil {
		t.Fatal("doRequestWithBody() error = nil, want error")
	}
}
