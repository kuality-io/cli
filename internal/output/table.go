package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func Table(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	sep := make([]string, len(headers))
	for i, h := range headers {
		sep[i] = strings.Repeat("-", len(h))
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))

	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	tw.Flush()
}

func SeverityColor(severity string, count int) string {
	if count == 0 {
		return fmt.Sprintf("%d", count)
	}
	switch strings.ToLower(severity) {
	case "high":
		return fmt.Sprintf("\033[31m%d\033[0m", count) // red
	case "medium":
		return fmt.Sprintf("\033[33m%d\033[0m", count) // yellow
	case "low":
		return fmt.Sprintf("\033[36m%d\033[0m", count) // cyan
	default:
		return fmt.Sprintf("%d", count)
	}
}

func StatusIcon(state string) string {
	switch strings.ToLower(state) {
	case "completed":
		return "\033[32m✓\033[0m" // green check
	case "failed":
		return "\033[31m✗\033[0m" // red x
	case "running":
		return "\033[33m⟳\033[0m" // yellow spinner
	case "pending":
		return "\033[90m◌\033[0m" // gray circle
	default:
		return "?"
	}
}
