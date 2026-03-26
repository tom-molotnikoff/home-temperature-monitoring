package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClientFromConfig(cmd interface{ Flags() interface{ GetString(string) (string, error) } }) (*Client, error) {
	return nil, fmt.Errorf("use NewClient instead")
}

func NewClient(serverURL, apiKey string, insecure bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested via --insecure flag
	}
	return &Client{
		BaseURL: serverURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	u := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	return req, nil
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not connect to %s\n", c.BaseURL)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		if len(data) > 0 {
			fmt.Fprintln(os.Stderr, string(data))
		}
		os.Exit(1)
	}

	return data, nil
}

func (c *Client) Get(path string, query url.Values) ([]byte, error) {
	if query != nil {
		path = path + "?" + query.Encode()
	}
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	req, err := c.newRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) Patch(path string, body interface{}) ([]byte, error) {
	req, err := c.newRequest("PATCH", path, body)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) Delete(path string) ([]byte, error) {
	req, err := c.newRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) Put(path string, body interface{}) ([]byte, error) {
	req, err := c.newRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) Head(path string) (int, error) {
	req, err := c.newRequest("HEAD", path, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not connect to %s\n", c.BaseURL)
		os.Exit(1)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func printJSON(data []byte) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return
	}
	fmt.Println(buf.String())
}
