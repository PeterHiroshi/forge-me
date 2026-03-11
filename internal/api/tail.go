package api

import (
	"fmt"
	"time"
)

// TailSession represents an active tail session
type TailSession struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// TailFilter holds filters for a tail session
type TailFilter struct {
	Status       []string
	Method       []string
	SamplingRate float64
	ClientIP     []string
	Headers      map[string]string
}

type tailRequestFilter struct {
	Status       []string          `json:"status,omitempty"`
	Method       []string          `json:"method,omitempty"`
	SamplingRate float64           `json:"sampling_rate,omitempty"`
	ClientIP     []string          `json:"client_ip,omitempty"`
	Header       *tailHeaderFilter `json:"header,omitempty"`
}

type tailHeaderFilter struct {
	Key   string `json:"key"`
	Query string `json:"query"`
}

type tailRequestBody struct {
	Filters []tailRequestFilter `json:"filters"`
}

func (f TailFilter) toRequestBody() tailRequestBody {
	rf := tailRequestFilter{
		Status:       f.Status,
		Method:       f.Method,
		SamplingRate: f.SamplingRate,
		ClientIP:     f.ClientIP,
	}
	for k, v := range f.Headers {
		rf.Header = &tailHeaderFilter{Key: k, Query: v}
		break
	}
	return tailRequestBody{Filters: []tailRequestFilter{rf}}
}

type tailResponse struct {
	Success bool        `json:"success"`
	Result  TailSession `json:"result"`
}

// CreateTail creates a new tail session for a worker script
func (c *Client) CreateTail(accountID, scriptName string, filter TailFilter) (*TailSession, error) {
	path := fmt.Sprintf("/accounts/%s/workers/scripts/%s/tails", accountID, scriptName)
	body := filter.toRequestBody()
	var resp tailResponse
	if err := c.doRequestWithBody("POST", path, body, &resp); err != nil {
		return nil, fmt.Errorf("creating tail: %w", err)
	}
	return &resp.Result, nil
}

// DeleteTail deletes an existing tail session
func (c *Client) DeleteTail(accountID, scriptName, tailID string) error {
	path := fmt.Sprintf("/accounts/%s/workers/scripts/%s/tails/%s", accountID, scriptName, tailID)
	if err := c.doRequest("DELETE", path, nil); err != nil {
		return fmt.Errorf("deleting tail: %w", err)
	}
	return nil
}
