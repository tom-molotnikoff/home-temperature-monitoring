package cmd

import (
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authMeCmd)
	authCmd.AddCommand(authSessionsCmd)
	authCmd.AddCommand(authRevokeSessionCmd)
	rootCmd.AddCommand(authCmd)
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with username and password",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]string{
			"username": username,
			"password": password,
		}
		data, err := client.Post("/api/auth/login", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	authLoginCmd.Flags().String("username", "", "Username")
	authLoginCmd.Flags().String("password", "", "Password")
	_ = authLoginCmd.MarkFlagRequired("username")
	_ = authLoginCmd.MarkFlagRequired("password")
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "End the current session",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Post("/api/auth/logout", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var authMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Get current authenticated user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/auth/me", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var authSessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List active sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/auth/sessions", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var authRevokeSessionCmd = &cobra.Command{
	Use:   "revoke-session [id]",
	Short: "Revoke a specific session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Delete("/api/auth/sessions/" + args[0])
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}
