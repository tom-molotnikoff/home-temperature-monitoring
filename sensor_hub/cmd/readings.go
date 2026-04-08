package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

var readingsCmd = &cobra.Command{
	Use:   "readings",
	Short: "Query sensor readings",
}

func init() {
	readingsCmd.AddCommand(readingsBetweenCmd)
	readingsCmd.AddCommand(readingsHourlyCmd)
	rootCmd.AddCommand(readingsCmd)
}

var readingsBetweenCmd = &cobra.Command{
	Use:   "between",
	Short: "Get readings between two dates",
	RunE: func(cmd *cobra.Command, args []string) error {
		sensor, _ := cmd.Flags().GetString("sensor")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		q := url.Values{}
		if sensor != "" {
			q.Set("sensor", sensor)
		}
		if start != "" {
			q.Set("start", start)
		}
		if end != "" {
			q.Set("end", end)
		}
		data, err := client.Get("/api/readings/between", q)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var readingsHourlyCmd = &cobra.Command{
	Use:   "hourly",
	Short: "Get hourly average readings between two dates",
	RunE: func(cmd *cobra.Command, args []string) error {
		sensor, _ := cmd.Flags().GetString("sensor")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		q := url.Values{}
		if sensor != "" {
			q.Set("sensor", sensor)
		}
		if start != "" {
			q.Set("start", start)
		}
		if end != "" {
			q.Set("end", end)
		}
		data, err := client.Get("/api/readings/hourly/between", q)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	readingsBetweenCmd.Flags().String("sensor", "", "Sensor name")
	readingsBetweenCmd.Flags().String("start", "", "Start date (YYYY-MM-DD)")
	readingsBetweenCmd.Flags().String("end", "", "End date (YYYY-MM-DD)")

	readingsHourlyCmd.Flags().String("sensor", "", "Sensor name")
	readingsHourlyCmd.Flags().String("start", "", "Start date (YYYY-MM-DD)")
	readingsHourlyCmd.Flags().String("end", "", "End date (YYYY-MM-DD)")
}
