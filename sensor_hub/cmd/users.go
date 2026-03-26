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

func init() {
	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersCreateCmd)
	usersCmd.AddCommand(usersDeleteCmd)
	usersCmd.AddCommand(usersChangePasswordCmd)
	usersCmd.AddCommand(usersSetMustChangeCmd)
	usersCmd.AddCommand(usersSetRolesCmd)
	rootCmd.AddCommand(usersCmd)
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

var usersChangePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Change a user's password",
	RunE: func(cmd *cobra.Command, args []string) error {
		userId, _ := cmd.Flags().GetInt("user-id")
		newPassword, _ := cmd.Flags().GetString("new-password")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]interface{}{
			"user_id":      userId,
			"new_password": newPassword,
		}
		data, err := client.Put("/api/users/password", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	usersChangePasswordCmd.Flags().Int("user-id", 0, "User ID")
	usersChangePasswordCmd.Flags().String("new-password", "", "New password")
	_ = usersChangePasswordCmd.MarkFlagRequired("user-id")
	_ = usersChangePasswordCmd.MarkFlagRequired("new-password")
}

var usersSetMustChangeCmd = &cobra.Command{
	Use:   "set-must-change [id]",
	Short: "Set whether a user must change their password",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mustChange, _ := cmd.Flags().GetBool("must-change")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]interface{}{
			"must_change": mustChange,
		}
		data, err := client.Patch("/api/users/"+args[0]+"/must_change", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	usersSetMustChangeCmd.Flags().Bool("must-change", false, "Whether user must change password on next login")
}

var usersSetRolesCmd = &cobra.Command{
	Use:   "set-roles [id]",
	Short: "Set roles for a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		roles, _ := cmd.Flags().GetStringSlice("roles")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]interface{}{
			"roles": roles,
		}
		data, err := client.Post("/api/users/"+args[0]+"/roles", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	usersSetRolesCmd.Flags().StringSlice("roles", nil, "Comma-separated list of role names")
}
