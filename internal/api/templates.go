package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListTemplates returns templates for the active team/user.
func (c *Client) ListTemplates(ctx context.Context, page, perPage int) (*ListTemplatesResponse, error) {
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		q.Set("perPage", strconv.Itoa(perPage))
	}
	path := "/api/v1/templates"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out ListTemplatesResponse
	if err := c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTemplate returns one template.
func (c *Client) GetTemplate(ctx context.Context, id int) (*Template, error) {
	var out Template
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/v1/templates/%d", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTemplate removes a template.
func (c *Client) DeleteTemplate(ctx context.Context, id int) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/api/v1/templates/%d", id), nil, nil)
}

// GenerateDocumentFromTemplate creates a new document seeded from a template.
// `recipients` maps the template's recipient IDs to actual {name,email} pairs.
func (c *Client) GenerateDocumentFromTemplate(ctx context.Context, templateID int, body any) (*Document, error) {
	var out Document
	path := fmt.Sprintf("/api/v1/templates/%d/generate-document", templateID)
	if err := c.do(ctx, "POST", path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
