package cmd

import (
	"fmt"
	"os"

	"github.com/kuality-io/cli/internal/client"
	"github.com/kuality-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var targetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "List configured targets",
	Long: `List all scan targets configured in your organization.

Examples:
  kuality targets
  kuality targets --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		targets, err := c.ListTargets()
		if err != nil {
			return fmt.Errorf("failed to list targets: %w", err)
		}

		if flagFormat == "json" {
			return output.JSON(os.Stdout, targets)
		}

		if len(targets) == 0 {
			fmt.Println("No targets configured.")
			return nil
		}

		headers := []string{"ID", "URL", "Type"}
		rows := make([][]string, len(targets))
		for i, t := range targets {
			id := t.ID
			if len(id) > 8 {
				id = id[:8]
			}
			rows[i] = []string{id, t.URL, t.Type}
		}
		output.Table(os.Stdout, headers, rows)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(targetsCmd)
}
