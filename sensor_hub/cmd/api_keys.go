package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var apiKeysCmd = &cobra.Command{
	Use:   "api-keys",
	Short: "Manage API keys",
}

func init() {
	apiKeysCmd.AddCommand(apiKeysListCmd)
	apiKeysCmd.AddCommand(apiKeysCreateCmd)
	apiKeysCmd.AddCommand(apiKeysRevokeCmd)
	apiKeysCmd.AddCommand(apiKeysDeleteCmd)
	apiKeysCmd.AddCommand(apiKeysUpdateExpiryCmd)
	rootCmd.AddCommand(apiKeysCmd)
}

var apiKeysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/api-keys/", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var apiKeysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]string{"name": name}
		data, err := client.Post("/api/api-keys/", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	apiKeysCreateCmd.Flags().String("name", "", "Name for the API key")
}

var apiKeysRevokeCmd = &cobra.Command{
	Use:   "revoke [id]",
	Short: "Revoke an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Post("/api/api-keys/"+args[0]+"/revoke", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var apiKeysDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Delete("/api/api-keys/" + args[0])
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var apiKeysUpdateExpiryCmd = &cobra.Command{
	Use:   "update-expiry [id]",
	Short: "Update the expiry date of an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		expiresAt, _ := cmd.Flags().GetString("expires-at")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		var body map[string]interface{}
		if expiresAt != "" {
			body = map[string]interface{}{"expires_at": expiresAt}
		} else {
			body = map[string]interface{}{"expires_at": nil}
		}
		data, err := client.Patch("/api/api-keys/"+args[0]+"/expiry", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	apiKeysUpdateExpiryCmd.Flags().String("expires-at", "", "New expiry date (RFC3339 format, e.g. 2026-12-31T23:59:59Z; omit to clear expiry)")
}
