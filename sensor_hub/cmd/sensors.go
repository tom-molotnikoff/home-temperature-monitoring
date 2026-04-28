package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	gen "example/sensorHub/gen"
)

var sensorsCmd = &cobra.Command{
	Use:   "sensors",
	Short: "Manage sensors",
}

func init() {
	sensorsCmd.AddCommand(sensorsListCmd)
	sensorsCmd.AddCommand(sensorsGetCmd)
	sensorsCmd.AddCommand(sensorsAddCmd)
	sensorsCmd.AddCommand(sensorsUpdateCmd)
	sensorsCmd.AddCommand(sensorsDeleteCmd)
	sensorsCmd.AddCommand(sensorsExistsCmd)
	sensorsCmd.AddCommand(sensorsListByDriverCmd)
	sensorsCmd.AddCommand(sensorsEnableCmd)
	sensorsCmd.AddCommand(sensorsDisableCmd)
	sensorsCmd.AddCommand(sensorsHealthCmd)
	sensorsCmd.AddCommand(sensorsStatsCmd)
	sensorsCmd.AddCommand(sensorsCollectCmd)
	sensorsCmd.AddCommand(sensorsPendingCmd)
	sensorsCmd.AddCommand(sensorsApproveCmd)
	sensorsCmd.AddCommand(sensorsDismissCmd)
	rootCmd.AddCommand(sensorsCmd)
}

var sensorsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sensors",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.GetAllSensors(ctx))
	},
}

var sensorsGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get a sensor by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.GetSensorByName(ctx, args[0]))
	},
}

// sensorBody builds a partial JSON body matching the historical CLI wire
// format ({name, sensor_driver, config}) for AddSensor / UpdateSensorById.
// We use the *WithBody* generated client variants so we don't need to
// populate every field of `gen.Sensor`, several of which are server-managed
// (id, status, health_*).
func sensorBody(name, driver string, config map[string]string) map[string]any {
	return map[string]any{
		"name":          name,
		"sensor_driver": driver,
		"config":        config,
	}
}

var sensorsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new sensor",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		driver, _ := cmd.Flags().GetString("driver")
		configPairs, _ := cmd.Flags().GetStringSlice("config")
		if name == "" || driver == "" {
			return fmt.Errorf("--name and --driver are required")
		}
		config, err := parseKVPairs(configPairs)
		if err != nil {
			return err
		}

		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		body, err := rawJSONReader(sensorBody(name, driver, config))
		if err != nil {
			return err
		}
		return consumeJSON(client.AddSensorWithBody(ctx, "application/json", body))
	},
}

func init() {
	sensorsAddCmd.Flags().String("name", "", "Sensor name")
	sensorsAddCmd.Flags().String("driver", "", "Sensor driver (e.g. sensor-hub-http-temperature)")
	sensorsAddCmd.Flags().StringSlice("config", nil, "Config key=value pairs (repeatable, e.g. --config url=http://...)")
}

var sensorsDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.DeleteSensorByName(ctx, args[0]))
	},
}

var sensorsEnableCmd = &cobra.Command{
	Use:   "enable [name]",
	Short: "Enable a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.EnableSensor(ctx, args[0]))
	},
}

var sensorsDisableCmd = &cobra.Command{
	Use:   "disable [name]",
	Short: "Disable a sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.DisableSensor(ctx, args[0]))
	},
}

var sensorsHealthCmd = &cobra.Command{
	Use:   "health [name]",
	Short: "Get sensor health status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		params := &gen.GetSensorHealthHistoryByNameParams{}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			params.Limit = &limit
		}
		return consumeJSON(client.GetSensorHealthHistoryByName(ctx, args[0], params))
	},
}

func init() {
	sensorsHealthCmd.Flags().Int("limit", 0, "Maximum number of health records to return")
}

var sensorsStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get total readings per sensor",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.GetTotalReadingsPerSensor(ctx))
	},
}

var sensorsCollectCmd = &cobra.Command{
	Use:   "collect [name]",
	Short: "Trigger sensor data collection (all or specific sensor)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		if len(args) > 0 {
			return consumeJSON(client.CollectFromSensor(ctx, args[0]))
		}
		return consumeJSON(client.CollectAllSensorReadings(ctx))
	},
}

var sensorsUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update an existing sensor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("sensor ID must be a number")
		}
		name, _ := cmd.Flags().GetString("name")
		driver, _ := cmd.Flags().GetString("driver")
		configPairs, _ := cmd.Flags().GetStringSlice("config")
		config, err := parseKVPairs(configPairs)
		if err != nil {
			return err
		}

		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		body, err := rawJSONReader(sensorBody(name, driver, config))
		if err != nil {
			return err
		}
		return consumeJSON(client.UpdateSensorByIdWithBody(ctx, id, "application/json", body))
	},
}

func init() {
	sensorsUpdateCmd.Flags().String("name", "", "Sensor name")
	sensorsUpdateCmd.Flags().String("driver", "", "Sensor driver")
	sensorsUpdateCmd.Flags().StringSlice("config", nil, "Config key=value pairs (repeatable)")
}

var sensorsExistsCmd = &cobra.Command{
	Use:   "exists [name]",
	Short: "Check if a sensor exists",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		status, err := consumeStatus(client.SensorExists(ctx, args[0]))
		if err != nil {
			return err
		}
		if status == 200 {
			fmt.Println("Sensor exists")
		} else {
			fmt.Println("Sensor not found")
		}
		return nil
	},
}

var sensorsListByDriverCmd = &cobra.Command{
	Use:   "list-by-driver [driver]",
	Short: "List sensors by driver",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.GetSensorsByDriver(ctx, args[0]))
	},
}

var sensorsPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List pending (auto-discovered) sensors awaiting approval",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.GetSensorsByStatus(ctx, gen.GetSensorsByStatusParamsStatus("pending")))
	},
}

var sensorsApproveCmd = &cobra.Command{
	Use:   "approve [id]",
	Short: "Approve a pending sensor (sets status to active)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("sensor ID must be a number")
		}
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.ApproveSensor(ctx, id))
	},
}

var sensorsDismissCmd = &cobra.Command{
	Use:   "dismiss [id]",
	Short: "Dismiss a pending sensor (hides from pending list)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("sensor ID must be a number")
		}
		client, ctx, err := newAPIClient(cmd)
		if err != nil {
			return err
		}
		return consumeJSON(client.DismissSensor(ctx, id))
	},
}
