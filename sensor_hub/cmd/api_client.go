package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	gen "example/sensorHub/gen"
)

// newAPIClient builds a generated `gen.Client` configured from the CLI's
// existing config sources (config file + flags). The returned client injects
// the API key on every request via a RequestEditorFn and honours the
// `--insecure` flag for self-signed TLS endpoints.
func newAPIClient(cmd *cobra.Command) (*gen.Client, context.Context, error) {
	serverURL, apiKey, insecure, err := loadClientConfig(cmd)
	if err != nil {
		return nil, nil, err
	}
	return buildAPIClient(serverURL, apiKey, insecure)
}

// newAPIClientNoAuth is used by the `health` command, which intentionally
// works without credentials so users can verify connectivity before
// configuring an API key.
func newAPIClientNoAuth(serverURL string, insecure bool) (*gen.Client, context.Context, error) {
	return buildAPIClient(serverURL, "", insecure)
}

func buildAPIClient(serverURL, apiKey string, insecure bool) (*gen.Client, context.Context, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested via --insecure flag
	}
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	// Generated Client base URL must include the /api prefix because operation
	// paths in the spec are mounted under /api by gen.RegisterHandlers.
	baseURL := strings.TrimRight(serverURL, "/") + "/api"

	opts := []gen.ClientOption{gen.WithHTTPClient(httpClient)}
	if apiKey != "" {
		opts = append(opts, gen.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("X-API-Key", apiKey)
			return nil
		}))
	}

	client, err := gen.NewClient(baseURL, opts...)
	if err != nil {
		return nil, nil, err
	}
	return client, context.Background(), nil
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
