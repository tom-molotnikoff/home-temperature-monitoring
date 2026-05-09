package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	gen "example/sensorHub/gen"
)

type clientConfig struct {
	serverURL string
	apiKey    string
	insecure  bool
}

// newAPIClient builds a generated `gen.Client` configured from the CLI's
// existing config sources (config file + flags). The returned client injects
// the API key on every request via a RequestEditorFn and honours the
// `--insecure` flag for self-signed TLS endpoints.
func newAPIClient(cmd *cobra.Command) (*gen.Client, context.Context, error) {
	cfg, err := loadResolvedClientConfig(cmd)
	if err != nil {
		return nil, nil, err
	}
	return buildAPIClient(cfg.serverURL, cfg.apiKey, cfg.insecure)
}

// newAPIClientNoAuth is used by the `health` command, which intentionally
// works without credentials so users can verify connectivity before
// configuring an API key.
func newAPIClientNoAuth(serverURL string, insecure bool) (*gen.Client, context.Context, error) {
	return buildAPIClient(serverURL, "", insecure)
}

func newAPIClientWithResponses(cmd *cobra.Command) (*gen.ClientWithResponses, context.Context, clientConfig, error) {
	cfg, err := loadResolvedClientConfig(cmd)
	if err != nil {
		return nil, nil, clientConfig{}, err
	}

	client, ctx, err := buildAPIClientWithResponses(cfg.serverURL, cfg.apiKey, cfg.insecure)
	if err != nil {
		return nil, nil, clientConfig{}, err
	}
	return client, ctx, cfg, nil
}

func loadResolvedClientConfig(cmd *cobra.Command) (clientConfig, error) {
	serverURL, apiKey, insecure, err := loadClientConfig(cmd)
	if err != nil {
		return clientConfig{}, err
	}
	return clientConfig{
		serverURL: serverURL,
		apiKey:    apiKey,
		insecure:  insecure,
	}, nil
}

func buildHTTPClient(insecure bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested via --insecure flag
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

func buildClientOptions(httpClient *http.Client, apiKey string) []gen.ClientOption {
	opts := []gen.ClientOption{gen.WithHTTPClient(httpClient)}
	if apiKey != "" {
		opts = append(opts, gen.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("X-API-Key", apiKey)
			return nil
		}))
	}
	return opts
}

func buildAPIClient(serverURL, apiKey string, insecure bool) (*gen.Client, context.Context, error) {
	// Generated Client base URL must include the /api prefix because operation
	// paths in the spec are mounted under /api by gen.RegisterHandlers.
	baseURL := strings.TrimRight(serverURL, "/") + "/api"
	opts := buildClientOptions(buildHTTPClient(insecure), apiKey)

	client, err := gen.NewClient(baseURL, opts...)
	if err != nil {
		return nil, nil, err
	}
	return client, context.Background(), nil
}

func buildAPIClientWithResponses(serverURL, apiKey string, insecure bool) (*gen.ClientWithResponses, context.Context, error) {
	baseURL := strings.TrimRight(serverURL, "/") + "/api"
	opts := buildClientOptions(buildHTTPClient(insecure), apiKey)

	client, err := gen.NewClientWithResponses(baseURL, opts...)
	if err != nil {
		return nil, nil, err
	}
	return client, context.Background(), nil
}

func currentReadingsSnapshot(ctx context.Context, cfg clientConfig) ([]gen.Reading, error) {
	wsURL, err := currentReadingsWSURL(cfg.serverURL)
	if err != nil {
		return nil, err
	}

	dialer := websocket.Dialer{
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: cfg.insecure}, //nolint:gosec // user-requested via --insecure flag
		HandshakeTimeout: 30 * time.Second,
	}
	headers := http.Header{}
	if cfg.apiKey != "" {
		headers.Set("X-API-Key", cfg.apiKey)
	}

	conn, resp, err := dialer.DialContext(ctx, wsURL, headers)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		if resp != nil {
			data, readErr := io.ReadAll(resp.Body)
			if readErr == nil && len(data) > 0 {
				return nil, apiResponseError(resp.StatusCode, data)
			}
		}
		return nil, err
	}
	defer conn.Close()
	requireReadDeadline := time.Now().Add(30 * time.Second)
	if err := conn.SetReadDeadline(requireReadDeadline); err != nil {
		return nil, err
	}

	var readings []gen.Reading
	if err := conn.ReadJSON(&readings); err != nil {
		return nil, fmt.Errorf("failed to read current readings snapshot: %w", err)
	}
	return readings, nil
}

func currentReadingsWSURL(serverURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(serverURL, "/"))
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}
	switch parsed.Scheme {
	case "http":
		parsed.Scheme = "ws"
	case "https":
		parsed.Scheme = "wss"
	default:
		return "", fmt.Errorf("unsupported server URL scheme %q", parsed.Scheme)
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/api/readings/ws/current"
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func apiResponseError(status int, body []byte) error {
	if len(body) > 0 {
		var errorResponse gen.ErrorResponse
		if json.Unmarshal(body, &errorResponse) == nil && errorResponse.Message != "" {
			return errors.New(errorResponse.Message)
		}
		return fmt.Errorf("HTTP %d: %s", status, strings.TrimSpace(string(body)))
	}
	return fmt.Errorf("HTTP %d", status)
}

// consumeJSON reads the response body, prints it as pretty-printed JSON on
// success, and replicates the original CLI's error behaviour on non-2xx
// responses (write status + body to stderr and exit with code 1).
func consumeJSON(resp *http.Response, err error) error {
	if err != nil {
		// Match the previous behaviour: connection failures aborted the
		// process directly. The wrapped error already includes the URL.
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("failed to read response: %w", readErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		if len(data) > 0 {
			fmt.Fprintln(os.Stderr, string(data))
		}
		os.Exit(1)
	}

	printJSON(data)
	return nil
}

// consumeStatus is the HEAD-request equivalent of consumeJSON. It returns the
// HTTP status code without consuming the body or printing anything.
func consumeStatus(resp *http.Response, err error) (int, error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

// rawJSONReader returns an io.Reader over the JSON encoding of `body`. It's
// used by the *WithBody client methods when we need to send already-encoded
// JSON (e.g. dashboard/mqtt --file payloads pass raw bytes through).
func rawJSONReader(body any) (io.Reader, error) {
	switch b := body.(type) {
	case json.RawMessage:
		return bytes.NewReader(b), nil
	case []byte:
		return bytes.NewReader(b), nil
	}
	out, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return bytes.NewReader(out), nil
}

func printJSON(data []byte) {
	if len(data) == 0 {
		return
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return
	}
	fmt.Println(buf.String())
}

// decodeBody reads and JSON-decodes a successful response body into `out`.
// It does not check status code — callers should handle that separately.
func decodeBody(resp *http.Response, out any) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	return json.Unmarshal(data, out)
}
