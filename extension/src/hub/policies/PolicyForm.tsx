import React, { useEffect, useState } from "react";
import { getClient } from "../../services/pocketbase";
import type { Policy, PolicyCondition, PolicyAction } from "../../models/types";

interface PolicyFormProps {
  policyId?: string;
  onBack: () => void;
  onSaved: () => void;
}

const EMPTY_POLICY: Partial<Policy> = {
  name: "",
  description: "",
  version: "1.0.0",
  enabled: true,
  scope_org: "",
  scope_project: "",
  scope_repo: "",
  schedule_cron: "0 1 * * *",
  schedule_timezone: "UTC",
  schedule_enabled: true,
  conditions: [],
  actions: [
    {
      type: "create_pr",
      parameters: {
        sourceRefName: "refs/heads/dev",
        targetRefName: "refs/heads/qa",
        title: "Auto PR: dev -> qa",
      },
    },
  ],
  constraints: null,
  tags: [],
};

export const PolicyForm: React.FC<PolicyFormProps> = ({
  policyId,
  onBack,
  onSaved,
}) => {
  const [policy, setPolicy] = useState<Partial<Policy>>(EMPTY_POLICY);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEdit = !!policyId;

  useEffect(() => {
    if (policyId) {
      loadPolicy(policyId);
    }
  }, [policyId]);

  const loadPolicy = async (id: string) => {
    try {
      setLoading(true);
      const loaded = await getClient().getPolicy(id);
      setPolicy(loaded);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load policy");
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!policy.name || !policy.scope_org || !policy.scope_project) {
      setError("Name, organization, and project are required");
      return;
    }

    try {
      setSaving(true);
      setError(null);
      if (isEdit && policyId) {
        await getClient().updatePolicy(policyId, policy);
      } else {
        await getClient().createPolicy(policy);
      }
      onSaved();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save policy");
    } finally {
      setSaving(false);
    }
  };

  const update = (field: string, value: unknown) => {
    setPolicy((prev) => ({ ...prev, [field]: value }));
  };

  const updateCondition = (
    index: number,
    field: string,
    value: unknown
  ) => {
    const conditions = [...(policy.conditions || [])];
    conditions[index] = { ...conditions[index], [field]: value };
    update("conditions", conditions);
  };

  const addCondition = () => {
    const conditions = [...(policy.conditions || []), { type: "branch_exists", parameters: {} }];
    update("conditions", conditions);
  };

  const removeCondition = (index: number) => {
    const conditions = [...(policy.conditions || [])];
    conditions.splice(index, 1);
    update("conditions", conditions);
  };

  const updateAction = (index: number, field: string, value: unknown) => {
    const actions = [...(policy.actions || [])];
    actions[index] = { ...actions[index], [field]: value };
    update("actions", actions);
  };

  const updateActionParam = (
    index: number,
    key: string,
    value: string
  ) => {
    const actions = [...(policy.actions || [])];
    const parameters = { ...actions[index].parameters, [key]: value };
    actions[index] = { ...actions[index], parameters };
    update("actions", actions);
  };

  const addAction = () => {
    const actions = [
      ...(policy.actions || []),
      { type: "create_pr", parameters: { sourceRefName: "", targetRefName: "", title: "" } },
    ];
    update("actions", actions);
  };

  const removeAction = (index: number) => {
    const actions = [...(policy.actions || [])];
    actions.splice(index, 1);
    update("actions", actions);
  };

  if (loading) {
    return <div style={{ padding: "20px", color: "#666" }}>Loading policy...</div>;
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
        <h2 style={{ margin: 0, fontSize: "18px", flex: 1 }}>
          {isEdit ? "Edit Policy" : "New Policy"}
        </h2>
        <button
          onClick={handleSave}
          disabled={saving}
          style={{
            padding: "6px 20px",
            background: "#0078d4",
            color: "white",
            border: "none",
            borderRadius: "2px",
            cursor: saving ? "wait" : "pointer",
            fontSize: "13px",
          }}
        >
          {saving ? "Saving..." : "Save"}
        </button>
      </div>

      {error && (
        <div style={{ padding: "10px", background: "#fde7e9", color: "#d13438", marginBottom: "16px", borderRadius: "2px" }}>
          {error}
        </div>
      )}

      {/* Form sections */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "24px" }}>
        {/* Left column */}
        <div>
          <Section title="General">
            <FormField label="Name" required>
              <TextInput value={policy.name || ""} onChange={(v) => update("name", v)} />
            </FormField>
            <FormField label="Description">
              <TextInput value={policy.description || ""} onChange={(v) => update("description", v)} />
            </FormField>
            <FormField label="Version">
              <TextInput value={policy.version || "1.0.0"} onChange={(v) => update("version", v)} />
            </FormField>
            <FormField label="Enabled">
              <input
                type="checkbox"
                checked={policy.enabled ?? true}
                onChange={(e) => update("enabled", e.target.checked)}
              />
            </FormField>
          </Section>

          <Section title="Scope">
            <FormField label="Organization" required>
              <TextInput value={policy.scope_org || ""} onChange={(v) => update("scope_org", v)} />
            </FormField>
            <FormField label="Project" required>
              <TextInput value={policy.scope_project || ""} onChange={(v) => update("scope_project", v)} />
            </FormField>
            <FormField label="Repository">
              <TextInput value={policy.scope_repo || ""} onChange={(v) => update("scope_repo", v)} placeholder="Leave empty for all repos" />
            </FormField>
          </Section>

          <Section title="Schedule">
            <FormField label="Cron Expression">
              <TextInput value={policy.schedule_cron || "0 1 * * *"} onChange={(v) => update("schedule_cron", v)} />
            </FormField>
            <FormField label="Timezone">
              <TextInput value={policy.schedule_timezone || "UTC"} onChange={(v) => update("schedule_timezone", v)} />
            </FormField>
            <FormField label="Schedule Active">
              <input
                type="checkbox"
                checked={policy.schedule_enabled ?? true}
                onChange={(e) => update("schedule_enabled", e.target.checked)}
              />
            </FormField>
          </Section>
        </div>

        {/* Right column */}
        <div>
          <Section
            title="Conditions"
            action={<SmallButton onClick={addCondition}>+ Add</SmallButton>}
          >
            {(policy.conditions || []).length === 0 ? (
              <div style={{ color: "#666", fontSize: "13px" }}>No conditions (always matches)</div>
            ) : (
              (policy.conditions || []).map((cond, i) => (
                <div
                  key={i}
                  style={{
                    marginBottom: "12px",
                    padding: "10px",
                    background: "#f9f9f9",
                    borderRadius: "2px",
                  }}
                >
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "6px" }}>
                    <select
                      value={cond.type}
                      onChange={(e) => updateCondition(i, "type", e.target.value)}
                      style={{ fontSize: "13px", padding: "4px" }}
                    >
                      <option value="branch_exists">branch_exists</option>
                      <option value="time_window">time_window</option>
                      <option value="pr_count">pr_count</option>
                      <option value="custom">custom</option>
                    </select>
                    <SmallButton onClick={() => removeCondition(i)}>Remove</SmallButton>
                  </div>
                  <textarea
                    value={JSON.stringify(cond.parameters, null, 2)}
                    onChange={(e) => {
                      try {
                        updateCondition(i, "parameters", JSON.parse(e.target.value));
                      } catch {
                        // Let user edit freely
                      }
                    }}
                    style={{ width: "100%", fontSize: "12px", fontFamily: "monospace", padding: "6px", minHeight: "60px" }}
                  />
                </div>
              ))
            )}
          </Section>

          <Section
            title="Actions"
            action={<SmallButton onClick={addAction}>+ Add</SmallButton>}
          >
            {(policy.actions || []).map((action, i) => (
              <div
                key={i}
                style={{
                  marginBottom: "12px",
                  padding: "10px",
                  background: "#f9f9f9",
                  borderRadius: "2px",
                }}
              >
                <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px" }}>
                  <select
                    value={action.type}
                    onChange={(e) => updateAction(i, "type", e.target.value)}
                    style={{ fontSize: "13px", padding: "4px" }}
                  >
                    <option value="create_pr">create_pr</option>
                  </select>
                  <SmallButton onClick={() => removeAction(i)}>Remove</SmallButton>
                </div>
                <FormField label="Source Ref">
                  <TextInput
                    value={(action.parameters.sourceRefName as string) || ""}
                    onChange={(v) => updateActionParam(i, "sourceRefName", v)}
                    placeholder="refs/heads/dev"
                  />
                </FormField>
                <FormField label="Target Ref">
                  <TextInput
                    value={(action.parameters.targetRefName as string) || ""}
                    onChange={(v) => updateActionParam(i, "targetRefName", v)}
                    placeholder="refs/heads/qa"
                  />
                </FormField>
                <FormField label="Title">
                  <TextInput
                    value={(action.parameters.title as string) || ""}
                    onChange={(v) => updateActionParam(i, "title", v)}
                    placeholder="Auto PR: dev -> qa"
                  />
                </FormField>
              </div>
            ))}
          </Section>
        </div>
      </div>
    </div>
  );
};

