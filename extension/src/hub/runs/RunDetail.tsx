import React from "react";
import type { Run } from "../../models/types";

interface RunDetailProps {
  run: Run;
  onBack: () => void;
}

export const RunDetail: React.FC<RunDetailProps> = ({ run, onBack }) => {
  const statusColor = (status: string) => {
    switch (status) {
      case "succeeded":
        return "#107c10";
      case "failed":
        return "#d13438";
      case "running":
        return "#0078d4";
      case "dry_run":
        return "#8a8886";
      default:
        return "#666";
    }
  };

  let resultSummary: Record<string, unknown> | null = null;
  if (run.result_summary) {
    try {
      resultSummary = JSON.parse(run.result_summary);
    } catch {
      resultSummary = null;
    }
  }

  return (
    <div>
      {/* Header */}
      <div style={{ display: "flex", alignItems: "center", gap: "12px", marginBottom: "20px" }}>
        <button
          onClick={onBack}
          style={{ background: "none", border: "none", cursor: "pointer", fontSize: "14px" }}
        >
          &larr; Back
        </button>
        <h2 style={{ margin: 0, fontSize: "18px", flex: 1 }}>Run Detail</h2>
        <span
          style={{
            padding: "4px 12px",
            borderRadius: "2px",
            background: statusColor(run.status) + "20",
            color: statusColor(run.status),
            fontWeight: 500,
            fontSize: "13px",
          }}
        >
          {run.status}
        </span>
      </div>

      {/* Run info grid */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "20px", marginBottom: "24px" }}>
        <div>
          <Section title="Run Info">
            <Field label="ID" value={run.id} />
            <Field label="Triggered By" value={run.triggered_by} />
            <Field label="Dry Run" value={run.dry_run ? "Yes" : "No"} />
            <Field label="Started" value={new Date(run.started_at).toLocaleString()} />
            <Field
              label="Completed"
              value={run.completed_at ? new Date(run.completed_at).toLocaleString() : "—"}
            />
          </Section>
        </div>

        <div>
          <Section title="Result">
            {run.error && (
              <div
                style={{
                  padding: "12px",
                  background: "#fde7e9",
                  color: "#d13438",
                  borderRadius: "2px",
                  fontSize: "13px",
                  marginBottom: "12px",
                }}
              >
                {run.error}
              </div>
            )}
            {resultSummary && (
              <pre
                style={{
                  background: "#f5f5f5",
                  padding: "12px",
                  borderRadius: "2px",
                  fontSize: "12px",
                  overflow: "auto",
                  maxHeight: "300px",
                }}
              >
                {JSON.stringify(resultSummary, null, 2)}
              </pre>
            )}
            {!run.error && !resultSummary && (
              <div style={{ color: "#666", fontSize: "13px" }}>No result details available</div>
            )}
          </Section>
        </div>
      </div>
    </div>
  );
};

const Section: React.FC<{ title: string; children: React.ReactNode }> = ({
  title,
  children,
}) => (
  <div style={{ marginBottom: "16px" }}>
    <h3 style={{ margin: "0 0 8px 0", fontSize: "14px", color: "#333" }}>{title}</h3>
    {children}
  </div>
);

const Field: React.FC<{ label: string; value: string }> = ({ label, value }) => (
  <div style={{ display: "flex", fontSize: "13px", marginBottom: "4px" }}>
    <span style={{ width: "120px", color: "#666" }}>{label}:</span>
    <span style={{ wordBreak: "break-all" }}>{value}</span>
  </div>
);
