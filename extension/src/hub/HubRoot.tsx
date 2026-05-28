import React, { useState } from "react";
import { PolicyList } from "./policies/PolicyList";
import { PolicyDetail } from "./policies/PolicyDetail";
import { PolicyForm } from "./policies/PolicyForm";
import { RunList } from "./runs/RunList";
import { RunDetail } from "./runs/RunDetail";
import type { Policy, Run } from "../models/types";

type Tab = "policies" | "runs";
type View =
  | { type: "policy-list" }
  | { type: "policy-detail"; policyId: string }
  | { type: "policy-form"; policyId?: string }
  | { type: "run-list" }
  | { type: "run-detail"; runId: string };

export const HubRoot: React.FC = () => {
  const [tab, setTab] = useState<Tab>("policies");
  const [view, setView] = useState<View>({ type: "policy-list" });
  const [selectedPolicy, setSelectedPolicy] = useState<Policy | null>(null);
  const [selectedRun, setSelectedRun] = useState<Run | null>(null);

  const handleSelectPolicy = (policy: Policy) => {
    setSelectedPolicy(policy);
    setView({ type: "policy-detail", policyId: policy.id });
  };

  const handleSelectRun = (run: Run) => {
    setSelectedRun(run);
    setView({ type: "run-detail", runId: run.id });
  };

  const handleNewPolicy = () => {
    setView({ type: "policy-form" });
  };

  const handleEditPolicy = (policyId: string) => {
    setView({ type: "policy-form", policyId });
  };

  const handleBack = () => {
    if (tab === "policies") {
      setSelectedPolicy(null);
      setView({ type: "policy-list" });
    } else {
      setSelectedRun(null);
      setView({ type: "run-list" });
    }
  };

  const handleTabChange = (newTab: Tab) => {
    setTab(newTab);
    if (newTab === "policies") {
      setSelectedPolicy(null);
      setView({ type: "policy-list" });
    } else {
      setSelectedRun(null);
      setView({ type: "run-list" });
    }
  };

  return (
    <div style={{ padding: "16px 24px", fontFamily: "Segoe UI, sans-serif" }}>
      {/* Tab bar */}
      <div
        style={{
          display: "flex",
          gap: "0",
          borderBottom: "1px solid #e0e0e0",
          marginBottom: "20px",
        }}
      >
        <TabButton
          active={tab === "policies"}
          onClick={() => handleTabChange("policies")}
        >
          Policies
        </TabButton>
        <TabButton
          active={tab === "runs"}
          onClick={() => handleTabChange("runs")}
        >
          Run History
        </TabButton>
      </div>

      {/* Content area */}
      {tab === "policies" && (
        <>
          {view.type === "policy-list" && (
            <PolicyList
              onSelectPolicy={handleSelectPolicy}
              onNewPolicy={handleNewPolicy}
            />
          )}
          {view.type === "policy-detail" && selectedPolicy && (
            <PolicyDetail
              policy={selectedPolicy}
              onBack={handleBack}
              onEdit={() => handleEditPolicy(selectedPolicy.id)}
            />
          )}
          {view.type === "policy-form" && (
            <PolicyForm
              policyId={view.policyId}
              onBack={handleBack}
              onSaved={handleBack}
            />
          )}
        </>
      )}

      {tab === "runs" && (
        <>
          {view.type === "run-list" && (
            <RunList
              onSelectRun={handleSelectRun}
              onSelectPolicy={(policy) => {
                setTab("policies");
                handleSelectPolicy(policy);
              }}
            />
          )}
          {view.type === "run-detail" && selectedRun && (
            <RunDetail run={selectedRun} onBack={handleBack} />
          )}
        </>
      )}
    </div>
  );
};

const TabButton: React.FC<{
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}> = ({ active, onClick, children }) => (
  <button
    onClick={onClick}
    style={{
      padding: "8px 20px",
      border: "none",
      background: "none",
      cursor: "pointer",
      fontSize: "14px",
      fontWeight: active ? 600 : 400,
      color: active ? "#0078d4" : "#666",
      borderBottom: active ? "2px solid #0078d4" : "2px solid transparent",
      transition: "all 0.15s ease",
    }}
  >
    {children}
  </button>
);
