package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var (
	flagAPIKey string
	flagFormat string
	flagQuiet  bool
)

var rootCmd = &cobra.Command{
	Use:   "kuality",
	Short: "Kuality CLI — scan any site from your terminal",
	Long: `Kuality CLI lets you run website quality scans from the command line.

Scan for accessibility, performance, SEO, security, and 30+ other
quality dimensions. Integrate into CI/CD pipelines with exit codes
and JUnit output.

Get your API key at https://kuality.io/settings/api-keys`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       Version,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key (overrides KUALITY_API_KEY and config file)")
	rootCmd.PersistentFlags().StringVarP(&flagFormat, "format", "f", "table", "Output format: table, json, junit")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress progress output")

	rootCmd.SetVersionTemplate(fmt.Sprintf("kuality %s (commit: %s, built: %s)\n", Version, Commit, Date))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
