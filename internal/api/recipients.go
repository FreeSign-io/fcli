package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// AddRecipient appends a recipient to an existing document.
func (c *Client) AddRecipient(ctx context.Context, documentID int, req CreateRecipientRequest) (*Recipient, error) {
	var out Recipient
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/v1/documents/%d/recipients", documentID), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateRecipient changes properties of a recipient (email, role, etc.).
func (c *Client) UpdateRecipient(ctx context.Context, documentID, recipientID int, req CreateRecipientRequest) (*Recipient, error) {
	var out Recipient
	if err := c.do(ctx, "PATCH", fmt.Sprintf("/api/v1/documents/%d/recipients/%d", documentID, recipientID), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteRecipient removes a recipient.
func (c *Client) DeleteRecipient(ctx context.Context, documentID, recipientID int) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/api/v1/documents/%d/recipients/%d", documentID, recipientID), nil, nil)
}

// helper used by streamGet in documents.go (factored here to keep imports tidy)
func newRequest(ctx context.Context, method, urlStr string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, urlStr, body)
}
