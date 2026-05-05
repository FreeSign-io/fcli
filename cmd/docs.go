package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/FreeSign-io/fcli/internal/api"
	"github.com/FreeSign-io/fcli/internal/output"
	"github.com/spf13/cobra"
)

var (
	sendTo                string
	sendSubject           string
	sendMessage           string
	sendSignatureAnchor   string
	sendImmediately       bool
	sendExternalID        string

	listStatus  string
	listAll     bool
	listPerPage int

	downloadOut      string
	downloadOriginal bool

	resendRecipient string
)

func init() {
	docsCmd := &cobra.Command{
		Use:   "docs",
		Short: "Manage documents (envelopes)",
	}

	sendCmd := &cobra.Command{
		Use:   "send <pdf>",
		Short: "Upload a PDF, attach recipients, optionally send for signature",
		Args:  cobra.ExactArgs(1),
		RunE:  runDocsSend,
	}
	sendCmd.Flags().StringVar(&sendTo, "to", "", `recipients in "Name:email,Name:email" form (required)`)
	sendCmd.Flags().StringVar(&sendSubject, "subject", "", "email subject for signature requests")
	sendCmd.Flags().StringVar(&sendMessage, "message", "", "email body for signature requests")
	sendCmd.Flags().StringVar(&sendSignatureAnchor, "signature-anchor", "", "place a signature field at the first occurrence of this text in the PDF (optional)")
	sendCmd.Flags().BoolVar(&sendImmediately, "send", false, "transition envelope from DRAFT to PENDING and dispatch emails (default: leave as DRAFT)")
	sendCmd.Flags().StringVar(&sendExternalID, "external-id", "", "idempotency key (e.g. your internal contract ID)")
	_ = sendCmd.MarkFlagRequired("to")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List documents",
		RunE:  runDocsList,
	}
	listCmd.Flags().StringVar(&listStatus, "status", "", "filter by status: draft|pending|completed|rejected")
	listCmd.Flags().BoolVar(&listAll, "all", false, "auto-paginate and return every document (slow for large accounts)")
	listCmd.Flags().IntVar(&listPerPage, "per-page", 20, "page size when not using --all")

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Show full detail for one document",
		Args:  cobra.ExactArgs(1),
		RunE:  runDocsGet,
	}

	statusCmd := &cobra.Command{
		Use:   "status <id>",
		Short: "Print a one-line summary of a document's signing progress",
		Args:  cobra.ExactArgs(1),
		RunE:  runDocsStatus,
	}

	downloadCmd := &cobra.Command{
		Use:   "download <id>",
		Short: "Download the signed PDF",
		Args:  cobra.ExactArgs(1),
		RunE:  runDocsDownload,
	}
	downloadCmd.Flags().StringVar(&downloadOut, "out", "", "output path (default: ./<title>.pdf)")
	downloadCmd.Flags().BoolVar(&downloadOriginal, "original", false, "download the unsigned original instead of the signed PDF")

	cancelCmd := &cobra.Command{
		Use:   "cancel <id>",
		Short: "Cancel a pending envelope (deletes it server-side)",
		Args:  cobra.ExactArgs(1),
		RunE:  runDocsCancel,
	}

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a document permanently",
		Args:  cobra.ExactArgs(1),
		RunE:  runDocsDelete,
	}

	resendCmd := &cobra.Command{
		Use:   "resend <id>",
		Short: "Resend the signature request email",
		Args:  cobra.ExactArgs(1),
		RunE:  runDocsResend,
	}
	resendCmd.Flags().StringVar(&resendRecipient, "recipient", "", "only resend to the recipient with this email (default: all unsigned)")

	docsCmd.AddCommand(sendCmd, listCmd, getCmd, statusCmd, downloadCmd, cancelCmd, deleteCmd, resendCmd)
	rootCmd.AddCommand(docsCmd)
}

// --- send ---

func runDocsSend(cmd *cobra.Command, args []string) error {
	pdfPath := args[0]
	pdfFile, err := os.Open(pdfPath)
	if err != nil {
		return fmt.Errorf("open pdf: %w", err)
	}
	defer pdfFile.Close()

	stat, err := pdfFile.Stat()
	if err != nil {
		return fmt.Errorf("stat pdf: %w", err)
	}

	recipients, err := parseRecipientList(sendTo)
	if err != nil {
		return err
	}

	client, _, err := loadClient()
	if err != nil {
		return err
	}

	title := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))

	createReq := api.CreateDocumentRequest{
		Title:      title,
		ExternalID: sendExternalID,
		Recipients: recipients,
	}
	if sendSubject != "" || sendMessage != "" {
		createReq.Meta = &api.DocumentMeta{Subject: sendSubject, Message: sendMessage}
	}

	fmt.Fprintf(os.Stderr, "→ Creating envelope...\n")
	created, err := client.CreateDocument(ctx(cmd), createReq)
	if err != nil {
		return fmt.Errorf("create document: %w", err)
	}

	fmt.Fprintf(os.Stderr, "→ Uploading %s (%s)...\n", filepath.Base(pdfPath), humanSize(stat.Size()))
	if err := client.PutBytes(ctx(cmd), created.UploadURL, "application/octet-stream", pdfFile, stat.Size()); err != nil {
		return fmt.Errorf("upload: %w (envelope %d kept as DRAFT — `fcli docs delete %d` to clean up)", err, created.DocumentID, created.DocumentID)
	}

	if sendSignatureAnchor != "" {
		fmt.Fprintf(os.Stderr, "→ Anchor placement is not implemented yet; signers will sign without pre-placed fields.\n")
		// Phase 2: integrate pdfcpu / unipdf anchor lookup.
	}

	if sendImmediately {
		fmt.Fprintf(os.Stderr, "→ Sending for signature...\n")
		t := true
		if _, err := client.SendDocument(ctx(cmd), created.DocumentID, api.SendDocumentRequest{SendEmail: &t}); err != nil {
			return fmt.Errorf("send: %w (envelope %d kept as DRAFT)", err, created.DocumentID)
		}
	}

	fmt.Printf("\n✓ Envelope %d %s\n", created.DocumentID, ifThen(sendImmediately, "sent", "created (DRAFT)"))
	fmt.Printf("  Title: %s\n", title)
	for _, r := range created.Recipients {
		fmt.Printf("  Signer: %s <%s>\n    %s\n", r.Name, r.Email, r.SigningURL)
	}
	return nil
}

