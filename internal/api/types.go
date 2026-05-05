// Package api is a hand-written Go client for the FreeSign HTTP API (/api/v1).
//
// Schemas mirror packages/api/v1/schema.ts in the FreeSign repo. Only fields
// fcli actually uses are typed; the rest are accepted as RawMessage so a
// future server change doesn't break the client.
package api

import (
	"encoding/json"
	"time"
)

// --- enums (mirror Prisma enums on the server) ---

type DocumentStatus string

const (
	DocumentStatusDraft     DocumentStatus = "DRAFT"
	DocumentStatusPending   DocumentStatus = "PENDING"
	DocumentStatusCompleted DocumentStatus = "COMPLETED"
	DocumentStatusRejected  DocumentStatus = "REJECTED"
)

type RecipientRole string

const (
	RecipientRoleSigner   RecipientRole = "SIGNER"
	RecipientRoleApprover RecipientRole = "APPROVER"
	RecipientRoleViewer   RecipientRole = "VIEWER"
	RecipientRoleCC       RecipientRole = "CC"
)

type SigningStatus string

const (
	SigningStatusPending  SigningStatus = "NOT_SIGNED"
	SigningStatusSigned   SigningStatus = "SIGNED"
	SigningStatusRejected SigningStatus = "REJECTED"
)

type FieldType string

const (
	FieldTypeSignature FieldType = "SIGNATURE"
	FieldTypeText      FieldType = "TEXT"
	FieldTypeDate      FieldType = "DATE"
	FieldTypeEmail     FieldType = "EMAIL"
	FieldTypeName      FieldType = "NAME"
	FieldTypeNumber    FieldType = "NUMBER"
	FieldTypeRadio     FieldType = "RADIO"
	FieldTypeCheckbox  FieldType = "CHECKBOX"
	FieldTypeDropdown  FieldType = "DROPDOWN"
	FieldTypeInitials  FieldType = "INITIALS"
)

// --- domain types ---

type Recipient struct {
	ID            int           `json:"id"`
	DocumentID    int           `json:"documentId,omitempty"`
	Name          string        `json:"name"`
	Email         string        `json:"email"`
	Role          RecipientRole `json:"role"`
	SigningOrder  *int          `json:"signingOrder,omitempty"`
	Token         string        `json:"token,omitempty"`
	SigningURL    string        `json:"signingUrl,omitempty"`
	SigningStatus SigningStatus `json:"signingStatus,omitempty"`
	SignedAt      *time.Time    `json:"signedAt,omitempty"`
	ReadStatus    string        `json:"readStatus,omitempty"`
	SendStatus    string        `json:"sendStatus,omitempty"`
}

type Field struct {
	ID          int             `json:"id"`
	DocumentID  int             `json:"documentId,omitempty"`
	RecipientID int             `json:"recipientId"`
	Type        FieldType       `json:"type"`
	PageNumber  int             `json:"pageNumber"`
	PageX       float64         `json:"pageX"`
	PageY       float64         `json:"pageY"`
	PageWidth   float64         `json:"pageWidth"`
	PageHeight  float64         `json:"pageHeight"`
	FieldMeta   json.RawMessage `json:"fieldMeta,omitempty"`
}

type DocumentMeta struct {
	Subject     string `json:"subject,omitempty"`
	Message     string `json:"message,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	DateFormat  string `json:"dateFormat,omitempty"`
	RedirectURL string `json:"redirectUrl,omitempty"`
	Language    string `json:"language,omitempty"`
}

type Document struct {
	ID          int             `json:"id"`
	ExternalID  string          `json:"externalId,omitempty"`
	Title       string          `json:"title"`
	Status      DocumentStatus  `json:"status"`
	UserID      int             `json:"userId,omitempty"`
	TeamID      *int            `json:"teamId,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt,omitempty"`
	CompletedAt *time.Time      `json:"completedAt,omitempty"`
	DeletedAt   *time.Time      `json:"deletedAt,omitempty"`
	Recipients  []Recipient     `json:"recipients,omitempty"`
	Fields      []Field         `json:"fields,omitempty"`
	Meta        json.RawMessage `json:"documentMeta,omitempty"`
}

// CreateDocumentRequest is the body for POST /api/v1/documents.
type CreateDocumentRequest struct {
	Title      string                  `json:"title"`
	ExternalID string                  `json:"externalId,omitempty"`
	Recipients []CreateRecipientInline `json:"recipients"`
	Meta       *DocumentMeta           `json:"meta,omitempty"`
	FormValues json.RawMessage         `json:"formValues,omitempty"`
}

type CreateRecipientInline struct {
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	Role         RecipientRole `json:"role,omitempty"`
	SigningOrder *int          `json:"signingOrder,omitempty"`
}

// CreateDocumentResponse comes back from POST /api/v1/documents.
type CreateDocumentResponse struct {
	UploadURL  string             `json:"uploadUrl"`
	DocumentID int                `json:"documentId"`
	Recipients []SigningRecipient `json:"recipients"`
}

type SigningRecipient struct {
	RecipientID int    `json:"recipientId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Token       string `json:"token"`
	SigningURL  string `json:"signingUrl"`
	Role        string `json:"role"`
}

// SendDocumentRequest is the body for POST /api/v1/documents/:id/send.
type SendDocumentRequest struct {
	SendEmail            *bool `json:"sendEmail,omitempty"`
	SendCompletionEmails *bool `json:"sendCompletionEmails,omitempty"`
}

// ListDocumentsResponse is returned from GET /api/v1/documents.
type ListDocumentsResponse struct {
	Data       []Document `json:"data"`
	TotalPages int        `json:"totalPages"`
	Page       int        `json:"page,omitempty"`
	PerPage    int        `json:"perPage,omitempty"`
}

// CreateRecipientRequest is the body for POST /api/v1/documents/:id/recipients.
type CreateRecipientRequest struct {
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	Role         RecipientRole `json:"role,omitempty"`
	SigningOrder *int          `json:"signingOrder,omitempty"`
}

type CreateFieldRequest struct {
	RecipientID int             `json:"recipientId"`
	Type        FieldType       `json:"type"`
	PageNumber  int             `json:"pageNumber"`
	PageX       float64         `json:"pageX"`
	PageY       float64         `json:"pageY"`
	PageWidth   float64         `json:"pageWidth"`
	PageHeight  float64         `json:"pageHeight"`
	FieldMeta   json.RawMessage `json:"fieldMeta,omitempty"`
}

// Template (subset; expand as commands grow).
type Template struct {
	ID         int             `json:"id"`
	Title      string          `json:"title"`
	Type       string          `json:"type,omitempty"`
	Recipients []Recipient     `json:"recipients,omitempty"`
	Fields     []Field         `json:"fields,omitempty"`
	Meta       json.RawMessage `json:"templateMeta,omitempty"`
}

type ListTemplatesResponse struct {
	Templates  []Template `json:"templates"`
	TotalPages int        `json:"totalPages"`
}

// Webhook.
type Webhook struct {
	ID         string    `json:"id"`
	WebhookURL string    `json:"webhookUrl"`
	EventTriggers []string `json:"eventTriggers"`
	Enabled    bool      `json:"enabled"`
	Secret     string    `json:"secret,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// APIError is the structured error shape ts-rest produces.
type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Body    []byte `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Body)
}
