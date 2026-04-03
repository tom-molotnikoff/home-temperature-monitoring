package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dashboardsCmd = &cobra.Command{
	Use:   "dashboards",
	Short: "Manage dashboards",
}

func init() {
	dashboardsCreateCmd.Flags().String("name", "", "Dashboard name")
	_ = dashboardsCreateCmd.MarkFlagRequired("name")
	dashboardsUpdateCmd.Flags().String("file", "", "Path to JSON file with dashboard data")
	_ = dashboardsUpdateCmd.MarkFlagRequired("file")

	dashboardsCmd.AddCommand(dashboardsListCmd)
	dashboardsCmd.AddCommand(dashboardsGetCmd)
	dashboardsCmd.AddCommand(dashboardsCreateCmd)
	dashboardsCmd.AddCommand(dashboardsDeleteCmd)
	dashboardsCmd.AddCommand(dashboardsUpdateCmd)
	rootCmd.AddCommand(dashboardsCmd)
}

var dashboardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all dashboards",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/dashboards/", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var dashboardsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a dashboard by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get(fmt.Sprintf("/api/dashboards/%s", id), nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var dashboardsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := struct {
			Name string `json:"name"`
		}{Name: name}
		data, err := client.Post("/api/dashboards/", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var dashboardsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a dashboard by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Delete(fmt.Sprintf("/api/dashboards/%s", id))
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var dashboardsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a dashboard by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		filePath, _ := cmd.Flags().GetString("file")
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		var body json.RawMessage = fileData
		data, err := client.Put(fmt.Sprintf("/api/dashboards/%s", id), body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}
