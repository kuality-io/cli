package cmd

import (
	"fmt"
	"os"

	"github.com/kuality-io/cli/internal/client"
	"github.com/kuality-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagReportType   string
	flagReportTarget string
)

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "View scan reports",
}

var reportsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent reports",
	Long: `List recent scan reports for your organization.

Examples:
  kuality reports list
  kuality reports list --type a11y
  kuality reports list --target example.com
  kuality reports list --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		reports, err := c.ListReports(flagReportType, flagReportTarget)
		if err != nil {
			return fmt.Errorf("failed to list reports: %w", err)
		}

		if flagFormat == "json" {
			return output.JSON(os.Stdout, reports)
		}

		if len(reports) == 0 {
			fmt.Println("No reports found.")
			return nil
		}

		headers := []string{"ID", "Target", "Type", "Status", "Score", "High", "Med", "Low"}
		rows := make([][]string, len(reports))
		for i, r := range reports {
			id := r.ID
			if len(id) > 8 {
				id = id[:8]
			}
			rows[i] = []string{
				id,
				r.Target,
				r.TypeOfScan,
				fmt.Sprintf("%s %s", output.StatusIcon(r.State), r.State),
				r.Score.String(),
				output.SeverityColor("high", r.High),
				output.SeverityColor("medium", r.Medium),
				output.SeverityColor("low", r.Low),
			}
		}
		output.Table(os.Stdout, headers, rows)
		return nil
	},
}

var reportsShowCmd = &cobra.Command{
	Use:   "show <report-id>",
	Short: "Show detailed report",
	Long: `Show the full details of a scan report.

Examples:
  kuality reports show abc123
  kuality reports show abc123 --format json
  kuality reports show abc123 --format junit`,
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

		if flagFormat == "junit" {
			data, err := c.GetReportJUnit(args[0])
			if err != nil {
				return fmt.Errorf("failed to fetch JUnit report: %w", err)
			}
			os.Stdout.Write(data)
			return nil
		}

		report, err := c.GetReport(args[0])
		if err != nil {
			return fmt.Errorf("failed to get report: %w", err)
		}

		if flagFormat == "json" {
			return output.JSON(os.Stdout, report)
		}

		fmt.Println()
		fmt.Printf("  Report ID:  %s\n", report.ID)
		fmt.Printf("  Target:     %s\n", report.Target)
		fmt.Printf("  Scan type:  %s\n", report.TypeOfScan)
		fmt.Printf("  Score:      %s\n", report.Score)
		fmt.Printf("  Status:     %s %s\n", output.StatusIcon(report.State), report.State)
		if report.StartDate != "" {
			fmt.Printf("  Started:    %s\n", report.StartDate)
		}
		if report.EndDate != "" {
			fmt.Printf("  Completed:  %s\n", report.EndDate)
		}
		if report.Error != "" {
			fmt.Printf("  Error:      %s\n", report.Error)
		}
		fmt.Println()

		headers := []string{"Severity", "Count"}
		rows := [][]string{
			{"High", output.SeverityColor("high", report.High)},
			{"Medium", output.SeverityColor("medium", report.Medium)},
			{"Low", output.SeverityColor("low", report.Low)},
			{"Info", fmt.Sprintf("%d", report.Info)},
			{"Total", fmt.Sprintf("%d", report.Total)},
		}
		output.Table(os.Stdout, headers, rows)
		fmt.Println()

		return nil
	},
}

func init() {
	reportsListCmd.Flags().StringVarP(&flagReportType, "type", "t", "", "Filter by scan type")
	reportsListCmd.Flags().StringVar(&flagReportTarget, "target", "", "Filter by target URL")

	reportsCmd.AddCommand(reportsListCmd)
	reportsCmd.AddCommand(reportsShowCmd)
	rootCmd.AddCommand(reportsCmd)
}
