package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"openaccess-sync/pkg/config"
	"strconv"
	"time"

	ansi "github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

const apiVersion = "1.0"

var (
	defaultConnectTimeout = 10 * time.Second
	defaultRequestTimeout = 30 * time.Second
)

type InstanceItem struct {
	PropertyMap map[string]any `json:"property_value_map"`
}

type InstanceResponse struct {
	PageNumber int            `json:"page_number"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
	TotalItems int            `json:"total_items"`
	Count      int            `json:"count"`
	ItemList   []InstanceItem `json:"item_list"`
	TypeName   string         `json:"type_name"`
}

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
func NewClient(cfg config.AppConfig) (*Client, error) {
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
	return c.getInstances(typeName, filter, false)
}

// GetInstancesWithProgress fetches all pages of the given type, displaying a
// progress bar as each page arrives.
// Returns the property_value_map contents for each item.
func (c *Client) GetInstancesWithProgress(typeName, filter string) ([]map[string]any, error) {
	return c.getInstances(typeName, filter, true)
}

// getInstances fetches all pages of the given type, and if required, displaying a
// progress bar as each page arrives.
func (c *Client) getInstances(typeName, filter string, progress bool) ([]map[string]any, error) {
	log.Printf("Fetching %s pages from OpenAccess API...", typeName)

	var bar *progressbar.ProgressBar
	var all []map[string]any

	for page := 1; ; page++ {
		items, totalPages, err := c.getInstancesPage(typeName, filter, page)
		if err != nil {
			return nil, err
		}

		if progress {
			if bar == nil {
				bar = progressbar.NewOptions(totalPages,
					progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
					progressbar.OptionEnableColorCodes(true),
					progressbar.OptionSetDescription(typeName),
					progressbar.OptionShowDescriptionAtLineEnd(),
					progressbar.OptionShowCount(),
					progressbar.OptionSetWidth(30),
					progressbar.OptionSetTheme(progressbar.Theme{
						Saucer:        "[green]=[reset]",
						SaucerHead:    "[green]>[reset]",
						SaucerPadding: " ",
						BarStart:      "[",
						BarEnd:        "]",
					}),
				)
			}
			_ = bar.Add(1)
		}

		all = append(all, items...)

		if page >= totalPages {
			break
		}
	}

	if progress {
		_ = bar.Finish()
	}

	fmt.Println()

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

	var result InstanceResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, 0, &ClientError{Message: "parse response: " + err.Error(), Method: "GET", URI: uri}
	}

	var items []map[string]any
	for _, item := range result.ItemList {
		items = append(items, item.PropertyMap)
	}

	return items, result.TotalPages, nil
}

// CreateInstance is a stub — not yet implemented.
/*

{"error":{"code":"openaccess.general.invalidrequestitem","message":"The property PERSONID did not exist or contained an invalid value."},"version":"1.0"}

{"property_value_map":{"ID":13426},"type_name":"Lnl_Cardholder","version":"1.0"}

{"property_value_map":{"BADGEKEY":13239},"type_name":"Lnl_Badge","version":"1.0"}

{"property_value_map":{"AccessLevelID":1,"BadgeKey":13239},"type_name":"Lnl_AccessLevelAssignment","version":"1.0"}
*/
func (c *Client) CreateInstance(typeName string, params map[string]any) (map[string]any, error) {
	uri := c.baseURL + "/instances?version=" + apiVersion

	body := make(map[string]any)
	body["type_name"] = typeName
	body["property_value_map"] = params

	resp, raw, err := c.do("POST", uri, body, c.authHeaders())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &ClientError{Message: "create failed", Method: "POST", URI: uri, StatusCode: resp.StatusCode}
	}

	var item InstanceItem
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, &ClientError{Message: "unmarshal response: " + err.Error(), Method: "POST", URI: uri, StatusCode: resp.StatusCode}
	}

	return item.PropertyMap, nil
}

// UpdateInstance is a stub — not yet implemented.
func (c *Client) UpdateInstance(typeName string, params map[string]any) error {
	uri := c.baseURL + "/instances?version=" + apiVersion

	body := make(map[string]any)
	body["type_name"] = typeName
	body["property_value_map"] = params

	resp, _, err := c.do("PUT", uri, body, c.authHeaders())
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return &ClientError{Message: "create failed", Method: "POST", URI: uri, StatusCode: resp.StatusCode}
	}

	return nil
}

// DeleteInstance deletes an instance by sending its representation.
func (c *Client) DeleteInstance(typeName string, params map[string]any) error {
	uri := c.baseURL + "/instances?version=" + apiVersion

	body := make(map[string]any)
	body["type_name"] = typeName
	body["property_value_map"] = params

	resp, _, err := c.do("DELETE", uri, body, c.authHeaders())
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return &ClientError{Message: "delete failed", Method: "DELETE", URI: uri, StatusCode: resp.StatusCode}
	}

	return nil
}

// GetCardholdersWithProgress fetches all pages of cardholder data, displaying a
// progress bar as each page arrives.
// Returns the property_value_map contents for each item.
func (c *Client) GetCardholdersWithProgress(detachedOnly bool, filter string) ([]map[string]any, error) {
	return c.getCardholders(detachedOnly, filter, true)
}

// getInstances fetches all pages of the given type, and if required, displaying a
// progress bar as each page arrives.
func (c *Client) getCardholders(detachedOnly bool, filter string, progress bool) ([]map[string]any, error) {
	log.Printf("Fetching Lnl_Cardholder pages from OpenAccess API...")

	var bar *progressbar.ProgressBar
	var all []map[string]any

	for page := 1; ; page++ {
		items, totalPages, err := c.getCardholderPage(detachedOnly, filter, page)
		if err != nil {
			return nil, err
		}

		if progress {
			if bar == nil {
				bar = progressbar.NewOptions(totalPages,
					progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
					progressbar.OptionEnableColorCodes(true),
					progressbar.OptionSetDescription("Lnl_Cardholder"),
					progressbar.OptionShowDescriptionAtLineEnd(),
					progressbar.OptionShowCount(),
					progressbar.OptionSetWidth(30),
					progressbar.OptionSetTheme(progressbar.Theme{
						Saucer:        "[green]=[reset]",
						SaucerHead:    "[green]>[reset]",
						SaucerPadding: " ",
						BarStart:      "[",
						BarEnd:        "]",
					}),
				)
			}
			_ = bar.Add(1)
		}

		all = append(all, items...)

		if page >= totalPages {
			break
		}
	}

	if progress {
		_ = bar.Finish()
	}

	fmt.Println()

	return all, nil
}

func (c *Client) getCardholderPage(detachedOnly bool, filter string, page int) ([]map[string]any, int, error) {
	params := url.Values{}
	params.Set("version", apiVersion)
	params.Set("page_size", strconv.Itoa(c.pageSize))
	params.Set("page_number", strconv.Itoa(page))

	if detachedOnly {
		params.Set("has_badges", "false")
	}

	if filter != "" {
		params.Set("filter", filter)
	}

	uri := c.baseURL + "/cardholders?" + params.Encode()
	resp, raw, err := c.do("GET", uri, nil, c.authHeaders())
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, 0, &ClientError{Message: "unexpected status", Method: "GET", URI: uri, StatusCode: resp.StatusCode}
	}

	var result InstanceResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, 0, &ClientError{Message: "parse response: " + err.Error(), Method: "GET", URI: uri}
	}

	var items []map[string]any
	for _, item := range result.ItemList {
		items = append(items, item.PropertyMap)
	}

	return items, result.TotalPages, nil
}
