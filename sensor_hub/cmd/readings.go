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
	rootCmd.AddCommand(readingsCmd)
}

var readingsBetweenCmd = &cobra.Command{
	Use:   "between",
	Short: "Get readings between two dates (auto-aggregated for large ranges)",
	RunE: func(cmd *cobra.Command, args []string) error {
		sensor, _ := cmd.Flags().GetString("sensor")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		aggregation, _ := cmd.Flags().GetString("aggregation")
		aggregationFn, _ := cmd.Flags().GetString("aggregation-function")

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
		if aggregation != "" {
			q.Set("aggregation", aggregation)
		}
		if aggregationFn != "" {
			q.Set("aggregation_function", aggregationFn)
		}
		data, err := client.Get("/api/readings/between", q)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	readingsBetweenCmd.Flags().String("sensor", "", "Sensor name")
	readingsBetweenCmd.Flags().String("start", "", "Start date (YYYY-MM-DD) or datetime (ISO 8601, e.g. 2024-01-15T10:30:00Z)")
	readingsBetweenCmd.Flags().String("end", "", "End date (YYYY-MM-DD) or datetime (ISO 8601, e.g. 2024-01-15T11:30:00Z)")
	readingsBetweenCmd.Flags().String("aggregation", "", "Override aggregation interval (ISO 8601 duration, e.g. PT1H, PT5M)")
	readingsBetweenCmd.Flags().String("aggregation-function", "", "Override aggregation function (avg, min, max, sum, count, last)")
}
