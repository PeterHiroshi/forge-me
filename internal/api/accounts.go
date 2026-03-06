package api

import "fmt"

// Account represents a Cloudflare account
type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

type accountsResponse struct {
	Success bool      `json:"success"`
	Result  []Account `json:"result"`
}

// ListAccounts lists all accounts accessible with the current API token
func (c *Client) ListAccounts() ([]Account, error) {
	path := "/accounts"

	var resp accountsResponse
	if err := c.doRequest("GET", path, &resp); err != nil {
		return nil, fmt.Errorf("listing accounts: %w", err)
	}

	return resp.Result, nil
}