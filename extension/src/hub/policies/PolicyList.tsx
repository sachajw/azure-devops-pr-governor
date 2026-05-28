import React, { useEffect, useState } from "react";
import { getClient } from "../../services/pocketbase";
import type { Policy } from "../../models/types";

interface PolicyListProps {
  onSelectPolicy: (policy: Policy) => void;
  onNewPolicy: () => void;
}

export const PolicyList: React.FC<PolicyListProps> = ({
  onSelectPolicy,
  onNewPolicy,
}) => {
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState("");

  useEffect(() => {
    loadPolicies();
  }, []);

  const loadPolicies = async () => {
    try {
      setLoading(true);
      setError(null);
      const result = await getClient().getPolicies();
      setPolicies(result.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load policies");
    } finally {
      setLoading(false);
    }
  };

  const filtered = policies.filter(
    (p) =>
      !filter ||
      p.name.toLowerCase().includes(filter.toLowerCase()) ||
      p.scope_org.toLowerCase().includes(filter.toLowerCase()) ||
      p.scope_project.toLowerCase().includes(filter.toLowerCase()) ||
      p.scope_repo.toLowerCase().includes(filter.toLowerCase())
  );

  if (loading) {
    return <div style={{ padding: "20px", color: "#666" }}>Loading policies...</div>;
  }

  if (error) {
    return (
      <div style={{ padding: "20px", color: "#d13438" }}>
        Error: {error}
        <button
          onClick={loadPolicies}
          style={{ marginLeft: "12px", cursor: "pointer" }}
        >
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
        <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
          <h2 style={{ margin: 0, fontSize: "18px" }}>Policies</h2>
          <span style={{ color: "#666", fontSize: "13px" }}>
            {policies.length} total
          </span>
        </div>
        <button
          onClick={onNewPolicy}
          style={{
            padding: "6px 16px",
            background: "#0078d4",
            color: "white",
            border: "none",
            borderRadius: "2px",
            cursor: "pointer",
            fontSize: "13px",
          }}
        >
          + New Policy
        </button>
      </div>

      {/* Filter */}
      <input
        type="text"
        placeholder="Filter by name or scope..."
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        style={{
          width: "300px",
          padding: "6px 10px",
          border: "1px solid #ccc",
          borderRadius: "2px",
          fontSize: "13px",
          marginBottom: "12px",
        }}
      />

      {/* Table */}
      {filtered.length === 0 ? (
        <div style={{ padding: "40px", textAlign: "center", color: "#666" }}>
          {policies.length === 0
            ? "No policies yet. Create one to get started."
            : "No policies match your filter."}
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
              <th style={{ padding: "8px 12px" }}>Name</th>
              <th style={{ padding: "8px 12px" }}>Scope</th>
              <th style={{ padding: "8px 12px" }}>Enabled</th>
              <th style={{ padding: "8px 12px" }}>Schedule</th>
              <th style={{ padding: "8px 12px" }}>Updated</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((policy) => (
              <tr
                key={policy.id}
                onClick={() => onSelectPolicy(policy)}
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
                <td style={{ padding: "8px 12px", fontWeight: 500 }}>
                  {policy.name}
                  {policy.description && (
                    <div
                      style={{
                        fontWeight: 400,
                        color: "#666",
                        fontSize: "12px",
                        marginTop: "2px",
                      }}
                    >
                      {policy.description}
                    </div>
                  )}
                </td>
                <td style={{ padding: "8px 12px" }}>
                  {policy.scope_org}/{policy.scope_project}
                  {policy.scope_repo && `/${policy.scope_repo}`}
                </td>
                <td style={{ padding: "8px 12px" }}>
                  <span
                    style={{
                      display: "inline-block",
                      width: "8px",
                      height: "8px",
                      borderRadius: "50%",
                      background: policy.enabled ? "#107c10" : "#a0a0a0",
                      marginRight: "6px",
                    }}
                  />
                  {policy.enabled ? "Enabled" : "Disabled"}
                </td>
                <td style={{ padding: "8px 12px" }}>
                  <code style={{ fontSize: "12px" }}>{policy.schedule_cron}</code>
                  {policy.schedule_timezone !== "UTC" && (
                    <span style={{ color: "#666", marginLeft: "4px" }}>
                      ({policy.schedule_timezone})
                    </span>
                  )}
                </td>
                <td style={{ padding: "8px 12px", color: "#666" }}>
                  {new Date(policy.updated).toLocaleDateString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
};
