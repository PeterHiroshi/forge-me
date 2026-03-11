package tail

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type TailEngine struct {
	url           string
	filters       TailFilter
	outputFormat  OutputFormat
	maxEvents     int
	eventsChan    chan TailEvent
	errorsChan    chan error
	stopChan      chan struct{}
	reconnectChan chan struct{}
	mu            sync.Mutex
	conn          *websocket.Conn
	eventCount    int
}

func NewTailEngine(url string, filters TailFilter, format OutputFormat, maxEvents int) *TailEngine {
	return &TailEngine{
		url:           url,
		filters:       filters,
		outputFormat:  format,
		maxEvents:     maxEvents,
		eventsChan:    make(chan TailEvent, 100),
		errorsChan:    make(chan error, 10),
		stopChan:      make(chan struct{}),
		reconnectChan: make(chan struct{}, 1),
	}
}

func (e *TailEngine) Start(ctx context.Context) {
	go e.connectionManager(ctx)
}

func (e *TailEngine) connectionManager(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := e.connect(ctx); err != nil {
				log.Printf("Tail connection error: %v", err)
				time.Sleep(e.backoffDuration())
				e.reconnectChan <- struct{}{}
			}
		}
	}
}

func (e *TailEngine) connect(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.DialContext(ctx, e.url, nil)
	if err != nil {
		return fmt.Errorf("websocket connect: %w", err)
	}
	e.conn = conn

	go e.receiveEvents(ctx)
	return nil
}

func (e *TailEngine) receiveEvents(ctx context.Context) {
	defer func() {
		if e.conn != nil {
			e.conn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := e.conn.ReadMessage()
			if err != nil {
				e.errorsChan <- fmt.Errorf("read message: %w", err)
				return
			}

			var event TailEvent
			if err := json.Unmarshal(message, &event); err != nil {
				e.errorsChan <- fmt.Errorf("parse event: %w", err)
				continue
			}

			e.mu.Lock()
			e.eventCount++
			shouldContinue := e.maxEvents == 0 || e.eventCount <= e.maxEvents
			e.mu.Unlock()

			if shouldContinue {
				e.eventsChan <- event
			} else {
				close(e.stopChan)
				return
			}
		}
	}
}

func (e *TailEngine) backoffDuration() time.Duration {
	return time.Second * time.Duration(5 * (1 + rand.Intn(3)))
}

func (e *TailEngine) GetEvents() <-chan TailEvent {
	return e.eventsChan
}

func (e *TailEngine) GetErrors() <-chan error {
	return e.errorsChan
}
