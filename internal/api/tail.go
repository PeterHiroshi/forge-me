package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TailFilter struct {
	Status       []string          `json:"status,omitempty"`
	Method       []string          `json:"method,omitempty"`
	ClientIP     []string          `json:"client_ip,omitempty"`
	Header       map[string]string `json:"header,omitempty"`
	SamplingRate float64           `json:"sampling_rate,omitempty"`
}

type TailSession struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (c *Client) CreateTail(accountID, scriptName string, filters TailFilter) (*TailSession, error) {
	path := fmt.Sprintf("/accounts/%s/workers/scripts/%s/tails", accountID, scriptName)
	
	payload := map[string]interface{}{
		"filters": []TailFilter{filters},
	}

	req, err := http.NewRequest("POST", c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(jsonPayload))

	var response struct {
		Result TailSession `json:"result"`
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, err
	}

	return &response.Result, nil
}

func (c *Client) DeleteTail(accountID, scriptName, tailID string) error {
	path := fmt.Sprintf("/accounts/%s/workers/scripts/%s/tails/%s", accountID, scriptName, tailID)
	
	req, err := http.NewRequest("DELETE", c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	return c.doRequest(req, nil)
}
