package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pangarabbit/azure-devops-pr-governor/internal/models"
	"github.com/pocketbase/pocketbase/core"
)

// AuditService records structured audit events for policy runs.
type AuditService struct {
	app core.App
}

// NewAuditService creates a new AuditService.
func NewAuditService(app core.App) *AuditService {
	return &AuditService{app: app}
}

// LogEvent records an audit event for a policy run.
func (s *AuditService) LogEvent(ctx context.Context, runID, eventType string, detail map[string]interface{}) error {
	collection, err := s.app.FindCollectionByNameOrId("audit_events")
	if err != nil {
		return fmt.Errorf("find audit_events collection: %w", err)
	}

	detailJSON, err := json.Marshal(detail)
	if err != nil {
		detailJSON = []byte("{}")
	}

	record := core.NewRecord(collection)
	record.Set("run", runID)
	record.Set("event_type", eventType)
	record.Set("detail", string(detailJSON))

	return s.app.Save(record)
}

// LogError records an error audit event.
func (s *AuditService) LogError(ctx context.Context, runID string, err error, detail map[string]interface{}) error {
	if detail == nil {
		detail = make(map[string]interface{})
	}
	detail["error"] = err.Error()

	return s.LogEvent(ctx, runID, string(models.EventTypeError), detail)
}

// GetRunAuditTrail returns all audit events for a given run.
func (s *AuditService) GetRunAuditTrail(ctx context.Context, runID string) ([]models.AuditEvent, error) {
	collection, err := s.app.FindCollectionByNameOrId("audit_events")
	if err != nil {
		return nil, fmt.Errorf("find audit_events collection: %w", err)
	}

	records, err := s.app.FindRecordsByFilter(
		collection.Id,
		"run = {:runId}",
		"created",
		0,
		0,
		map[string]interface{}{"runId": runID},
	)
	if err != nil {
		return nil, fmt.Errorf("query audit events: %w", err)
	}

	events := make([]models.AuditEvent, 0, len(records))
	for _, r := range records {
		prID := int(r.GetInt("ado_pr_id"))
		event := models.AuditEvent{
			ID:         r.Id,
			RunID:      r.GetString("run"),
			EventType:  models.AuditEventType(r.GetString("event_type")),
			Actor:      r.GetString("actor"),
			Detail:     r.GetString("detail"),
			AdoOrg:     r.GetString("ado_org"),
			AdoProject: r.GetString("ado_project"),
			AdoRepo:    r.GetString("ado_repo"),
			Created:    r.GetDateTime("created").Time(),
		}
		if prID > 0 {
			event.AdoPrID = &prID
		}
		events = append(events, event)
	}

	return events, nil
}

// GetRecentEvents returns recent audit events across all runs.
func (s *AuditService) GetRecentEvents(ctx context.Context, limit int) ([]models.AuditEvent, error) {
	collection, err := s.app.FindCollectionByNameOrId("audit_events")
	if err != nil {
		return nil, fmt.Errorf("find audit_events collection: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}

	records, err := s.app.FindRecordsByFilter(
		collection.Id,
		"1=1",
		"-created",
		limit,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("query recent audit events: %w", err)
	}

	events := make([]models.AuditEvent, 0, len(records))
	for _, r := range records {
		prID := int(r.GetInt("ado_pr_id"))
		event := models.AuditEvent{
			ID:         r.Id,
			RunID:      r.GetString("run"),
			EventType:  models.AuditEventType(r.GetString("event_type")),
			Actor:      r.GetString("actor"),
			Detail:     r.GetString("detail"),
			AdoOrg:     r.GetString("ado_org"),
			AdoProject: r.GetString("ado_project"),
			AdoRepo:    r.GetString("ado_repo"),
			Created:    r.GetDateTime("created").Time(),
		}
		if prID > 0 {
			event.AdoPrID = &prID
		}
		events = append(events, event)
	}

	return events, nil
}

// LogEventWithADO records an audit event with Azure DevOps resource context.
func (s *AuditService) LogEventWithADO(ctx context.Context, runID, eventType string, detail map[string]interface{}, org, project, repo string, prID *int) error {
	collection, err := s.app.FindCollectionByNameOrId("audit_events")
	if err != nil {
		return fmt.Errorf("find audit_events collection: %w", err)
	}

	detailJSON, err := json.Marshal(detail)
	if err != nil {
		detailJSON = []byte("{}")
	}

	record := core.NewRecord(collection)
	record.Set("run", runID)
	record.Set("event_type", eventType)
	record.Set("detail", string(detailJSON))
	record.Set("ado_org", org)
	record.Set("ado_project", project)
	record.Set("ado_repo", repo)
	if prID != nil {
		record.Set("ado_pr_id", *prID)
	}

	return s.app.Save(record)
}

// FormatTime formats a time for storage.
func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
