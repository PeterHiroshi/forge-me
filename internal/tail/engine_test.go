package tail

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TestEngine_ReceivesEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade error: %v", err)
			return
		}
		defer conn.Close()

		event := TailEvent{
			Outcome:        "ok",
			ScriptName:     "test-worker",
			EventTimestamp: time.Now().UnixMilli(),
			Event: TailEventDetail{
				Request:  TailRequest{URL: "https://example.com/test", Method: "GET"},
				Response: TailResponse{Status: 200},
			},
		}

		data, _ := json.Marshal(event)
		conn.WriteMessage(websocket.TextMessage, data)
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	var received []TailEvent
	var mu sync.Mutex

	engine := NewEngine(EngineConfig{
		WebSocketURL: wsURL,
		OnEvent: func(event TailEvent) {
			mu.Lock()
			received = append(received, event)
			mu.Unlock()
		},
	})

	go engine.Run()
	time.Sleep(200 * time.Millisecond)
	engine.Stop()

	mu.Lock()
	defer mu.Unlock()

	if len(received) == 0 {
		t.Fatal("Expected at least 1 event, got 0")
	}
	if received[0].Outcome != "ok" {
		t.Errorf("Outcome = %q, want %q", received[0].Outcome, "ok")
	}
}

func TestEngine_StopsGracefully(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for i := 0; i < 100; i++ {
			event := TailEvent{
				Outcome:        "ok",
				EventTimestamp: time.Now().UnixMilli(),
				Event: TailEventDetail{
					Request:  TailRequest{URL: "https://example.com", Method: "GET"},
					Response: TailResponse{Status: 200},
				},
			}
			data, _ := json.Marshal(event)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eventCount := 0
	var mu sync.Mutex

	engine := NewEngine(EngineConfig{
		WebSocketURL: wsURL,
		OnEvent: func(event TailEvent) {
			mu.Lock()
			eventCount++
			mu.Unlock()
		},
	})

	go engine.Run()
	time.Sleep(150 * time.Millisecond)
	engine.Stop()

	mu.Lock()
	count := eventCount
	mu.Unlock()

	if count == 0 {
		t.Error("Should have received some events before stop")
	}
	if count >= 100 {
		t.Error("Should have stopped before receiving all events")
	}
}

func TestEngine_MaxEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for i := 0; i < 20; i++ {
			event := TailEvent{
				Outcome:        "ok",
				EventTimestamp: time.Now().UnixMilli(),
				Event: TailEventDetail{
					Request:  TailRequest{URL: "https://example.com", Method: "GET"},
					Response: TailResponse{Status: 200},
				},
			}
			data, _ := json.Marshal(event)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eventCount := 0
	var mu sync.Mutex

	engine := NewEngine(EngineConfig{
		WebSocketURL: wsURL,
		MaxEvents:    5,
		OnEvent: func(event TailEvent) {
			mu.Lock()
			eventCount++
			mu.Unlock()
		},
	})

	engine.Run() // Blocks until max events

	mu.Lock()
	count := eventCount
	mu.Unlock()

	if count != 5 {
		t.Errorf("Event count = %d, want 5", count)
	}
}

func TestEngine_SearchFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		events := []TailEvent{
			{
				Outcome: "ok", EventTimestamp: time.Now().UnixMilli(),
				Event: TailEventDetail{
					Request: TailRequest{URL: "https://example.com", Method: "GET"}, Response: TailResponse{Status: 200},
				},
				Logs: []TailLog{{Level: "log", Message: []string{"matching keyword here"}}},
			},
			{
				Outcome: "ok", EventTimestamp: time.Now().UnixMilli(),
				Event: TailEventDetail{
					Request: TailRequest{URL: "https://example.com", Method: "GET"}, Response: TailResponse{Status: 200},
				},
				Logs: []TailLog{{Level: "log", Message: []string{"no match"}}},
			},
		}

		for _, e := range events {
			data, _ := json.Marshal(e)
			conn.WriteMessage(websocket.TextMessage, data)
			time.Sleep(10 * time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	var received []TailEvent
	var mu sync.Mutex

	engine := NewEngine(EngineConfig{
		WebSocketURL: wsURL,
		Search:       "keyword",
		OnEvent: func(event TailEvent) {
			mu.Lock()
			received = append(received, event)
			mu.Unlock()
		},
	})

	go engine.Run()
	time.Sleep(200 * time.Millisecond)
	engine.Stop()

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 1 {
		t.Errorf("Expected 1 matching event, got %d", len(received))
	}
}

func TestEngine_SinceFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		events := []TailEvent{
			{
				Outcome: "ok", EventTimestamp: time.Now().Add(-2 * time.Hour).UnixMilli(),
				Event: TailEventDetail{
					Request: TailRequest{URL: "https://example.com/old", Method: "GET"}, Response: TailResponse{Status: 200},
				},
			},
			{
				Outcome: "ok", EventTimestamp: time.Now().UnixMilli(),
				Event: TailEventDetail{
					Request: TailRequest{URL: "https://example.com/new", Method: "GET"}, Response: TailResponse{Status: 200},
				},
			},
		}

		for _, e := range events {
			data, _ := json.Marshal(e)
			conn.WriteMessage(websocket.TextMessage, data)
			time.Sleep(10 * time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	var received []TailEvent
	var mu sync.Mutex

	engine := NewEngine(EngineConfig{
		WebSocketURL: wsURL,
		Since:        1 * time.Hour,
		OnEvent: func(event TailEvent) {
			mu.Lock()
			received = append(received, event)
			mu.Unlock()
		},
	})

	go engine.Run()
	time.Sleep(200 * time.Millisecond)
	engine.Stop()

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 1 {
		t.Errorf("Expected 1 event (recent only), got %d", len(received))
	}
}
