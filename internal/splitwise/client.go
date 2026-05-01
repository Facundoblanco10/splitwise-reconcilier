package splitwise

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://secure.splitwise.com/api/v3.0"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *Client) get(path string, out any) error {
	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

func (c *Client) GetCurrentUser() (*User, error) {
	var result CurrentUserResponse
	if err := c.get("/get_current_user", &result); err != nil {
		return nil, err
	}
	return &result.User, nil
}

func (c *Client) GetGroups() ([]Group, error) {
	var result GroupsResponse
	if err := c.get("/get_groups", &result); err != nil {
		return nil, err
	}
	return result.Groups, nil
}
