package cmd

import (
	"fmt"
	"os"

	"github.com/kuality-io/cli/internal/client"
	"github.com/kuality-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "Show Kuality Scores for your targets",
	Long: `Show the current Kuality Score for all targets in your organization.

Examples:
  kuality score
  kuality score --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		scores, err := c.ListScores()
		if err != nil {
			return fmt.Errorf("failed to get scores: %w", err)
		}

		if flagFormat == "json" {
			return output.JSON(os.Stdout, scores)
		}

		if len(scores) == 0 {
			fmt.Println("No scores available. Run a scan first.")
			return nil
		}

		headers := []string{"Target", "Score", "Type"}
		rows := make([][]string, len(scores))
		for i, s := range scores {
			rows[i] = []string{s.Target, s.Score.String(), s.Type}
		}
		output.Table(os.Stdout, headers, rows)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scoreCmd)
}