// --- list ---

func runDocsList(cmd *cobra.Command, _ []string) error {
	client, _, err := loadClient()
	if err != nil {
		return err
	}

	wanted := strings.ToUpper(listStatus)

	page := 1
	perPage := listPerPage
	all := []api.Document{}

	for {
		resp, err := client.ListDocuments(ctx(cmd), page, perPage)
		if err != nil {
			return err
		}
		for _, d := range resp.Data {
			if wanted == "" || string(d.Status) == wanted {
				all = append(all, d)
			}
		}
		if !listAll || page >= resp.TotalPages {
			break
		}
		page++
	}

	return output.Render(os.Stdout, outputFormat(), all,
		[]string{"ID", "Title", "Status", "Recipients", "Created"},
		func() [][]string {
			rows := make([][]string, 0, len(all))
			for _, d := range all {
				rows = append(rows, []string{
					strconv.Itoa(d.ID),
					d.Title,
					string(d.Status),
					strconv.Itoa(len(d.Recipients)),
					d.CreatedAt.Format(time.RFC3339),
				})
			}
			return rows
		})
}

// --- get / status ---

func runDocsGet(cmd *cobra.Command, args []string) error {
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
	return output.Render(os.Stdout, outputFormat(), doc, nil, nil)
}

func runDocsStatus(cmd *cobra.Command, args []string) error {
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

	signed := 0
	for _, r := range doc.Recipients {
		if r.SigningStatus == api.SigningStatusSigned {
			signed++
		}
	}
	fmt.Printf("[%s] %s — %d/%d signed\n", doc.Status, doc.Title, signed, len(doc.Recipients))
	return nil
}

// --- download ---

func runDocsDownload(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	client, _, err := loadClient()
	if err != nil {
		return err
	}

	out := downloadOut
	if out == "" {
		doc, err := client.GetDocument(ctx(cmd), id)
		if err != nil {
			return err
		}
		out = sanitizeFilename(doc.Title) + ".pdf"
	}

	var w io.Writer
	if out == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(out)
		if err != nil {
			return fmt.Errorf("create %s: %w", out, err)
		}
		defer f.Close()
		w = f
	}

	if err := client.DownloadDocument(ctx(cmd), id, downloadOriginal, w); err != nil {
		return err
	}
	if out != "-" {
		fmt.Fprintf(os.Stderr, "✓ Saved %s\n", out)
	}
	return nil
}

// --- cancel / delete / resend ---

func runDocsCancel(cmd *cobra.Command, args []string) error { return runDocsDelete(cmd, args) }

func runDocsDelete(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	client, _, err := loadClient()
	if err != nil {
		return err
	}
	if err := client.DeleteDocument(ctx(cmd), id); err != nil {
		return err
	}
	fmt.Printf("✓ Document %d deleted\n", id)
	return nil
}

func runDocsResend(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	client, _, err := loadClient()
	if err != nil {
		return err
	}

	var rids []int
	if resendRecipient != "" {
		doc, err := client.GetDocument(ctx(cmd), id)
		if err != nil {
			return err
		}
		for _, r := range doc.Recipients {
			if strings.EqualFold(r.Email, resendRecipient) {
				rids = append(rids, r.ID)
			}
		}
		if len(rids) == 0 {
			return fmt.Errorf("no recipient with email %q on document %d", resendRecipient, id)
		}
	}

	if err := client.ResendDocument(ctx(cmd), id, rids); err != nil {
		return err
	}
	fmt.Printf("✓ Resent document %d\n", id)
	return nil
}

// --- helpers ---

// parseRecipientList parses "Name:email,Name:email" into structs.
func parseRecipientList(s string) ([]api.CreateRecipientInline, error) {
	if strings.TrimSpace(s) == "" {
		return nil, fmt.Errorf("--to is empty")
	}
	out := []api.CreateRecipientInline{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		colon := strings.LastIndex(part, ":")
		if colon < 1 {
			return nil, fmt.Errorf(`recipient %q: expected "Name:email"`, part)
		}
		name := strings.TrimSpace(part[:colon])
		email := strings.TrimSpace(part[colon+1:])
		if name == "" || email == "" || !strings.Contains(email, "@") {
			return nil, fmt.Errorf(`recipient %q: name or email empty/invalid`, part)
		}
		out = append(out, api.CreateRecipientInline{Name: name, Email: email, Role: api.RecipientRoleSigner})
	}
	return out, nil
}

func sanitizeFilename(s string) string {
	repl := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "\x00", "_")
	return strings.TrimSpace(repl.Replace(s))
}

func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func ifThen[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
