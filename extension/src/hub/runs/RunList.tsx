import React, { useEffect, useState } from "react";
import { getClient } from "../../services/pocketbase";
import type { Run, Policy } from "../../models/types";

interface RunListProps {
  onSelectRun: (run: Run) => void;
  onSelectPolicy: (policy: Policy) => void;
}

export const RunList: React.FC<RunListProps> = ({
  onSelectRun,
  onSelectPolicy,
}) => {
  const [runs, setRuns] = useState<Run[]>([]);
  const [policies, setPolicies] = useState<Map<string, Policy>>(new Map());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>("");

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [runsResult, policiesResult] = await Promise.all([
        getClient().getRuns(),
        getClient().getPolicies(1, 200),
      ]);
      setRuns(runsResult.items);
      setPolicies(
        new Map(policiesResult.items.map((p) => [p.id, p]))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load runs");
    } finally {
      setLoading(false);
    }
  };

  const filtered = runs.filter(
    (r) => !statusFilter || r.status === statusFilter
  );

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

  const policyName = (policyId: string) => {
    const p = policies.get(policyId);
    return p ? p.name : policyId;
  };

  if (loading) {
    return <div style={{ padding: "20px", color: "#666" }}>Loading runs...</div>;
  }

  if (error) {
    return (
      <div style={{ padding: "20px", color: "#d13438" }}>
        Error: {error}
        <button onClick={loadData} style={{ marginLeft: "12px", cursor: "pointer" }}>
          Retry
        </button>
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "16px",
        }}
      >
        <h2 style={{ margin: 0, fontSize: "18px" }}>Run History</h2>
        <span style={{ color: "#666", fontSize: "13px" }}>
          {runs.length} total runs
        </span>
      </div>

      {/* Filter */}
      <select
        value={statusFilter}
        onChange={(e) => setStatusFilter(e.target.value)}
        style={{
          padding: "6px 10px",
          border: "1px solid #ccc",
          borderRadius: "2px",
          fontSize: "13px",
          marginBottom: "12px",
        }}
      >
        <option value="">All statuses</option>
        <option value="succeeded">Succeeded</option>
        <option value="failed">Failed</option>
        <option value="running">Running</option>
        <option value="pending">Pending</option>
        <option value="dry_run">Dry Run</option>
        <option value="skipped">Skipped</option>
      </select>

      {/* Table */}
      {filtered.length === 0 ? (
        <div style={{ padding: "40px", textAlign: "center", color: "#666" }}>
          No runs found.
        </div>
      ) : (
        <table
          style={{
            width: "100%",
            borderCollapse: "collapse",
            fontSize: "13px",
          }}
        >
          <thead>
            <tr style={{ borderBottom: "2px solid #e0e0e0", textAlign: "left" }}>
              <th style={{ padding: "8px 12px" }}>Status</th>
              <th style={{ padding: "8px 12px" }}>Policy</th>
              <th style={{ padding: "8px 12px" }}>Triggered By</th>
              <th style={{ padding: "8px 12px" }}>Dry Run</th>
              <th style={{ padding: "8px 12px" }}>Started</th>
              <th style={{ padding: "8px 12px" }}>Completed</th>
              <th style={{ padding: "8px 12px" }}>Error</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((run) => (
              <tr
                key={run.id}
                onClick={() => onSelectRun(run)}
                style={{
                  borderBottom: "1px solid #f0f0f0",
                  cursor: "pointer",
                }}
                onMouseOver={(e) =>
                  (e.currentTarget.style.background = "#f5f5f5")
                }
                onMouseOut={(e) =>
                  (e.currentTarget.style.background = "transparent")
                }
              >
                <td style={{ padding: "8px 12px" }}>
                  <span style={{ color: statusColor(run.status), fontWeight: 500 }}>
                    {run.status}
                  </span>
                </td>
                <td style={{ padding: "8px 12px" }}>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      const p = policies.get(run.policy);
                      if (p) onSelectPolicy(p);
                    }}
                    style={{
                      background: "none",
                      border: "none",
                      color: "#0078d4",
                      cursor: "pointer",
                      padding: 0,
                      fontSize: "13px",
                      textDecoration: "underline",
                    }}
                  >
                    {policyName(run.policy)}
                  </button>
                </td>
                <td style={{ padding: "8px 12px" }}>{run.triggered_by}</td>
                <td style={{ padding: "8px 12px" }}>
                  {run.dry_run ? "Yes" : "No"}
                </td>
                <td style={{ padding: "8px 12px" }}>
                  {new Date(run.started_at).toLocaleString()}
                </td>
                <td style={{ padding: "8px 12px" }}>
                  {run.completed_at
                    ? new Date(run.completed_at).toLocaleString()
                    : "—"}
                </td>
                <td
                  style={{
                    padding: "8px 12px",
                    color: "#d13438",
                    maxWidth: "200px",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  }}
                >
                  {run.error || "—"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
};
