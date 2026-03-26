package cmd

import (
	"github.com/spf13/cobra"
)

var oauthCmd = &cobra.Command{
	Use:   "oauth",
	Short: "OAuth configuration commands",
}

func init() {
	oauthCmd.AddCommand(oauthStatusCmd)
	oauthCmd.AddCommand(oauthAuthorizeCmd)
	oauthCmd.AddCommand(oauthSubmitCodeCmd)
	oauthCmd.AddCommand(oauthReloadCmd)
	rootCmd.AddCommand(oauthCmd)
}

var oauthStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get OAuth configuration status",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/oauth/status", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var oauthAuthorizeCmd = &cobra.Command{
	Use:   "authorize",
	Short: "Start OAuth authorization flow",
	Long:  "Initiates the OAuth authorization flow and returns an auth URL to visit along with a state token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/oauth/authorize", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var oauthSubmitCodeCmd = &cobra.Command{
	Use:   "submit-code",
	Short: "Submit an OAuth authorization code",
	Long:  "Submit the authorization code received from the OAuth provider along with the state token from the authorize step.",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, _ := cmd.Flags().GetString("code")
		state, _ := cmd.Flags().GetString("state")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]string{
			"code":  code,
			"state": state,
		}
		data, err := client.Post("/api/oauth/submit-code", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	oauthSubmitCodeCmd.Flags().String("code", "", "Authorization code from OAuth provider")
	oauthSubmitCodeCmd.Flags().String("state", "", "State token from the authorize step")
	_ = oauthSubmitCodeCmd.MarkFlagRequired("code")
	_ = oauthSubmitCodeCmd.MarkFlagRequired("state")
}

var oauthReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload OAuth configuration from disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Post("/api/oauth/reload", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}
