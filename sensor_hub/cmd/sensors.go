package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var sensorsCmd = &cobra.Command{
	Use:   "sensors",
	Short: "Manage sensors",
}

func init() {
	sensorsCmd.AddCommand(sensorsListCmd)
	sensorsCmd.AddCommand(sensorsGetCmd)
	sensorsCmd.AddCommand(sensorsAddCmd)
	sensorsCmd.AddCommand(sensorsDeleteCmd)
	sensorsCmd.AddCommand(sensorsEnableCmd)
	sensorsCmd.AddCommand(sensorsDisableCmd)
	sensorsCmd.AddCommand(sensorsHealthCmd)
	sensorsCmd.AddCommand(sensorsStatsCmd)
	sensorsCmd.AddCommand(sensorsCollectCmd)
	rootCmd.AddCommand(sensorsCmd)
}

var sensorsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sensors",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Get("/api/sensors/", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var sensorsGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get a sensor by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Get("/api/sensors/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var sensorsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new sensor",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		sensorType, _ := cmd.Flags().GetString("type")
		url, _ := cmd.Flags().GetString("url")

		if name == "" || sensorType == "" || url == "" {
			return fmt.Errorf("--name, --type, and --url are required")
		}

		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		body := map[string]string{
			"name": name,
			"type": strings.ToLower(sensorType),
			"url":  url,
		}
		data, err := client.Post("/api/sensors/", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	sensorsAddCmd.Flags().String("name", "", "Sensor name")
	sensorsAddCmd.Flags().String("type", "", "Sensor type (e.g. indoor, outdoor)")
	sensorsAddCmd.Flags().String("url", "", "Sensor URL")
}

var sensorsDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Delete("/api/sensors/" + args[0])
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var sensorsEnableCmd = &cobra.Command{
	Use:   "enable [name]",
	Short: "Enable a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Post("/api/sensors/enable/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var sensorsDisableCmd = &cobra.Command{
	Use:   "disable [name]",
	Short: "Disable a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Post("/api/sensors/disable/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var sensorsHealthCmd = &cobra.Command{
	Use:   "health [name]",
	Short: "Get sensor health status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Get("/api/sensors/health/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var sensorsStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get total readings per sensor",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Get("/api/sensors/stats/total-readings", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var sensorsCollectCmd = &cobra.Command{
	Use:   "collect [name]",
	Short: "Trigger sensor data collection (all or specific sensor)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		path := "/api/sensors/collect"
		if len(args) > 0 {
			path = "/api/sensors/collect/" + args[0]
		}
		data, err := client.Post(path, nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}
