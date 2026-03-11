package tail

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type EngineConfig struct {
	WebSocketURL string
	MaxEvents    int
	Search       string
	Since        time.Duration
	OnEvent      func(TailEvent)
	OnError      func(error)
}

type Engine struct {
	config     EngineConfig
	conn       *websocket.Conn
	mu         sync.Mutex
	stopChan   chan struct{}
	stopped    bool
	eventCount int
	sinceTime  time.Time
}

func NewEngine(config EngineConfig) *Engine {
	var sinceTime time.Time
	if config.Since > 0 {
		sinceTime = time.Now().Add(-config.Since)
	}
	return &Engine{
		config:   config,
		stopChan: make(chan struct{}),
		sinceTime: sinceTime,
	}
}

// Run connects to the WebSocket and processes events. Blocks until Stop is called or MaxEvents is reached.
func (e *Engine) Run() {
	for {
		select {
		case <-e.stopChan:
			return
		default:
			if err := e.connectAndRead(); err != nil {
				e.mu.Lock()
				stopped := e.stopped
				e.mu.Unlock()
				if stopped {
					return
				}
				log.Printf("Tail connection error: %v", err)
				time.Sleep(2 * time.Second)
			}
		}
	}
}

// Stop gracefully shuts down the engine.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.stopped {
		e.stopped = true
		close(e.stopChan)
		if e.conn != nil {
			e.conn.Close()
		}
	}
}

func (e *Engine) connectAndRead() error {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(e.config.WebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("websocket connect: %w", err)
	}

	e.mu.Lock()
	e.conn = conn
	e.mu.Unlock()

	defer func() {
		conn.Close()
		e.mu.Lock()
		if e.conn == conn {
			e.conn = nil
		}
		e.mu.Unlock()
	}()

	for {
		select {
		case <-e.stopChan:
			return nil
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				e.mu.Lock()
				stopped := e.stopped
				e.mu.Unlock()
				if stopped {
					return nil
				}
				return fmt.Errorf("read message: %w", err)
			}

			var event TailEvent
			if err := json.Unmarshal(message, &event); err != nil {
				if e.config.OnError != nil {
					e.config.OnError(fmt.Errorf("parse event: %w", err))
				}
				continue
			}

			if !e.matchesFilters(event) {
				continue
			}

			e.mu.Lock()
			e.eventCount++
			count := e.eventCount
			maxEvents := e.config.MaxEvents
			e.mu.Unlock()

			if e.config.OnEvent != nil {
				e.config.OnEvent(event)
			}

			if maxEvents > 0 && count >= maxEvents {
				e.Stop()
				return nil
			}
		}
	}
}

func (e *Engine) matchesFilters(event TailEvent) bool {
	// Since filter
	if !e.sinceTime.IsZero() {
		eventTime := event.Time()
		if eventTime.Before(e.sinceTime) {
			return false
		}
	}

	// Search filter
	if e.config.Search != "" {
		found := false
		for _, l := range event.Logs {
			for _, msg := range l.Message {
				if strings.Contains(msg, e.config.Search) {
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
