package azuredevops

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a thin HTTP client for the Azure DevOps REST API.
type Client struct {
	orgURL     string
	pat        string
	apiVersion string
	httpClient *http.Client
}

// NewClient creates a new Azure DevOps API client.
func NewClient(orgURL, pat string) *Client {
	return &Client{
		orgURL:     orgURL,
		pat:        pat,
		apiVersion: "7.1",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreatePRRequest is the payload for creating a pull request.
type CreatePRRequest struct {
	SourceRefName string `json:"sourceRefName"`
	TargetRefName string `json:"targetRefName"`
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	IsDraft       *bool  `json:"isDraft,omitempty"`
}

// PRResult is the response from creating a pull request.
type PRResult struct {
	ID            int    `json:"pullRequestId"`
	Status        string `json:"status"`
	CreatedBy     string `json:"createdBy,omitempty"`
	SourceRefName string `json:"sourceRefName,omitempty"`
	TargetRefName string `json:"targetRefName,omitempty"`
	URL           string `json:"url,omitempty"`
}

// Ref represents a Git ref (branch/tag) in Azure DevOps.
type Ref struct {
	Name     string `json:"name"`
	ObjectID string `json:"objectId"`
}

// RefsResponse is the paginated response for listing refs.
type RefsResponse struct {
	Value []Ref `json:"value"`
	Count int   `json:"count"`
}

// ADOApiError represents an error from the Azure DevOps API.
type ADOApiError struct {
	StatusCode int
	Message    string
}

func (e *ADOApiError) Error() string {
	return fmt.Sprintf("azure devops api error: status %d: %s", e.StatusCode, e.Message)
}

func (c *Client) authHeader() string {
	encoded := base64.StdEncoding.EncodeToString([]byte(":" + c.pat))
	return "Basic " + encoded
}

func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, resp.StatusCode, &ADOApiError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return respBody, resp.StatusCode, nil
}

// CreatePullRequest creates a new pull request in the specified repository.
func (c *Client) CreatePullRequest(ctx context.Context, project, repoID string, req *CreatePRRequest) (*PRResult, error) {
	url := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests?api-version=%s",
		c.orgURL, project, repoID, c.apiVersion)

	body, _, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}

	var result PRResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// GetRefs lists refs (branches/tags) in a repository, optionally filtered.
func (c *Client) GetRefs(ctx context.Context, project, repoID, filter string) ([]Ref, error) {
	url := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/refs?api-version=%s",
		c.orgURL, project, repoID, c.apiVersion)

	if filter != "" {
		url += "&filter=" + filter
	}

	body, _, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("get refs: %w", err)
	}

	var resp RefsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.Value, nil
}

// BranchExists checks whether a branch exists in the repository.
func (c *Client) BranchExists(ctx context.Context, project, repoID, branchName string) (bool, error) {
	refs, err := c.GetRefs(ctx, project, repoID, branchName)
	if err != nil {
		return false, err
	}
	return len(refs) > 0, nil
}
