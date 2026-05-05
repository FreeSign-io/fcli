package api

import (
	"context"
	"fmt"
)

// AddField adds a single field (signature/text/etc.) to a document.
func (c *Client) AddField(ctx context.Context, documentID int, req CreateFieldRequest) (*Field, error) {
	var out struct {
		Fields Field `json:"fields"`
	}
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/v1/documents/%d/fields", documentID), req, &out); err != nil {
		return nil, err
	}
	return &out.Fields, nil
}

// AddFields adds multiple fields in one request.
func (c *Client) AddFields(ctx context.Context, documentID int, reqs []CreateFieldRequest) ([]Field, error) {
	var out struct {
		Fields []Field `json:"fields"`
	}
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/v1/documents/%d/fields", documentID), reqs, &out); err != nil {
		return nil, err
	}
	return out.Fields, nil
}

// UpdateField patches a field.
func (c *Client) UpdateField(ctx context.Context, documentID, fieldID int, req CreateFieldRequest) (*Field, error) {
	var out Field
	if err := c.do(ctx, "PATCH", fmt.Sprintf("/api/v1/documents/%d/fields/%d", documentID, fieldID), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteField removes a field.
func (c *Client) DeleteField(ctx context.Context, documentID, fieldID int) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/api/v1/documents/%d/fields/%d", documentID, fieldID), nil, nil)
}
