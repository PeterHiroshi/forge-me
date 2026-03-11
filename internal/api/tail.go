package api

import (
	"fmt"
	"time"
)

type TailFilter struct {
	Status       []string          `json:"status,omitempty"`
	Method       []string          `json:"method,omitempty"`
	ClientIP     []string          `json:"client_ip,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	SamplingRate float64           `json:"sampling_rate,omitempty"`
}

type tailRequestBody struct {
	Filters []TailFilter `json:"filters"`
}

func (f TailFilter) toRequestBody() tailRequestBody {
	return tailRequestBody{
		Filters: []TailFilter{f},
	}
}

type TailSession struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *Client) CreateTail(accountID, scriptName string, filters TailFilter) (*TailSession, error) {
	path := fmt.Sprintf("/accounts/%s/workers/scripts/%s/tails", accountID, scriptName)

	payload := filters.toRequestBody()

	var response struct {
		Result TailSession `json:"result"`
	}

	if err := c.doRequestWithBody("POST", path, payload, &response); err != nil {
		return nil, err
	}

	return &response.Result, nil
}

func (c *Client) DeleteTail(accountID, scriptName, tailID string) error {
	path := fmt.Sprintf("/accounts/%s/workers/scripts/%s/tails/%s", accountID, scriptName, tailID)
	return c.doRequest("DELETE", path, nil)
}
