package cmd

import (
	"github.com/spf13/cobra"
)

var driversCmd = &cobra.Command{
	Use:   "drivers",
	Short: "Manage sensor drivers",
}

var driversListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available sensor drivers",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/drivers", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(driversCmd)
	driversCmd.AddCommand(driversListCmd)
}
