package cmd

import (
	"fmt"

	"github.com/FreeSign-io/fcli/internal/config"
	"github.com/FreeSign-io/fcli/internal/keychain"
	"github.com/spf13/cobra"
)

var profileBaseURL string

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage fcli configuration",
	}

	profilesCmd := &cobra.Command{
		Use:   "profiles",
		Short: "Manage named FreeSign profiles (one per instance)",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configured profiles",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			fmt.Printf("Active profile: %s\n", cfg.ActiveProfile)
			for name, p := range cfg.Profiles {
				marker := " "
				if name == cfg.ActiveProfile {
					marker = "*"
				}
				fmt.Printf(" %s %-12s %s\n", marker, name, p.BaseURL)
			}
			return nil
		},
	}

	addCmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new profile pointing at a FreeSign instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			name := args[0]
			if profileBaseURL == "" {
				return fmt.Errorf("--base-url is required")
			}
			cfg.Profiles[name] = config.Profile{Name: name, BaseURL: profileBaseURL}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("✓ Added profile %q -> %s\n", name, profileBaseURL)
			return nil
		},
	}
	addCmd.Flags().StringVar(&profileBaseURL, "base-url", "", "base URL of the FreeSign instance (required)")

	useCmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Set the active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			name := args[0]
			if _, ok := cfg.Profiles[name]; !ok {
				return fmt.Errorf("unknown profile %q", name)
			}
			cfg.ActiveProfile = name
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("✓ Active profile set to %q\n", name)
			return nil
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a profile (and its stored token)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			name := args[0]
			if name == "default" {
				return fmt.Errorf("cannot delete the default profile")
			}
			delete(cfg.Profiles, name)
			if cfg.ActiveProfile == name {
				cfg.ActiveProfile = "default"
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			_ = keychain.Delete(name)
			fmt.Printf("✓ Removed profile %q\n", name)
			return nil
		},
	}

	profilesCmd.AddCommand(listCmd, addCmd, useCmd, deleteCmd)
	configCmd.AddCommand(profilesCmd)
	rootCmd.AddCommand(configCmd)
}
