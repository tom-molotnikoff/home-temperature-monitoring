package cmd

import (
	"github.com/spf13/cobra"
)

var propertiesCmd = &cobra.Command{
	Use:   "properties",
	Short: "Manage application properties",
}

func init() {
	propertiesCmd.AddCommand(propertiesGetCmd)
	propertiesCmd.AddCommand(propertiesSetCmd)
	rootCmd.AddCommand(propertiesCmd)
}

var propertiesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get all application properties",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/properties", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var propertiesSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Update application properties",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]string{key: value}
		data, err := client.Patch("/api/properties", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	propertiesSetCmd.Flags().String("key", "", "Property key")
	propertiesSetCmd.Flags().String("value", "", "Property value")
}
