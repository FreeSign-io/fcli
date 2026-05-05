package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/FreeSign-io/fcli/internal/api"
	"github.com/FreeSign-io/fcli/internal/config"
	"github.com/FreeSign-io/fcli/internal/keychain"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	authToken string
)

func init() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with a FreeSign instance",
	}

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Store an API token for the active profile",
		RunE:  runAuthLogin,
	}
	loginCmd.Flags().StringVar(&authToken, "token", "", "token to store (default: prompt interactively)")

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show which profile is active and whether a token is set",
		RunE:  runAuthStatus,
	}

	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Validate the stored token by calling the FreeSign API",
		RunE:  runAuthWhoami,
	}

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Forget the stored token for the active profile",
		RunE:  runAuthLogout,
	}

	authCmd.AddCommand(loginCmd, statusCmd, whoamiCmd, logoutCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	profileName := flags.Profile
	if profileName == "" {
		profileName = cfg.ActiveProfile
	}
	profile, ok := cfg.Profiles[profileName]
	if !ok {
		return fmt.Errorf("unknown profile %q", profileName)
	}

	token := strings.TrimSpace(authToken)
	if token == "" {
		fmt.Fprintf(os.Stderr, "Paste your FreeSign API token for %s (input hidden): ", profile.BaseURL)
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			// Non-TTY: read a line plain (e.g. piped from a secrets manager).
			scan := bufio.NewScanner(os.Stdin)
			if scan.Scan() {
				token = strings.TrimSpace(scan.Text())
			}
		} else {
			b, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return fmt.Errorf("read token: %w", err)
			}
			token = strings.TrimSpace(string(b))
		}
	}
	if token == "" {
		return errors.New("token is empty")
	}

	// Verify against the API before persisting (cheap GET).
	if err := verifyToken(cmd, profile.BaseURL, token); err != nil {
		return fmt.Errorf("token rejected: %w", err)
	}

	if err := keychain.Set(profileName, token); err != nil {
		return err
	}

	fmt.Printf("✓ Authenticated to %s (profile: %s)\n", profile.BaseURL, profileName)
	return nil
}

func verifyToken(cmd *cobra.Command, baseURL, token string) error {
	client, err := api.New(api.Options{BaseURL: baseURL, Token: token})
	if err != nil {
		return err
	}
	// Hit a cheap, auth-required endpoint.
	if _, err := client.ListDocuments(ctx(cmd), 1, 1); err != nil {
		return err
	}
	return nil
}

func runAuthStatus(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	fmt.Printf("Active profile: %s\n", cfg.ActiveProfile)
	for name, p := range cfg.Profiles {
		token, _ := keychain.Get(name)
		state := "no token"
		if token != "" {
			state = mask(token)
		}
		marker := " "
		if name == cfg.ActiveProfile {
			marker = "*"
		}
		fmt.Printf(" %s %-12s  %-30s  %s\n", marker, name, p.BaseURL, state)
	}
	return nil
}

func runAuthWhoami(cmd *cobra.Command, _ []string) error {
	client, profile, err := loadClient()
	if err != nil {
		return err
	}
	if _, err := client.ListDocuments(ctx(cmd), 1, 1); err != nil {
		return err
	}
	fmt.Printf("✓ Authenticated to %s (profile: %s)\n", profile.BaseURL, profile.Name)
	return nil
}

func runAuthLogout(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	profileName := flags.Profile
	if profileName == "" {
		profileName = cfg.ActiveProfile
	}
	if err := keychain.Delete(profileName); err != nil {
		return err
	}
	fmt.Printf("✓ Removed token for profile %q\n", profileName)
	return nil
}

func mask(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "…" + s[len(s)-4:]
}
