// Package cmd wires up the cobra command tree for fcli.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/FreeSign-io/fcli/internal/api"
	"github.com/FreeSign-io/fcli/internal/config"
	"github.com/FreeSign-io/fcli/internal/keychain"
	"github.com/FreeSign-io/fcli/internal/output"
	"github.com/spf13/cobra"
)

// Global flags exposed via the root command.
type globalFlags struct {
	Profile string
	BaseURL string
	Output  string
	Verbose bool
}

var flags globalFlags

// rootCmd is the entry point for `fcli`.
var rootCmd = &cobra.Command{
	Use:           "fcli",
	Short:         "fcli — the FreeSign command-line tool",
	Long:          "fcli drives the FreeSign HTTP API from the terminal: send PDFs for signature, list and download envelopes, manage recipients/fields/templates/webhooks. https://freesign.io",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute is called by main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printError(err)
		os.Exit(exitCodeForError(err))
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flags.Profile, "profile", "p", "", "named profile to use (default: active profile from config)")
	rootCmd.PersistentFlags().StringVar(&flags.BaseURL, "base-url", "", "override the FreeSign instance URL (e.g. https://sign.example.com)")
	rootCmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", "", "output format: json, yaml, table (default: auto)")
	rootCmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "enable verbose HTTP logs")
}

// --- shared helpers used by command handlers ---

// loadClient builds an *api.Client from the active profile + keychain.
// Prefer this over instantiating Client directly inside handlers.
func loadClient() (*api.Client, *config.Profile, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	profileName := flags.Profile
	if profileName == "" {
		profileName = cfg.ActiveProfile
	}

	profile, ok := cfg.Profiles[profileName]
	if !ok {
		return nil, nil, fmt.Errorf("unknown profile %q (run `fcli config profiles list`)", profileName)
	}

	if flags.BaseURL != "" {
		profile.BaseURL = flags.BaseURL
	}
	if profile.BaseURL == "" {
		return nil, nil, errors.New("profile has no base URL; set it with `fcli config profiles add`")
	}

	token, err := keychain.Get(profileName)
	if err != nil && !errors.Is(err, keychain.ErrNotFound) {
		return nil, nil, fmt.Errorf("read keychain: %w", err)
	}

	client, err := api.New(api.Options{BaseURL: profile.BaseURL, Token: token})
	if err != nil {
		return nil, nil, err
	}
	return client, &profile, nil
}

// outputFormat parses --output into an output.Format.
func outputFormat() output.Format {
	switch strings.ToLower(flags.Output) {
	case "":
		return output.FormatAuto
	case "json":
		return output.FormatJSON
	case "yaml", "yml":
		return output.FormatYAML
	case "table":
		return output.FormatTable
	default:
		fmt.Fprintf(os.Stderr, "warning: unknown --output %q, falling back to auto\n", flags.Output)
		return output.FormatAuto
	}
}

// ctx returns a request context tied to the command's context (cobra gives us
// a parent that's cancelled on SIGINT).
func ctx(cmd *cobra.Command) context.Context {
	if cmd.Context() != nil {
		return cmd.Context()
	}
	return context.Background()
}

// printError formats an error for the user. Auth/notfound/rate-limit get
// shorter, friendlier messages than a raw stack trace.
func printError(err error) {
	switch {
	case api.IsAuthError(err):
		fmt.Fprintln(os.Stderr, "error: not authenticated. Run `fcli auth login`.")
	case api.IsNotFound(err):
		fmt.Fprintln(os.Stderr, "error: not found.")
	case api.IsRateLimited(err):
		fmt.Fprintln(os.Stderr, "error: rate-limited (100 req/min). Try again in a minute.")
	default:
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}

func exitCodeForError(err error) int {
	switch {
	case api.IsAuthError(err):
		return 3
	case api.IsNotFound(err):
		return 4
	case api.IsRateLimited(err):
		return 5
	default:
		return 1
	}
}
