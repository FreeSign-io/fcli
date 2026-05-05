package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FreeSign-io/fcli/internal/api"
)

func TestClient_AuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		if got != "Bearer api_test_token" {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer api_test_token")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"totalPages":0}`))
	}))
	t.Cleanup(srv.Close)

	c, err := api.New(api.Options{BaseURL: srv.URL, Token: "api_test_token"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.ListDocuments(context.Background(), 1, 1); err != nil {
		t.Fatal(err)
	}
}

func TestClient_StripsBearerPrefix(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	c, _ := api.New(api.Options{BaseURL: srv.URL, Token: "Bearer api_xyz"})
	_, _ = c.ListDocuments(context.Background(), 1, 1)
	if got != "Bearer api_xyz" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer api_xyz")
	}
}

func TestClient_RetriesOn5xx(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"data":[],"totalPages":0}`))
	}))
	t.Cleanup(srv.Close)

	c, _ := api.New(api.Options{BaseURL: srv.URL, Token: "api_x", MaxRetries: 5})
	if _, err := c.ListDocuments(context.Background(), 1, 1); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls (2 retries), got %d", calls)
	}
}

func TestClient_PropagatesAPIErrorOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid token"}`))
	}))
	t.Cleanup(srv.Close)

	c, _ := api.New(api.Options{BaseURL: srv.URL, Token: "api_bad"})
	_, err := c.ListDocuments(context.Background(), 1, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !api.IsAuthError(err) {
		t.Errorf("IsAuthError = false; err = %v", err)
	}
	var apiErr *api.APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected *api.APIError")
	}
	if !strings.Contains(apiErr.Message, "invalid token") {
		t.Errorf("APIError.Message = %q, want to contain 'invalid token'", apiErr.Message)
	}
}

func TestClient_DocumentRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/api/v1/documents":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"uploadUrl": "https://example.invalid/upload",
				"documentId": 42,
				"recipients": [{"recipientId": 1, "name": "X", "email": "x@y.z", "token": "tok", "signingUrl": "u", "role": "SIGNER"}]
			}`))
		case r.Method == "GET" && r.URL.Path == "/api/v1/documents/42":
			_, _ = w.Write([]byte(`{
				"id": 42, "title": "Test", "status": "DRAFT",
				"createdAt": "2026-05-04T00:00:00Z",
				"recipients": []
			}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	c, _ := api.New(api.Options{BaseURL: srv.URL, Token: "api_x"})
	created, err := c.CreateDocument(context.Background(), api.CreateDocumentRequest{Title: "Test"})
	if err != nil {
		t.Fatal(err)
	}
	if created.DocumentID != 42 {
		t.Errorf("DocumentID = %d, want 42", created.DocumentID)
	}
	doc, err := c.GetDocument(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "Test" {
		t.Errorf("Title = %q, want Test", doc.Title)
	}
}
