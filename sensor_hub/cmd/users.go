package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
}

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage roles and permissions",
}

func init() {
	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersGetCmd)
	usersCmd.AddCommand(usersCreateCmd)
	usersCmd.AddCommand(usersDeleteCmd)
	rootCmd.AddCommand(usersCmd)

	rolesCmd.AddCommand(rolesListCmd)
	rolesCmd.AddCommand(rolesPermissionsCmd)
	rootCmd.AddCommand(rolesCmd)
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/users/", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var usersGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a user by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("user ID must be a number")
		}
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/users/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var usersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		email, _ := cmd.Flags().GetString("email")

		if username == "" || password == "" {
			return fmt.Errorf("--username and --password are required")
		}

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]string{
			"username": username,
			"password": password,
		}
		if email != "" {
			body["email"] = email
		}
		data, err := client.Post("/api/users/", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	usersCreateCmd.Flags().String("username", "", "Username")
	usersCreateCmd.Flags().String("password", "", "Password")
	usersCreateCmd.Flags().String("email", "", "Email (optional)")
}

var usersDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a user by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("user ID must be a number")
		}
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Delete("/api/users/" + args[0])
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var rolesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all roles",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/roles/", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var rolesPermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "List all permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/roles/permissions", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}
