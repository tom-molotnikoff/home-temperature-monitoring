package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

var readingsCmd = &cobra.Command{
	Use:   "readings",
	Short: "Query temperature readings",
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
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		q := url.Values{}
		if sensor != "" {
			q.Set("sensor", sensor)
		}
		if from != "" {
			q.Set("from", from)
		}
		if to != "" {
			q.Set("to", to)
		}
		data, err := client.Get("/api/temperature/readings/between", q)
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
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")

		serverURL, apiKey, insecure, err := loadClientConfig(cmd)
		if err != nil {
			return err
		}
		client := NewClient(serverURL, apiKey, insecure)
		q := url.Values{}
		if sensor != "" {
			q.Set("sensor", sensor)
		}
		if from != "" {
			q.Set("from", from)
		}
		if to != "" {
			q.Set("to", to)
		}
		data, err := client.Get("/api/temperature/readings/hourly/between", q)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	readingsBetweenCmd.Flags().String("sensor", "", "Sensor name")
	readingsBetweenCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	readingsBetweenCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")

	readingsHourlyCmd.Flags().String("sensor", "", "Sensor name")
	readingsHourlyCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	readingsHourlyCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
}
