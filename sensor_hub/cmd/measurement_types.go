package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var measurementTypesCmd = &cobra.Command{
	Use:   "measurement-types",
	Short: "Query measurement types",
}

func init() {
	measurementTypesCmd.AddCommand(measurementTypesListCmd)
	measurementTypesCmd.AddCommand(measurementTypesForSensorCmd)
	rootCmd.AddCommand(measurementTypesCmd)
}

var measurementTypesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all measurement types",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		var params url.Values
		hasReadings, _ := cmd.Flags().GetBool("has-readings")
		if hasReadings {
			params = url.Values{"has_readings": {"true"}}
		}
		data, err := client.Get("/api/measurement-types", params)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var measurementTypesForSensorCmd = &cobra.Command{
	Use:   "for-sensor [sensor-id]",
	Short: "List measurement types supported by a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get(fmt.Sprintf("/api/sensors/by-id/%s/measurement-types", args[0]), nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	measurementTypesListCmd.Flags().Bool("has-readings", false, "Only return types that have at least one reading")
}
