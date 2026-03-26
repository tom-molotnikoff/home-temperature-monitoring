package cmd

import (
	"github.com/spf13/cobra"
)

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage roles and permissions",
}

func init() {
	rolesCmd.AddCommand(rolesListCmd)
	rolesCmd.AddCommand(rolesListPermissionsCmd)
	rolesCmd.AddCommand(rolesGetPermissionsCmd)
	rolesCmd.AddCommand(rolesAssignPermissionCmd)
	rolesCmd.AddCommand(rolesRemovePermissionCmd)
	rootCmd.AddCommand(rolesCmd)
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

var rolesListPermissionsCmd = &cobra.Command{
	Use:   "list-permissions",
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

var rolesGetPermissionsCmd = &cobra.Command{
	Use:   "get-permissions [roleId]",
	Short: "Get permissions for a specific role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/roles/"+args[0]+"/permissions", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var rolesAssignPermissionCmd = &cobra.Command{
	Use:   "assign-permission [roleId]",
	Short: "Assign a permission to a role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		permissionId, _ := cmd.Flags().GetInt("permission-id")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]interface{}{
			"permission_id": permissionId,
		}
		data, err := client.Post("/api/roles/"+args[0]+"/permissions", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	rolesAssignPermissionCmd.Flags().Int("permission-id", 0, "Permission ID to assign")
	_ = rolesAssignPermissionCmd.MarkFlagRequired("permission-id")
}

var rolesRemovePermissionCmd = &cobra.Command{
	Use:   "remove-permission [roleId] [permissionId]",
	Short: "Remove a permission from a role",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Delete("/api/roles/" + args[0] + "/permissions/" + args[1])
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}