// Reusable form primitives

const Section: React.FC<{
  title: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}> = ({ title, action, children }) => (
  <div style={{ marginBottom: "20px" }}>
    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "10px" }}>
      <h3 style={{ margin: 0, fontSize: "14px", color: "#333" }}>{title}</h3>
      {action}
    </div>
    {children}
  </div>
);

const FormField: React.FC<{
  label: string;
  required?: boolean;
  children: React.ReactNode;
}> = ({ label, required, children }) => (
  <div style={{ marginBottom: "10px" }}>
    <label style={{ display: "block", fontSize: "12px", color: "#666", marginBottom: "3px" }}>
      {label}
      {required && <span style={{ color: "#d13438" }}> *</span>}
    </label>
    {children}
  </div>
);

const TextInput: React.FC<{
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
}> = ({ value, onChange, placeholder }) => (
  <input
    type="text"
    value={value}
    onChange={(e) => onChange(e.target.value)}
    placeholder={placeholder}
    style={{
      width: "100%",
      padding: "6px 8px",
      border: "1px solid #ccc",
      borderRadius: "2px",
      fontSize: "13px",
      boxSizing: "border-box",
    }}
  />
);

const SmallButton: React.FC<{
  onClick: () => void;
  children: React.ReactNode;
}> = ({ onClick, children }) => (
  <button
    onClick={onClick}
    style={{
      padding: "2px 10px",
      fontSize: "12px",
      background: "none",
      border: "1px solid #ccc",
      borderRadius: "2px",
      cursor: "pointer",
    }}
  >
    {children}
  </button>
);
