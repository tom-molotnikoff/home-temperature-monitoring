package cmd

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage alert rules",
}

func init() {
	alertsCmd.AddCommand(alertsListCmd)
	alertsCmd.AddCommand(alertsGetCmd)
	alertsCmd.AddCommand(alertsCreateCmd)
	alertsCmd.AddCommand(alertsUpdateCmd)
	alertsCmd.AddCommand(alertsDeleteCmd)
	alertsCmd.AddCommand(alertsHistoryCmd)
	rootCmd.AddCommand(alertsCmd)
}

var alertsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all alert rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/alerts", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var alertsGetCmd = &cobra.Command{
	Use:   "get [sensorId]",
	Short: "Get alert rules for a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/alerts/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var alertsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new alert rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		sensorId, _ := cmd.Flags().GetInt("sensor-id")
		alertType, _ := cmd.Flags().GetString("type")
		threshold, _ := cmd.Flags().GetFloat64("threshold")

		if sensorId == 0 || alertType == "" {
			return fmt.Errorf("--sensor-id and --type are required")
		}

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]interface{}{
			"sensor_id":  sensorId,
			"alert_type": alertType,
			"threshold":  threshold,
		}
		data, err := client.Post("/api/alerts", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	alertsCreateCmd.Flags().Int("sensor-id", 0, "Sensor ID")
	alertsCreateCmd.Flags().String("type", "", "Alert type")
	alertsCreateCmd.Flags().Float64("threshold", 0, "Threshold value")
}

var alertsDeleteCmd = &cobra.Command{
	Use:   "delete [sensorId]",
	Short: "Delete alert rules for a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Delete("/api/alerts/" + args[0])
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var alertsHistoryCmd = &cobra.Command{
	Use:   "history [sensorId]",
	Short: "Get alert history for a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sensorId := args[0]
		if _, err := strconv.Atoi(sensorId); err != nil {
			return fmt.Errorf("sensor ID must be a number")
		}

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		query := url.Values{}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			query.Set("limit", strconv.Itoa(limit))
		}
		data, err := client.Get("/api/alerts/"+sensorId+"/history", query)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	alertsHistoryCmd.Flags().Int("limit", 0, "Maximum number of history records (1-100, default 50)")
}

var alertsUpdateCmd = &cobra.Command{
	Use:   "update [sensorId]",
	Short: "Update an alert rule for a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alertType, _ := cmd.Flags().GetString("alert-type")
		highThreshold, _ := cmd.Flags().GetFloat64("high-threshold")
		lowThreshold, _ := cmd.Flags().GetFloat64("low-threshold")
		enabled, _ := cmd.Flags().GetBool("enabled")
		rateLimitHours, _ := cmd.Flags().GetInt("rate-limit-hours")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]interface{}{
			"AlertType":      alertType,
			"HighThreshold":  highThreshold,
			"LowThreshold":   lowThreshold,
			"Enabled":        enabled,
			"RateLimitHours": rateLimitHours,
		}
		data, err := client.Put("/api/alerts/"+args[0], body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	alertsUpdateCmd.Flags().String("alert-type", "", "Alert type")
	alertsUpdateCmd.Flags().Float64("high-threshold", 0, "High threshold")
	alertsUpdateCmd.Flags().Float64("low-threshold", 0, "Low threshold")
	alertsUpdateCmd.Flags().Bool("enabled", true, "Whether the alert is enabled")
	alertsUpdateCmd.Flags().Int("rate-limit-hours", 0, "Rate limit in hours")
}
