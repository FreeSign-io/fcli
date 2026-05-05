package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/FreeSign-io/fcli/internal/api"
	"github.com/FreeSign-io/fcli/internal/output"
	"github.com/spf13/cobra"
)

var (
	recipName  string
	recipEmail string
	recipRole  string
	recipOrder int
)

func init() {
	cmdGroup := &cobra.Command{
		Use:   "recipients",
		Short: "Manage document recipients",
	}

	addCmd := &cobra.Command{
		Use:   "add <doc-id>",
		Short: "Add a recipient to a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id: %w", err)
			}
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			role := api.RecipientRoleSigner
			if recipRole != "" {
				role = api.RecipientRole(recipRole)
			}
			req := api.CreateRecipientRequest{Name: recipName, Email: recipEmail, Role: role}
			if recipOrder > 0 {
				req.SigningOrder = &recipOrder
			}
			r, err := client.AddRecipient(ctx(cmd), id, req)
			if err != nil {
				return err
			}
			fmt.Printf("✓ Recipient %d added: %s <%s>\n  %s\n", r.ID, r.Name, r.Email, r.SigningURL)
			return nil
		},
	}
	addCmd.Flags().StringVar(&recipName, "name", "", "recipient name (required)")
	addCmd.Flags().StringVar(&recipEmail, "email", "", "recipient email (required)")
	addCmd.Flags().StringVar(&recipRole, "role", "SIGNER", "role: SIGNER, APPROVER, VIEWER, CC")
	addCmd.Flags().IntVar(&recipOrder, "order", 0, "signing order (0 = parallel)")
	_ = addCmd.MarkFlagRequired("name")
	_ = addCmd.MarkFlagRequired("email")

	listCmd := &cobra.Command{
		Use:   "list <doc-id>",
		Short: "List recipients on a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id: %w", err)
			}
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			doc, err := client.GetDocument(ctx(cmd), id)
			if err != nil {
				return err
			}
			return output.Render(os.Stdout, outputFormat(), doc.Recipients,
				[]string{"ID", "Name", "Email", "Role", "Status"},
				func() [][]string {
					rows := make([][]string, 0, len(doc.Recipients))
					for _, r := range doc.Recipients {
						rows = append(rows, []string{
							strconv.Itoa(r.ID), r.Name, r.Email, string(r.Role), string(r.SigningStatus),
						})
					}
					return rows
				})
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update <doc-id> <recipient-id>",
		Short: "Update a recipient",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			docID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			rid, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			req := api.CreateRecipientRequest{Name: recipName, Email: recipEmail, Role: api.RecipientRole(recipRole)}
			if recipOrder > 0 {
				req.SigningOrder = &recipOrder
			}
			r, err := client.UpdateRecipient(ctx(cmd), docID, rid, req)
			if err != nil {
				return err
			}
			fmt.Printf("✓ Recipient %d updated: %s <%s>\n", r.ID, r.Name, r.Email)
			return nil
		},
	}
	updateCmd.Flags().StringVar(&recipName, "name", "", "new name")
	updateCmd.Flags().StringVar(&recipEmail, "email", "", "new email")
	updateCmd.Flags().StringVar(&recipRole, "role", "", "new role")
	updateCmd.Flags().IntVar(&recipOrder, "order", 0, "new signing order")

	removeCmd := &cobra.Command{
		Use:   "remove <doc-id> <recipient-id>",
		Short: "Remove a recipient",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			docID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			rid, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			if err := client.DeleteRecipient(ctx(cmd), docID, rid); err != nil {
				return err
			}
			fmt.Printf("✓ Recipient %d removed from doc %d\n", rid, docID)
			return nil
		},
	}

	cmdGroup.AddCommand(addCmd, listCmd, updateCmd, removeCmd)
	rootCmd.AddCommand(cmdGroup)
}
