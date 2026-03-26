package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check connectivity to a Sensor Hub server",
	RunE:  runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	serverURL, _, err := loadClientConfig(cmd)
	if err != nil {
		// Allow health check with just --server flag (no API key needed)
		serverFlag, _ := cmd.Flags().GetString("server")
		if serverFlag != "" {
			serverURL = serverFlag
		} else {
			return err
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	resp, err := client.Get(serverURL + "/api/health")
	latency := time.Since(start)

	if err != nil {
		result := map[string]interface{}{
			"status":  "error",
			"server":  serverURL,
			"message": err.Error(),
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
		return fmt.Errorf("could not connect to %s", serverURL)
	}
	defer resp.Body.Close()

	result := map[string]interface{}{
		"status":     "ok",
		"server":     serverURL,
		"latency_ms": latency.Milliseconds(),
		"http_code":  resp.StatusCode,
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}
