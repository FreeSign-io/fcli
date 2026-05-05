package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/FreeSign-io/fcli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	cmdGroup := &cobra.Command{
		Use:   "templates",
		Short: "Manage document templates",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List templates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			resp, err := client.ListTemplates(ctx(cmd), 1, 50)
			if err != nil {
				return err
			}
			return output.Render(os.Stdout, outputFormat(), resp.Templates,
				[]string{"ID", "Title", "Recipients", "Fields"},
				func() [][]string {
					rows := make([][]string, 0, len(resp.Templates))
					for _, t := range resp.Templates {
						rows = append(rows, []string{
							strconv.Itoa(t.ID), t.Title,
							strconv.Itoa(len(t.Recipients)), strconv.Itoa(len(t.Fields)),
						})
					}
					return rows
				})
		},
	}

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Show full detail for one template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			t, err := client.GetTemplate(ctx(cmd), id)
			if err != nil {
				return err
			}
			return output.Render(os.Stdout, outputFormat(), t, nil, nil)
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			if err := client.DeleteTemplate(ctx(cmd), id); err != nil {
				return err
			}
			fmt.Printf("✓ Template %d deleted\n", id)
			return nil
		},
	}

	cmdGroup.AddCommand(listCmd, getCmd, deleteCmd)
	rootCmd.AddCommand(cmdGroup)
}
