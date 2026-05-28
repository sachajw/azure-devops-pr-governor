import React, { useEffect, useState } from "react";
import { getClient } from "../../services/pocketbase";
import type { Policy, Run, SimulateResult } from "../../models/types";

interface PolicyDetailProps {
  policy: Policy;
  onBack: () => void;
  onEdit: () => void;
}

export const PolicyDetail: React.FC<PolicyDetailProps> = ({
  policy,
  onBack,
  onEdit,
}) => {
  const [runs, setRuns] = useState<Run[]>([]);
  const [simulating, setSimulating] = useState(false);
  const [executing, setExecuting] = useState(false);
  const [simulateResult, setSimulateResult] = useState<SimulateResult | null>(
    null
  );
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadRuns();
  }, [policy.id]);

  const loadRuns = async () => {
    try {
      const result = await getClient().getRuns(policy.id, 1, 10);
      setRuns(result.items);
    } catch {
      // Non-critical: runs are supplementary info
    }
  };

  const handleSimulate = async () => {
    try {
      setSimulating(true);
      setError(null);
      const result = await getClient().simulate(policy.id);
      setSimulateResult(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Simulation failed");
    } finally {
      setSimulating(false);
    }
  };

  const handleExecute = async (dryRun: boolean) => {
    try {
      setExecuting(true);
      setError(null);
      await getClient().execute(policy.id, dryRun);
      loadRuns();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Execution failed");
    } finally {
      setExecuting(false);
    }
  };

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
        <h2 style={{ margin: 0, fontSize: "18px", flex: 1 }}>{policy.name}</h2>
        <button
          onClick={handleSimulate}
          disabled={simulating}
          style={{
            padding: "6px 16px",
            background: "#fff",
            border: "1px solid #0078d4",
            color: "#0078d4",
            borderRadius: "2px",
            cursor: simulating ? "wait" : "pointer",
            fontSize: "13px",
          }}
        >
          {simulating ? "Simulating..." : "Simulate"}
        </button>
        <button
          onClick={() => handleExecute(true)}
          disabled={executing}
          style={{
            padding: "6px 16px",
            background: "#fff",
            border: "1px solid #666",
            borderRadius: "2px",
            cursor: executing ? "wait" : "pointer",
            fontSize: "13px",
          }}
        >
          Dry Run
        </button>
        <button
          onClick={() => handleExecute(false)}
          disabled={executing}
          style={{
            padding: "6px 16px",
            background: "#0078d4",
            color: "white",
            border: "none",
            borderRadius: "2px",
            cursor: executing ? "wait" : "pointer",
            fontSize: "13px",
          }}
        >
          Execute
        </button>
        <button
          onClick={onEdit}
          style={{
            padding: "6px 16px",
            background: "#fff",
            border: "1px solid #ccc",
            borderRadius: "2px",
            cursor: "pointer",
            fontSize: "13px",
          }}
        >
          Edit
        </button>
      </div>

      {error && (
        <div style={{ padding: "10px", background: "#fde7e9", color: "#d13438", marginBottom: "16px", borderRadius: "2px" }}>
          {error}
        </div>
      )}

      {/* Simulate results */}
      {simulateResult && (
        <div style={{ marginBottom: "20px", padding: "16px", background: "#f5f5f5", borderRadius: "4px" }}>
          <h3 style={{ margin: "0 0 12px 0", fontSize: "14px" }}>Simulation Results</h3>
          <div style={{ marginBottom: "8px" }}>
            <strong>Conditions:</strong>
            {simulateResult.conditions.map((c, i) => (
              <div key={i} style={{ marginLeft: "12px", fontSize: "13px" }}>
                <span style={{ color: c.met ? "#107c10" : "#d13438" }}>
                  {c.met ? "PASS" : "FAIL"}
                </span>{" "}
                {c.type}: {c.detail}
              </div>
            ))}
          </div>
          <div>
            <strong>Actions:</strong>
            {simulateResult.actions.map((a, i) => (
              <div key={i} style={{ marginLeft: "12px", fontSize: "13px" }}>
                <span style={{ color: a.success ? "#107c10" : "#d13438" }}>
                  {a.success ? "WOULD RUN" : "SKIPPED"}
                </span>{" "}
                {a.type}: {a.detail}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Policy details grid */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "20px", marginBottom: "24px" }}>
        {/* Summary */}
        <div>
          <Section title="Summary">
            <Field label="Description" value={policy.description || "—"} />
            <Field label="Version" value={policy.version} />
            <Field label="Enabled" value={policy.enabled ? "Yes" : "No"} />
          </Section>
          <Section title="Scope">
            <Field label="Organization" value={policy.scope_org} />
            <Field label="Project" value={policy.scope_project} />
            <Field label="Repository" value={policy.scope_repo || "All repos"} />
          </Section>
        </div>

        <div>
          <Section title="Schedule">
            <Field label="Cron" value={policy.schedule_cron} />
            <Field label="Timezone" value={policy.schedule_timezone} />
            <Field label="Active" value={policy.schedule_enabled ? "Yes" : "No"} />
          </Section>
          <Section title="Actions">
            {policy.actions.length === 0 ? (
              <div style={{ color: "#666", fontSize: "13px" }}>No actions defined</div>
            ) : (
              policy.actions.map((action, i) => (
                <div key={i} style={{ marginBottom: "8px", fontSize: "13px" }}>
                  <strong>{action.type}</strong>
                  <pre style={{ margin: "4px 0", fontSize: "12px", background: "#f5f5f5", padding: "8px" }}>
                    {JSON.stringify(action.parameters, null, 2)}
                  </pre>
                </div>
              ))
            )}
          </Section>
        </div>
      </div>

      {/* Conditions */}
      <Section title="Conditions">
        {policy.conditions.length === 0 ? (
          <div style={{ color: "#666", fontSize: "13px" }}>No conditions defined (always matches)</div>
        ) : (
          policy.conditions.map((cond, i) => (
            <div key={i} style={{ marginBottom: "8px", fontSize: "13px" }}>
              <strong>{cond.type}</strong>
              <pre style={{ margin: "4px 0", fontSize: "12px", background: "#f5f5f5", padding: "8px" }}>
                {JSON.stringify(cond.parameters, null, 2)}
              </pre>
            </div>
          ))
        )}
      </Section>

      {/* Recent runs */}
      <Section title="Recent Runs">
        {runs.length === 0 ? (
          <div style={{ color: "#666", fontSize: "13px" }}>No runs yet</div>
        ) : (
          <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "13px" }}>
            <thead>
              <tr style={{ borderBottom: "2px solid #e0e0e0", textAlign: "left" }}>
                <th style={{ padding: "6px 8px" }}>Status</th>
                <th style={{ padding: "6px 8px" }}>Triggered By</th>
                <th style={{ padding: "6px 8px" }}>Dry Run</th>
                <th style={{ padding: "6px 8px" }}>Started</th>
                <th style={{ padding: "6px 8px" }}>Error</th>
              </tr>
            </thead>
            <tbody>
              {runs.map((run) => (
                <tr key={run.id} style={{ borderBottom: "1px solid #f0f0f0" }}>
                  <td style={{ padding: "6px 8px" }}>
                    <span style={{ color: statusColor(run.status), fontWeight: 500 }}>
                      {run.status}
                    </span>
                  </td>
                  <td style={{ padding: "6px 8px" }}>{run.triggered_by}</td>
                  <td style={{ padding: "6px 8px" }}>{run.dry_run ? "Yes" : "No"}</td>
                  <td style={{ padding: "6px 8px" }}>
                    {new Date(run.started_at).toLocaleString()}
                  </td>
                  <td style={{ padding: "6px 8px", color: "#d13438", maxWidth: "300px", overflow: "hidden", textOverflow: "ellipsis" }}>
                    {run.error || "—"}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>
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
    <span>{value}</span>
  </div>
);
