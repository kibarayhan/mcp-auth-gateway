package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-gateway",
	Short: "MCP Auth Gateway — secure proxy for MCP servers",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the gateway",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		fmt.Fprintf(os.Stderr, "Starting gateway with config: %s\n", configPath)
		return nil
	},
}

func init() {
	startCmd.Flags().StringP("config", "c", "gateway.yaml", "Path to config file")
	rootCmd.AddCommand(startCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
