package nodeclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
	"vpnbot-core-go/internal/models"
)

type Client struct {
	httpClient *fasthttp.Client
	token      string
	timeout    time.Duration
}

func New(token string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &fasthttp.Client{
			MaxIdleConnDuration: 90 * time.Second,
			MaxConnsPerHost:     10,
			ReadTimeout:         timeout,
			WriteTimeout:        timeout,
		},
		token:   token,
		timeout: timeout,
	}
}

func (c *Client) CreateUser(endpoint, inbound, userID string) (*models.XrayUserResponse, error) {
	reqBody := map[string]string{
		"inbound": inbound,
		"id":      userID,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	url := fmt.Sprintf("http://%s:8000/user", endpoint)
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.SetBody(bodyBytes)

	if err := c.httpClient.DoTimeout(req, resp, c.timeout); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("node returned status %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result models.XrayUserResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) DeleteUser(endpoint, userID string) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	url := fmt.Sprintf("http://%s:8000/user?id=%s", endpoint, userID)
	req.SetRequestURI(url)
	req.Header.SetMethod("DELETE")
	req.Header.Set("Authorization", "Bearer "+c.token)

	if err := c.httpClient.DoTimeout(req, resp, c.timeout); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK && resp.StatusCode() != fasthttp.StatusNotFound {
		return fmt.Errorf("node returned status %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

func (c *Client) GetUsers(endpoint string) (*models.XrayUsersResponse, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	url := fmt.Sprintf("http://%s:8000/user", endpoint)
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	if err := c.httpClient.DoTimeout(req, resp, c.timeout); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("node returned status %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result models.XrayUsersResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetInbounds(endpoint string) (*models.XrayInboundsResponse, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	url := fmt.Sprintf("http://%s:8000/inbound", endpoint)
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	if err := c.httpClient.DoTimeout(req, resp, c.timeout); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("node returned status %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result models.XrayInboundsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

