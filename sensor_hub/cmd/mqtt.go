package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var mqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "Manage MQTT brokers and subscriptions",
}

// ============================================================================
// Broker commands
// ============================================================================

var mqttBrokersCmd = &cobra.Command{
	Use:   "brokers",
	Short: "Manage MQTT brokers",
}

var mqttBrokersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all MQTT brokers",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get("/api/mqtt/brokers", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mqttBrokersGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get an MQTT broker by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get(fmt.Sprintf("/api/mqtt/brokers/%s", args[0]), nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mqttBrokersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new MQTT broker",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		brokerType, _ := cmd.Flags().GetString("type")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		enabled, _ := cmd.Flags().GetBool("enabled")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		useTLS, _ := cmd.Flags().GetBool("tls")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{
			"name":    name,
			"type":    brokerType,
			"host":    host,
			"port":    port,
			"enabled": enabled,
			"use_tls": useTLS,
		}
		if username != "" {
			body["username"] = username
		}
		if password != "" {
			body["password"] = password
		}

		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Post("/api/mqtt/brokers", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mqttBrokersUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update an MQTT broker",
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
		data, err := client.Put(fmt.Sprintf("/api/mqtt/brokers/%s", id), body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mqttBrokersDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete an MQTT broker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Delete(fmt.Sprintf("/api/mqtt/brokers/%s", args[0]))
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mqttBrokersEnableCmd = &cobra.Command{
	Use:   "enable [id]",
	Short: "Enable an MQTT broker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return toggleBrokerEnabled(cmd, args[0], true)
	},
}

var mqttBrokersDisableCmd = &cobra.Command{
	Use:   "disable [id]",
	Short: "Disable an MQTT broker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return toggleBrokerEnabled(cmd, args[0], false)
	},
}

func toggleBrokerEnabled(cmd *cobra.Command, idStr string, enabled bool) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("invalid broker ID: %s", idStr)
	}

	serverURL, apiKey, insecure, err := loadClientConfig(cmd)
	if err != nil {
		return err
	}
	client := NewClient(serverURL, apiKey, insecure)

	// Fetch current broker
	brokerData, err := client.Get(fmt.Sprintf("/api/mqtt/brokers/%d", id), nil)
	if err != nil {
		return err
	}
	var broker map[string]interface{}
	if err := json.Unmarshal(brokerData, &broker); err != nil {
		return err
	}
	broker["enabled"] = enabled

	data, err := client.Put(fmt.Sprintf("/api/mqtt/brokers/%d", id), broker)
	if err != nil {
		return err
	}
	printJSON(data)
	return nil
}

// ============================================================================
// Subscription commands
// ============================================================================

var mqttSubscriptionsCmd = &cobra.Command{
	Use:   "subscriptions",
	Short: "Manage MQTT subscriptions",
}

var mqttSubscriptionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List MQTT subscriptions",
	RunE: func(cmd *cobra.Command, args []string) error {
		brokerID, _ := cmd.Flags().GetInt("broker-id")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)

		var query url.Values
		if brokerID > 0 {
			query = url.Values{"broker_id": {strconv.Itoa(brokerID)}}
		}

		data, err := client.Get("/api/mqtt/subscriptions", query)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mqttSubscriptionsGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get an MQTT subscription by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Get(fmt.Sprintf("/api/mqtt/subscriptions/%s", args[0]), nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mqttSubscriptionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new MQTT subscription",
	RunE: func(cmd *cobra.Command, args []string) error {
		brokerID, _ := cmd.Flags().GetInt("broker-id")
		topic, _ := cmd.Flags().GetString("topic")
		driverType, _ := cmd.Flags().GetString("driver")
		qos, _ := cmd.Flags().GetInt("qos")
		enabled, _ := cmd.Flags().GetBool("enabled")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{
			"broker_id":     brokerID,
			"topic_pattern": topic,
			"driver_type":   driverType,
			"qos":           qos,
			"enabled":       enabled,
		}

		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Post("/api/mqtt/subscriptions", body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mqttSubscriptionsUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update an MQTT subscription",
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
		data, err := client.Put(fmt.Sprintf("/api/mqtt/subscriptions/%s", id), body)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mqttSubscriptionsDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete an MQTT subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		data, err := client.Delete(fmt.Sprintf("/api/mqtt/subscriptions/%s", args[0]))
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	// Broker flags
	mqttBrokersCreateCmd.Flags().String("name", "", "Broker name")
	mqttBrokersCreateCmd.Flags().String("type", "external", "Broker type (embedded or external)")
	mqttBrokersCreateCmd.Flags().String("host", "", "Broker host")
	mqttBrokersCreateCmd.Flags().Int("port", 1883, "Broker port")
	mqttBrokersCreateCmd.Flags().Bool("enabled", true, "Enable the broker")
	mqttBrokersCreateCmd.Flags().String("username", "", "Broker username")
	mqttBrokersCreateCmd.Flags().String("password", "", "Broker password")
	mqttBrokersCreateCmd.Flags().Bool("tls", false, "Use TLS")
	_ = mqttBrokersCreateCmd.MarkFlagRequired("name")
	_ = mqttBrokersCreateCmd.MarkFlagRequired("host")

	mqttBrokersUpdateCmd.Flags().String("file", "", "Path to JSON file with broker data")
	_ = mqttBrokersUpdateCmd.MarkFlagRequired("file")

	// Subscription flags
	mqttSubscriptionsListCmd.Flags().Int("broker-id", 0, "Filter by broker ID")

	mqttSubscriptionsCreateCmd.Flags().Int("broker-id", 0, "Broker ID")
	mqttSubscriptionsCreateCmd.Flags().String("topic", "", "MQTT topic pattern")
	mqttSubscriptionsCreateCmd.Flags().String("driver", "", "Driver type (e.g. mqtt-zigbee2mqtt)")
	mqttSubscriptionsCreateCmd.Flags().Int("qos", 0, "MQTT QoS level (0, 1, or 2)")
	mqttSubscriptionsCreateCmd.Flags().Bool("enabled", true, "Enable the subscription")
	_ = mqttSubscriptionsCreateCmd.MarkFlagRequired("broker-id")
	_ = mqttSubscriptionsCreateCmd.MarkFlagRequired("topic")
	_ = mqttSubscriptionsCreateCmd.MarkFlagRequired("driver")

	mqttSubscriptionsUpdateCmd.Flags().String("file", "", "Path to JSON file with subscription data")
	_ = mqttSubscriptionsUpdateCmd.MarkFlagRequired("file")

	// Wire up command tree
	mqttBrokersCmd.AddCommand(mqttBrokersListCmd)
	mqttBrokersCmd.AddCommand(mqttBrokersGetCmd)
	mqttBrokersCmd.AddCommand(mqttBrokersCreateCmd)
	mqttBrokersCmd.AddCommand(mqttBrokersUpdateCmd)
	mqttBrokersCmd.AddCommand(mqttBrokersDeleteCmd)
	mqttBrokersCmd.AddCommand(mqttBrokersEnableCmd)
	mqttBrokersCmd.AddCommand(mqttBrokersDisableCmd)

	mqttSubscriptionsCmd.AddCommand(mqttSubscriptionsListCmd)
	mqttSubscriptionsCmd.AddCommand(mqttSubscriptionsGetCmd)
	mqttSubscriptionsCmd.AddCommand(mqttSubscriptionsCreateCmd)
	mqttSubscriptionsCmd.AddCommand(mqttSubscriptionsUpdateCmd)
	mqttSubscriptionsCmd.AddCommand(mqttSubscriptionsDeleteCmd)

	mqttCmd.AddCommand(mqttBrokersCmd)
	mqttCmd.AddCommand(mqttSubscriptionsCmd)
	rootCmd.AddCommand(mqttCmd)
}
