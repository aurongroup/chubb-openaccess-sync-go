package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const apiVersion = "1.0"

var (
	defaultConnectTimeout = 10 * time.Second
	defaultRequestTimeout = 30 * time.Second
)

// ClientError represents an error from the OpenAccess API.
type ClientError struct {
	Message    string
	Method     string
	URI        string
	StatusCode int
}

func (e *ClientError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("%s %s: %s (HTTP %d)", e.Method, e.URI, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("%s %s: %s", e.Method, e.URI, e.Message)
}

// Client wraps the OpenAccess REST API.
type Client struct {
	baseURL      string
	appID        string
	user         string
	password     string
	directory    string
	pageSize     int
	sessionToken string
	http         *http.Client
}

// NewClient creates a Client, pings, and authenticates.
func NewClient(cfg AppConfig) (*Client, error) {
	transport := &http.Transport{
		TLSHandshakeTimeout:   defaultConnectTimeout,
		ResponseHeaderTimeout: defaultRequestTimeout,
	}

	if cfg.Insecure {
		log.Println("SSL certificate validation disabled")
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	c := &Client{
		baseURL:   cfg.Endpoint,
		appID:     cfg.Application,
		user:      cfg.User,
		password:  cfg.Password,
		directory: cfg.Directory,
		pageSize:  cfg.PageSize,
		http: &http.Client{
			Timeout:   defaultRequestTimeout,
			Transport: transport,
		},
	}

	if err := c.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	if err := c.Authenticate(); err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	return c, nil
}

func (c *Client) do(method, uri string, body any, extraHeaders map[string]string) (*http.Response, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, &ClientError{Message: "marshal request body: " + err.Error(), Method: method, URI: uri}
		}

		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, uri, bodyReader)
	if err != nil {
		return nil, nil, &ClientError{Message: err.Error(), Method: method, URI: uri}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, &ClientError{Message: "network error: " + err.Error(), Method: method, URI: uri}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, &ClientError{Message: "read response body: " + err.Error(), Method: method, URI: uri, StatusCode: resp.StatusCode}
	}

	return resp, raw, nil
}

func (c *Client) authHeaders() map[string]string {
	return map[string]string{
		"Session-Token":  c.sessionToken,
		"Application-Id": c.appID,
	}
}

// Ping verifies the API is reachable.
func (c *Client) Ping() error {
	uri := c.baseURL + "/version?version=" + apiVersion

	resp, _, err := c.do("GET", uri, nil, map[string]string{"Application-Id": c.appID})
	if err != nil {
		return err
	}

	log.Printf("ping status=%d", resp.StatusCode)
	return nil
}

// Authenticate logs in and stores the session token.
func (c *Client) Authenticate() error {
	log.Printf("Authenticating user=%s directory=%s", c.user, c.directory)

	uri := c.baseURL + "/authentication?version=" + apiVersion

	body := map[string]string{
		"user_name":    c.user,
		"password":     c.password,
		"directory_id": c.directory,
	}

	resp, raw, err := c.do("POST", uri, body, map[string]string{"Application-Id": c.appID})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return &ClientError{Message: "authentication failed", Method: "POST", URI: uri, StatusCode: resp.StatusCode}
	}

	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return &ClientError{Message: "parse auth response: " + err.Error(), Method: "POST", URI: uri}
	}

	token, _ := result["session_token"].(string)
	if token == "" {
		return &ClientError{Message: "no session_token in auth response", Method: "POST", URI: uri}
	}
	c.sessionToken = token

	return nil
}

// Close ends the session.
func (c *Client) Close() error {
	if c.sessionToken == "" {
		return nil
	}

	uri := c.baseURL + "/authentication?version=" + apiVersion
	_, _, err := c.do("DELETE", uri, nil, c.authHeaders())
	c.sessionToken = ""

	return err
}

// GetInstances fetches all pages of the given type from the API.
// Returns the property_value_map contents for each item.
func (c *Client) GetInstances(typeName, filter string) ([]map[string]any, error) {
	var all []map[string]any

	for page := 1; ; page++ {
		items, totalPages, err := c.getInstancesPage(typeName, filter, page)
		if err != nil {
			return nil, err
		}

		all = append(all, items...)
		if page >= totalPages {
			break
		}
	}
	return all, nil
}

func (c *Client) getInstancesPage(typeName, filter string, page int) ([]map[string]any, int, error) {
	params := url.Values{}
	params.Set("version", apiVersion)
	params.Set("type_name", typeName)
	params.Set("page_size", strconv.Itoa(c.pageSize))
	params.Set("page_number", strconv.Itoa(page))
	if filter != "" {
		params.Set("filter", filter)
	}

	uri := c.baseURL + "/instances?" + params.Encode()
	resp, raw, err := c.do("GET", uri, nil, c.authHeaders())
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, 0, &ClientError{Message: "unexpected status", Method: "GET", URI: uri, StatusCode: resp.StatusCode}
	}

	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, 0, &ClientError{Message: "parse response: " + err.Error(), Method: "GET", URI: uri}
	}

	if _, ok := result["page_number"]; !ok {
		return nil, 0, &ClientError{Message: "missing page_number in response", Method: "GET", URI: uri}
	}

	tp, ok := result["total_pages"]
	if !ok {
		return nil, 0, &ClientError{Message: "missing total_pages in response", Method: "GET", URI: uri}
	}
	totalPages := int(tp.(float64)) // FIXME

	var items []map[string]any
	if itemList, ok := result["item_list"].([]any); ok {
		for _, item := range itemList {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}

			props, ok := m["property_value_map"].(map[string]any)
			if !ok {
				continue
			}

			items = append(items, props)
		}
	}

	return items, totalPages, nil
}

// CreateInstance is a stub — not yet implemented.
func (c *Client) CreateInstance(_ string, _ map[string]any) (int, error) {
	return 0, nil
}

// UpdateInstance is a stub — not yet implemented.
func (c *Client) UpdateInstance(_ string, _ map[string]any) error {
	return nil
}

// DeleteInstance deletes an instance by sending its representation.
func (c *Client) DeleteInstance(typeName string, body map[string]any) error {
	uri := c.baseURL + "/instances?version=" + apiVersion

	resp, _, err := c.do("DELETE", uri, body, c.authHeaders())
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return &ClientError{Message: "delete failed", Method: "DELETE", URI: uri, StatusCode: resp.StatusCode}
	}

	return nil
}
