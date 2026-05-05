package api

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

// CreateDocument prepares an envelope and returns a presigned upload URL plus
// recipient signing URLs. The PDF must be uploaded to the returned UploadURL
// (see PutBytes) before the document can be sent.
func (c *Client) CreateDocument(ctx context.Context, req CreateDocumentRequest) (*CreateDocumentResponse, error) {
	var out CreateDocumentResponse
	if err := c.do(ctx, "POST", "/api/v1/documents", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendDocument transitions an envelope from DRAFT to PENDING and dispatches
// signature request emails to all recipients.
func (c *Client) SendDocument(ctx context.Context, id int, req SendDocumentRequest) (*Document, error) {
	var out Document
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/v1/documents/%d/send", id), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDocument fetches one document by ID.
func (c *Client) GetDocument(ctx context.Context, id int) (*Document, error) {
	var out Document
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/v1/documents/%d", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListDocuments returns one page of documents. perPage defaults to 10 server-side.
func (c *Client) ListDocuments(ctx context.Context, page, perPage int) (*ListDocumentsResponse, error) {
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		q.Set("perPage", strconv.Itoa(perPage))
	}

	path := "/api/v1/documents"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out ListDocumentsResponse
	if err := c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteDocument permanently removes a document.
func (c *Client) DeleteDocument(ctx context.Context, id int) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/api/v1/documents/%d", id), nil, nil)
}

// ResendDocument re-emails the signature request to selected (or all) recipients.
func (c *Client) ResendDocument(ctx context.Context, id int, recipientIDs []int) error {
	body := map[string]any{}
	if len(recipientIDs) > 0 {
		body["recipients"] = recipientIDs
	}
	return c.do(ctx, "POST", fmt.Sprintf("/api/v1/documents/%d/resend", id), body, nil)
}

// DownloadDocument streams the signed PDF (or original if downloadOriginal=true)
// from /api/v1/documents/:id/download into w. Requires S3 storage.
func (c *Client) DownloadDocument(ctx context.Context, id int, downloadOriginal bool, w io.Writer) error {
	q := url.Values{}
	if downloadOriginal {
		q.Set("downloadOriginalDocument", "true")
	}
	path := fmt.Sprintf("/api/v1/documents/%d/download", id)
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	// First call returns a JSON envelope with a presigned download URL.
	var meta struct {
		DownloadURL string `json:"downloadUrl"`
	}
	if err := c.do(ctx, "GET", path, nil, &meta); err != nil {
		return err
	}
	if meta.DownloadURL == "" {
		return fmt.Errorf("server did not return a download URL")
	}
	return c.streamGet(ctx, meta.DownloadURL, w)
}

func (c *Client) streamGet(ctx context.Context, getURL string, w io.Writer) error {
	req, err := newRequest(ctx, "GET", getURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("download GET: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return &APIError{Status: resp.StatusCode, Message: fmt.Sprintf("download failed: HTTP %d", resp.StatusCode)}
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("stream body: %w", err)
	}
	return nil
}
