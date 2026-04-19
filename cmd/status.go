package cmd

import (
	"fmt"
	"os"

	"github.com/kuality-io/cli/internal/client"
	"github.com/kuality-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <scan-id>",
	Short: "Check the status of a scan",
	Long: `Check the current status of a previously started scan.

Examples:
  kuality status abc123
  kuality status abc123 --format json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		status, err := c.GetScanStatus(args[0])
		if err != nil {
			return fmt.Errorf("failed to get scan status: %w", err)
		}

		if flagFormat == "json" {
			return output.JSON(os.Stdout, status)
		}

		state := status.Status
		if state == "" {
			state = status.State
		}

		fmt.Printf("Scan ID:    %s\n", status.ScanID)
		fmt.Printf("Report ID:  %s\n", status.ReportID)
		fmt.Printf("Target:     %s\n", status.Target)
		fmt.Printf("Status:     %s %s\n", output.StatusIcon(state), state)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
