package api

import (
	"context"
	"fmt"
)

// CreateWebhookRequest is the body for POST /api/v1/webhooks.
type CreateWebhookRequest struct {
	WebhookURL    string   `json:"webhookUrl"`
	EventTriggers []string `json:"eventTriggers"`
	Enabled       bool     `json:"enabled"`
}

// ListWebhooks returns all webhooks the caller can see.
func (c *Client) ListWebhooks(ctx context.Context) ([]Webhook, error) {
	var out []Webhook
	if err := c.do(ctx, "GET", "/api/v1/webhooks", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateWebhook registers a new webhook.
func (c *Client) CreateWebhook(ctx context.Context, req CreateWebhookRequest) (*Webhook, error) {
	var out Webhook
	if err := c.do(ctx, "POST", "/api/v1/webhooks", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteWebhook removes a webhook by ID.
func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/api/v1/webhooks/%s", id), nil, nil)
}
