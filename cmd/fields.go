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
	fieldRecipientID int
	fieldType        string
	fieldPage        int
	fieldX           float64
	fieldY           float64
	fieldWidth       float64
	fieldHeight      float64
)

func init() {
	cmdGroup := &cobra.Command{
		Use:   "fields",
		Short: "Manage signature/text/date fields on a document",
	}

	addCmd := &cobra.Command{
		Use:   "add <doc-id>",
		Short: "Add a field at given page coordinates (percentage units)",
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
			req := api.CreateFieldRequest{
				RecipientID: fieldRecipientID,
				Type:        api.FieldType(fieldType),
				PageNumber:  fieldPage,
				PageX:       fieldX,
				PageY:       fieldY,
				PageWidth:   fieldWidth,
				PageHeight:  fieldHeight,
			}
			f, err := client.AddField(ctx(cmd), id, req)
			if err != nil {
				return err
			}
			fmt.Printf("✓ Field %d (%s) added on page %d at %.1f%%,%.1f%%\n", f.ID, f.Type, f.PageNumber, f.PageX, f.PageY)
			return nil
		},
	}
	addCmd.Flags().IntVar(&fieldRecipientID, "recipient", 0, "recipient ID this field belongs to (required)")
	addCmd.Flags().StringVar(&fieldType, "type", "SIGNATURE", "field type: SIGNATURE, TEXT, DATE, EMAIL, NAME, NUMBER, RADIO, CHECKBOX, DROPDOWN, INITIALS")
	addCmd.Flags().IntVar(&fieldPage, "page", 1, "1-indexed page number")
	addCmd.Flags().Float64Var(&fieldX, "x", 0, "x position (% from left)")
	addCmd.Flags().Float64Var(&fieldY, "y", 0, "y position (% from top)")
	addCmd.Flags().Float64Var(&fieldWidth, "width", 15, "width (% of page)")
	addCmd.Flags().Float64Var(&fieldHeight, "height", 8, "height (% of page)")
	_ = addCmd.MarkFlagRequired("recipient")

	listCmd := &cobra.Command{
		Use:   "list <doc-id>",
		Short: "List fields on a document",
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
			doc, err := client.GetDocument(ctx(cmd), id)
			if err != nil {
				return err
			}
			return output.Render(os.Stdout, outputFormat(), doc.Fields,
				[]string{"ID", "Type", "Page", "X%", "Y%", "Recipient"},
				func() [][]string {
					rows := make([][]string, 0, len(doc.Fields))
					for _, f := range doc.Fields {
						rows = append(rows, []string{
							strconv.Itoa(f.ID),
							string(f.Type),
							strconv.Itoa(f.PageNumber),
							fmt.Sprintf("%.1f", f.PageX),
							fmt.Sprintf("%.1f", f.PageY),
							strconv.Itoa(f.RecipientID),
						})
					}
					return rows
				})
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove <doc-id> <field-id>",
		Short: "Remove a field",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			docID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			fid, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			client, _, err := loadClient()
			if err != nil {
				return err
			}
			if err := client.DeleteField(ctx(cmd), docID, fid); err != nil {
				return err
			}
			fmt.Printf("✓ Field %d removed from doc %d\n", fid, docID)
			return nil
		},
	}

	cmdGroup.AddCommand(addCmd, listCmd, removeCmd)
	rootCmd.AddCommand(cmdGroup)
}
