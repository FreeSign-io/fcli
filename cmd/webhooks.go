package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/FreeSign-io/fcli/internal/api"
	"github.com/FreeSign-io/fcli/internal/output"
	"github.com/spf13/cobra"
)

var (
	webhookURL    string
	webhookEvents string
)

func init() {
	cmdGroup := &cobra.Command{
		Use:   "webhooks",
		Short: "Manage webhook subscriptions",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			hooks, err := client.ListWebhooks(ctx(cmd))
			if err != nil {
				return err
			}
			return output.Render(os.Stdout, outputFormat(), hooks,
				[]string{"ID", "URL", "Events", "Enabled"},
				func() [][]string {
					rows := make([][]string, 0, len(hooks))
					for _, h := range hooks {
						rows = append(rows, []string{
							h.ID, h.WebhookURL, strings.Join(h.EventTriggers, ","), boolStr(h.Enabled),
						})
					}
					return rows
				})
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Register a new webhook",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			if webhookURL == "" || webhookEvents == "" {
				return fmt.Errorf("--url and --events are required")
			}
			req := api.CreateWebhookRequest{
				WebhookURL:    webhookURL,
				EventTriggers: strings.Split(webhookEvents, ","),
				Enabled:       true,
			}
			h, err := client.CreateWebhook(ctx(cmd), req)
			if err != nil {
				return err
			}
			fmt.Printf("✓ Webhook %s created\n  Secret: %s\n", h.ID, h.Secret)
			return nil
		},
	}
	createCmd.Flags().StringVar(&webhookURL, "url", "", "endpoint URL (required)")
	createCmd.Flags().StringVar(&webhookEvents, "events", "", "comma-separated events e.g. DOCUMENT_SENT,DOCUMENT_COMPLETED (required)")

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a webhook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			if err := client.DeleteWebhook(ctx(cmd), args[0]); err != nil {
				return err
			}
			fmt.Printf("✓ Webhook %s deleted\n", args[0])
			return nil
		},
	}

	cmdGroup.AddCommand(listCmd, createCmd, deleteCmd)
	rootCmd.AddCommand(cmdGroup)
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
