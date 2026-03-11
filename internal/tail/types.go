package tail

import (
	"encoding/json"
	"time"
)

type TailLog struct {
	Level     string    `json:"level"`
	Message   []string  `json:"message"`
	Timestamp int64     `json:"timestamp"`
}

type TailException struct {
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	Timestamp int64     `json:"timestamp"`
}

type TailEvent struct {
	Outcome         string              `json:"outcome"`
	ScriptName      string              `json:"scriptName"`
	Exceptions      []TailException     `json:"exceptions"`
	Logs           []TailLog            `json:"logs"`
	EventTimestamp int64               `json:"eventTimestamp"`
	Event          TailEventDetails     `json:"event"`
}

type TailEventDetails struct {
	Request  TailRequest  `json:"request"`
	Response TailResponse `json:"response,omitempty"`
}

type TailRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

type TailResponse struct {
	Status int `json:"status"`
}
