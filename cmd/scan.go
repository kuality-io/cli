package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kuality-io/cli/internal/client"
	"github.com/kuality-io/cli/internal/config"
	"github.com/kuality-io/cli/internal/output"
	"github.com/spf13/cobra"
)

var validScanTypes = []string{
	"a11y", "webvitals", "seo", "formaudit", "brokenlinks", "cookie",
	"headers", "jsaudit", "tech", "cms", "api", "firefox", "webkit",
	"uxaudit", "animation", "colorblind", "assets", "screenreader",
	"performancebudget", "assetaudit", "bundlesize", "ttfb", "throttle",
	"memoryleak", "touchaudit", "touchsize", "orientation", "pwa",
	"mobilelighthouse", "contract", "synthetic", "cdnaudit", "graphql",
	"openapi", "privacyscan", "csp", "cors", "ssl", "email", "dns",
	"web", "port",
}

var (
	flagScanType string
	flagFailOn   string
	flagNoWait   bool
	flagTimeout  int
)

var scanCmd = &cobra.Command{
	Use:   "scan <url>",
	Short: "Run a quality scan on a URL",
	Long: `Run a scan against a URL and wait for results.

By default, scans all quality dimensions. Use --type to run a
specific scan type. Use --fail-on to set a severity threshold
for non-zero exit codes (useful in CI/CD pipelines).

Examples:
  kuality scan example.com
  kuality scan example.com --type a11y
  kuality scan example.com --type a11y --fail-on high
  kuality scan example.com --type seo --format json
  kuality scan example.com --type webvitals --format junit
  kuality scan example.com --no-wait`,
	Args: cobra.ExactArgs(1),
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringVarP(&flagScanType, "type", "t", "web", "Scan type (a11y, seo, webvitals, headers, ssl, ...)")
	scanCmd.Flags().StringVar(&flagFailOn, "fail-on", "", "Exit non-zero if findings at this severity or above (high, medium, low)")
	scanCmd.Flags().BoolVar(&flagNoWait, "no-wait", false, "Start scan and exit without waiting for results")
	scanCmd.Flags().IntVar(&flagTimeout, "timeout", 600, "Maximum seconds to wait for scan completion")

	scanCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validScanTypes, cobra.ShellCompDirectiveNoFileComp
	})
	scanCmd.RegisterFlagCompletionFunc("fail-on", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"high", "medium", "low"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	target := args[0]

	if flagFailOn != "" && flagFailOn != "high" && flagFailOn != "medium" && flagFailOn != "low" {
		return fmt.Errorf("--fail-on must be one of: high, medium, low")
	}

	if !isValidScanType(flagScanType) {
		return fmt.Errorf("unknown scan type %q. Run 'kuality scan --help' for valid types", flagScanType)
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	c, err := client.New(cfg)
	if err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Printf("Starting %s scan on %s...\n", flagScanType, target)
	}

	scan, err := c.CreateScan(target, flagScanType)
	if err != nil {
		return fmt.Errorf("failed to start scan: %w", err)
	}

	if !flagQuiet {
		fmt.Printf("Scan started (ID: %s)\n", scan.ScanID)
	}

	if flagNoWait {
		if flagFormat == "json" {
			return output.JSON(os.Stdout, scan)
		}
		fmt.Printf("Scan ID: %s\nPoll with: kuality status %s\n", scan.ScanID, scan.ScanID)
		return nil
	}

	reportID, err := waitForScan(c, scan.ScanID, time.Duration(flagTimeout)*time.Second)
	if err != nil {
		return err
	}

	if flagFormat == "junit" {
		data, err := c.GetReportJUnit(reportID)
		if err != nil {
			return fmt.Errorf("failed to fetch JUnit report: %w", err)
		}
		os.Stdout.Write(data)
		return nil
	}

	report, err := c.GetReport(reportID)
	if err != nil {
		return fmt.Errorf("failed to fetch report: %w", err)
	}

	if report.State == "failed" {
		return fmt.Errorf("scan failed: %s", report.Error)
	}

	switch flagFormat {
	case "json":
		return output.JSON(os.Stdout, report)
	default:
		printReport(report)
	}

	if flagFailOn != "" {
		return checkThreshold(report, flagFailOn)
	}

	return nil
}

func waitForScan(c *client.Client, scanID string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

	for time.Now().Before(deadline) {
		status, err := c.GetScanStatus(scanID)
		if err != nil {
			return "", fmt.Errorf("failed to check scan status: %w", err)
		}

		state := status.Status
		if state == "" {
			state = status.State
		}

		switch strings.ToLower(state) {
		case "completed":
			if !flagQuiet {
				fmt.Print("\r\033[K")
				fmt.Println("Scan completed.")
			}
			return status.ReportID, nil
		case "failed":
			if !flagQuiet {
				fmt.Print("\r\033[K")
			}
			return status.ReportID, nil
		}

		if !flagQuiet {
			fmt.Printf("\r\033[K%s Scanning... (%s)", spinner[i%len(spinner)], state)
			i++
		}

		time.Sleep(3 * time.Second)
	}

	return "", fmt.Errorf("scan timed out after %s. Check status with: kuality status %s", timeout, scanID)
}

func printReport(r *client.Report) {
	fmt.Println()
	fmt.Printf("  Target:     %s\n", r.Target)
	fmt.Printf("  Scan type:  %s\n", r.TypeOfScan)
	fmt.Printf("  Score:      %s\n", r.Score)
	fmt.Printf("  Status:     %s %s\n", output.StatusIcon(r.State), r.State)
	fmt.Println()

	headers := []string{"Severity", "Count"}
	rows := [][]string{
		{"High", output.SeverityColor("high", r.High)},
		{"Medium", output.SeverityColor("medium", r.Medium)},
		{"Low", output.SeverityColor("low", r.Low)},
		{"Info", fmt.Sprintf("%d", r.Info)},
		{"Total", fmt.Sprintf("%d", r.Total)},
	}
	output.Table(os.Stdout, headers, rows)
	fmt.Println()
}

func checkThreshold(r *client.Report, failOn string) error {
	var count int
	switch failOn {
	case "high":
		count = r.High
	case "medium":
		count = r.High + r.Medium
	case "low":
		count = r.High + r.Medium + r.Low
	}

	if count > 0 {
		return fmt.Errorf("quality gate failed: %d finding(s) at %q severity or above", count, failOn)
	}
	return nil
}

func isValidScanType(t string) bool {
	for _, v := range validScanTypes {
		if v == t {
			return true
		}
	}
	return false
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if flagAPIKey != "" {
		cfg.APIKey = flagAPIKey
	}

	return cfg, nil
}
