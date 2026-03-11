package tail

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// EngineConfig configures the tail engine
type EngineConfig struct {
	WebSocketURL string
	OnEvent      func(TailEvent)
	OnError      func(error)
	MaxEvents    int
	Search       string
	Since        time.Duration
}

// Engine manages the WebSocket connection and event processing
type Engine struct {
	config     EngineConfig
	stopCh     chan struct{}
	stoppedCh  chan struct{}
	stopOnce   sync.Once
	eventCount int
}

// NewEngine creates a new tail engine
func NewEngine(config EngineConfig) *Engine {
	return &Engine{
		config:    config,
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}
}

// Run connects to the WebSocket and processes events. Blocks until stopped or max events reached.
func (e *Engine) Run() {
	defer close(e.stoppedCh)

	conn, _, err := websocket.DefaultDialer.Dial(e.config.WebSocketURL, nil)
	if err != nil {
		if e.config.OnError != nil {
			e.config.OnError(err)
		}
		return
	}
	defer conn.Close()

	eventCh := make(chan TailEvent)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					return
				}
				select {
				case errCh <- err:
				default:
				}
				return
			}

			var event TailEvent
			if err := json.Unmarshal(message, &event); err != nil {
				continue
			}

			select {
			case eventCh <- event:
			case <-e.stopCh:
				return
			}
		}
	}()

	for {
		select {
		case <-e.stopCh:
			conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return

		case event, ok := <-eventCh:
			if !ok {
				return
			}

			if !e.matchesFilters(event) {
				continue
			}

			if e.config.OnEvent != nil {
				e.config.OnEvent(event)
			}

			e.eventCount++
			if e.config.MaxEvents > 0 && e.eventCount >= e.config.MaxEvents {
				return
			}

		case err := <-errCh:
			if e.config.OnError != nil {
				e.config.OnError(err)
			}
			return
		}
	}
}

// Stop signals the engine to shut down gracefully
func (e *Engine) Stop() {
	e.stopOnce.Do(func() {
		close(e.stopCh)
	})
	<-e.stoppedCh
}

// matchesFilters checks if an event passes client-side filters
func (e *Engine) matchesFilters(event TailEvent) bool {
	if e.config.Since > 0 {
		cutoff := time.Now().Add(-e.config.Since)
		if event.Time().Before(cutoff) {
			return false
		}
	}

	if e.config.Search != "" {
		found := false
		for _, log := range event.Logs {
			for _, msg := range log.Message {
				if strings.Contains(strings.ToLower(msg), strings.ToLower(e.config.Search)) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
