package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kuality-io/cli/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage API key authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store your API key",
	Long: `Authenticate with your Kuality API key.

Get your key from https://kuality.io/settings/api-keys

The key is stored in ~/.kuality/config.yaml with restricted
file permissions (0600). You can also set the KUALITY_API_KEY
environment variable instead.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Enter your API key: ")
		reader := bufio.NewReader(os.Stdin)
		key, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("cannot read input: %w", err)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		if !strings.HasPrefix(key, "ku_") {
			return fmt.Errorf("invalid API key format. Keys start with 'ku_'")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cfg.APIKey = key
		if err := config.Save(cfg); err != nil {
			return err
		}

		dir, _ := config.Dir()
		fmt.Printf("API key saved to %s/config.yaml\n", dir)
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.APIKey == "" {
			fmt.Println("Not authenticated. Run 'kuality auth login' or set KUALITY_API_KEY.")
			return nil
		}

		source := "config file"
		if os.Getenv("KUALITY_API_KEY") != "" {
			source = "KUALITY_API_KEY env var"
		}
		if flagAPIKey != "" {
			source = "--api-key flag"
		}

		prefix := cfg.APIKey
		if len(prefix) > 10 {
			prefix = prefix[:10] + "..."
		}
		fmt.Printf("Authenticated via %s (key: %s)\n", source, prefix)
		fmt.Printf("API endpoint: %s\n", cfg.BaseURL)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cfg.APIKey = ""
		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println("API key removed.")

		if os.Getenv("KUALITY_API_KEY") != "" {
			fmt.Println("Note: KUALITY_API_KEY environment variable is still set.")
		}

		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
