package cmd

import (
	"fmt"
	"net/url"
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
	notificationsCmd.AddCommand(notificationsBulkReadCmd)
	notificationsCmd.AddCommand(notificationsBulkDismissCmd)
	notificationsCmd.AddCommand(notificationsPreferencesCmd)
	notificationsCmd.AddCommand(notificationsSetPreferenceCmd)
	rootCmd.AddCommand(notificationsCmd)
}

var notificationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		query := url.Values{}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			query.Set("limit", strconv.Itoa(limit))
		}
		if offset, _ := cmd.Flags().GetInt("offset"); offset > 0 {
			query.Set("offset", strconv.Itoa(offset))
		}
		if includeDismissed, _ := cmd.Flags().GetBool("include-dismissed"); includeDismissed {
			query.Set("include_dismissed", "true")
		}
		data, err := client.Get("/api/notifications", query)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	notificationsListCmd.Flags().Int("limit", 0, "Maximum number of notifications (default 50)")
	notificationsListCmd.Flags().Int("offset", 0, "Offset for pagination")
	notificationsListCmd.Flags().Bool("include-dismissed", false, "Include dismissed notifications")
}

var notificationsReadCmd = &cobra.Command{
	Use:   "read [id]",
	Short: "Mark a notification as read",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("notification ID must be a number")
		}
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
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
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
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
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/notifications/unread-count", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var notificationsBulkReadCmd = &cobra.Command{
	Use:   "bulk-read",
	Short: "Mark all notifications as read",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Post("/api/notifications/bulk/read", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var notificationsBulkDismissCmd = &cobra.Command{
	Use:   "bulk-dismiss",
	Short: "Dismiss all notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Post("/api/notifications/bulk/dismiss", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var notificationsPreferencesCmd = &cobra.Command{
	Use:   "preferences",
	Short: "Get notification channel preferences",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/notifications/preferences", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var notificationsSetPreferenceCmd = &cobra.Command{
	Use:   "set-preference",
	Short: "Set a notification channel preference",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		emailEnabled, _ := cmd.Flags().GetBool("email-enabled")
		inappEnabled, _ := cmd.Flags().GetBool("inapp-enabled")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		body := map[string]interface{}{
			"category":      category,
			"email_enabled": emailEnabled,
			"inapp_enabled": inappEnabled,
		}
		data, err := client.Post("/api/notifications/preferences", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	notificationsSetPreferenceCmd.Flags().String("category", "", "Notification category (required)")
	notificationsSetPreferenceCmd.Flags().Bool("email-enabled", false, "Enable email notifications")
	notificationsSetPreferenceCmd.Flags().Bool("inapp-enabled", false, "Enable in-app notifications")
	_ = notificationsSetPreferenceCmd.MarkFlagRequired("category")
}
