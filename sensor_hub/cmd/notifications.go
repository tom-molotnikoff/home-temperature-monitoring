package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var notificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "Manage notifications",
}

func init() {
	notificationsCmd.AddCommand(notificationsListCmd)
	notificationsCmd.AddCommand(notificationsReadCmd)
	notificationsCmd.AddCommand(notificationsDismissCmd)
	notificationsCmd.AddCommand(notificationsUnreadCountCmd)
	rootCmd.AddCommand(notificationsCmd)
}

var notificationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Get("/api/notifications", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var notificationsReadCmd = &cobra.Command{
	Use:   "read [id]",
	Short: "Mark a notification as read",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("notification ID must be a number")
		}
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Post("/api/notifications/"+args[0]+"/read", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var notificationsDismissCmd = &cobra.Command{
	Use:   "dismiss [id]",
	Short: "Dismiss a notification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("notification ID must be a number")
		}
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Post("/api/notifications/"+args[0]+"/dismiss", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var notificationsUnreadCountCmd = &cobra.Command{
	Use:   "unread-count",
	Short: "Get count of unread notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey)
		data, err := client.Get("/api/notifications/unread-count", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}
