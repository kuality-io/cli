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

var validTestTypes = []string{
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
	flagTestType string
	flagFailOn   string
	flagNoWait   bool
	flagTimeout  int
)

var testCmd = &cobra.Command{
	Use:   "test <url>",
	Short: "Run a quality test on a URL",
	Long: `Run a test against a URL and wait for results.

By default, tests all quality dimensions. Use --type to run a
specific test type. Use --fail-on to set a severity threshold
for non-zero exit codes (useful in CI/CD pipelines).

Examples:
  kuality test example.com
  kuality test example.com --type a11y
  kuality test example.com --type a11y --fail-on high
  kuality test example.com --type seo --format json
  kuality test example.com --type webvitals --format junit
  kuality test example.com --no-wait`,
	Args: cobra.ExactArgs(1),
	RunE: runTest,
}

func init() {
	testCmd.Flags().StringVarP(&flagTestType, "type", "t", "web", "Test type (a11y, seo, webvitals, headers, ssl, ...)")
	testCmd.Flags().StringVar(&flagFailOn, "fail-on", "", "Exit non-zero if findings at this severity or above (high, medium, low)")
	testCmd.Flags().BoolVar(&flagNoWait, "no-wait", false, "Start test and exit without waiting for results")
	testCmd.Flags().IntVar(&flagTimeout, "timeout", 600, "Maximum seconds to wait for test completion")

	testCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validTestTypes, cobra.ShellCompDirectiveNoFileComp
	})
	testCmd.RegisterFlagCompletionFunc("fail-on", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"high", "medium", "low"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.AddCommand(testCmd)
}

func runTest(cmd *cobra.Command, args []string) error {
	target := args[0]

	if flagFailOn != "" && flagFailOn != "high" && flagFailOn != "medium" && flagFailOn != "low" {
		return fmt.Errorf("--fail-on must be one of: high, medium, low")
	}

	if !isValidTestType(flagTestType) {
		return fmt.Errorf("unknown test type %q. Run 'kuality test --help' for valid types", flagTestType)
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
		fmt.Printf("Starting %s test on %s...\n", flagTestType, target)
	}

	test, err := c.CreateTest(target, flagTestType)
	if err != nil {
		return fmt.Errorf("failed to start test: %w", err)
	}

	if !flagQuiet {
		fmt.Printf("Test started (ID: %s)\n", test.TestID)
	}

	if flagNoWait {
		if flagFormat == "json" {
			return output.JSON(os.Stdout, test)
		}
		fmt.Printf("Test ID: %s\nPoll with: kuality status %s\n", test.TestID, test.TestID)
		return nil
	}

	reportID, err := waitForTest(c, test.TestID, time.Duration(flagTimeout)*time.Second)
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
		return fmt.Errorf("test failed: %s", report.Error)
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

func waitForTest(c *client.Client, testID string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

	for time.Now().Before(deadline) {
		status, err := c.GetTestStatus(testID)
		if err != nil {
			return "", fmt.Errorf("failed to check test status: %w", err)
		}

		state := status.Status
		if state == "" {
			state = status.State
		}

		switch strings.ToLower(state) {
		case "completed":
			if !flagQuiet {
				fmt.Print("\r\033[K")
				fmt.Println("Test completed.")
			}
			return status.ReportID, nil
		case "failed":
			if !flagQuiet {
				fmt.Print("\r\033[K")
			}
			return status.ReportID, nil
		}

		if !flagQuiet {
			fmt.Printf("\r\033[K%s Testing... (%s)", spinner[i%len(spinner)], state)
			i++
		}

		time.Sleep(3 * time.Second)
	}

	return "", fmt.Errorf("test timed out after %s. Check status with: kuality status %s", timeout, testID)
}

func printReport(r *client.Report) {
	fmt.Println()
	fmt.Printf("  Target:     %s\n", r.Target)
	fmt.Printf("  Test type:  %s\n", r.TypeOfTest)
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

func isValidTestType(t string) bool {
	for _, v := range validTestTypes {
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
