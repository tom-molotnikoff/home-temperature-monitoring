package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "sensor-hub",
	Short: "Home temperature monitoring system",
	Long:  "Sensor Hub — a home temperature monitoring system.\nRun as a server with 'serve' or use CLI commands to interact with a remote instance.",
}

func Execute(version string) {
	Version = version
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
