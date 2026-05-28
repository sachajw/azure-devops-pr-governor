package azuredevops

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://dev.azure.com/myorg", "my-pat")
	if client.orgURL != "https://dev.azure.com/myorg" {
		t.Errorf("expected orgURL to be set, got %s", client.orgURL)
	}
	if client.apiVersion != "7.1" {
		t.Errorf("expected apiVersion 7.1, got %s", client.apiVersion)
	}
}

func TestCreatePullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("expected Authorization header")
		}

		var req CreatePRRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.SourceRefName != "refs/heads/dev" {
			t.Errorf("expected sourceRefName refs/heads/dev, got %s", req.SourceRefName)
		}
		if req.TargetRefName != "refs/heads/qa" {
			t.Errorf("expected targetRefName refs/heads/qa, got %s", req.TargetRefName)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PRResult{
			ID:            42,
			Status:        "active",
			SourceRefName: req.SourceRefName,
			TargetRefName: req.TargetRefName,
		})
	}))
	defer server.Close()

	client := NewClient("https://dev.azure.com/myorg", "my-pat")
	client.orgURL = server.URL

	result, err := client.CreatePullRequest(context.Background(), "myproject", "myrepo", &CreatePRRequest{
		SourceRefName: "refs/heads/dev",
		TargetRefName: "refs/heads/qa",
		Title:        "Auto PR: dev -> qa",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 42 {
		t.Errorf("expected PR ID 42, got %d", result.ID)
	}
}

func TestGetRefs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		filter := r.URL.Query().Get("filter")
		if filter != "refs/heads/dev" {
			t.Errorf("expected filter refs/heads/dev, got %s", filter)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RefsResponse{
			Value: []Ref{
				{Name: "refs/heads/dev", ObjectID: "abc123"},
			},
			Count: 1,
		})
	}))
	defer server.Close()

	client := NewClient("https://dev.azure.com/myorg", "my-pat")
	client.orgURL = server.URL

	refs, err := client.GetRefs(context.Background(), "myproject", "myrepo", "refs/heads/dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	if refs[0].Name != "refs/heads/dev" {
		t.Errorf("expected ref name refs/heads/dev, got %s", refs[0].Name)
	}
}

func TestBranchExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RefsResponse{
			Value: []Ref{
				{Name: "refs/heads/dev", ObjectID: "abc123"},
			},
			Count: 1,
		})
	}))
	defer server.Close()

	client := NewClient("https://dev.azure.com/myorg", "my-pat")
	client.orgURL = server.URL

	exists, err := client.BranchExists(context.Background(), "myproject", "myrepo", "refs/heads/dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected branch to exist")
	}
}

func TestBranchNotExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RefsResponse{
			Value: []Ref{},
			Count:  0,
		})
	}))
	defer server.Close()

	client := NewClient("https://dev.azure.com/myorg", "my-pat")
	client.orgURL = server.URL

	exists, err := client.BranchExists(context.Background(), "myproject", "myrepo", "refs/heads/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected branch not to exist")
	}
}

func TestApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient("https://dev.azure.com/myorg", "my-pat")
	client.orgURL = server.URL

	_, err := client.CreatePullRequest(context.Background(), "myproject", "myrepo", &CreatePRRequest{
		SourceRefName: "refs/heads/dev",
		TargetRefName: "refs/heads/qa",
		Title:        "Test",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var adoErr *ADOApiError
	if !errors.As(err, &adoErr) {
		t.Fatalf("expected ADOApiError, got %T", err)
	}
	if adoErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", adoErr.StatusCode)
	}
}
