import * as SDK from "azure-devops-extension-sdk";
import React from "react";
import { createRoot } from "react-dom/client";
import { HubRoot } from "./hub/HubRoot";

SDK.init({ loaded: false, applyTheme: true })
  .then(() => {
    const container = document.getElementById("root");
    if (!container) throw new Error("Root element not found");

    const root = createRoot(container);
    root.render(
      <React.StrictMode>
        <HubRoot />
      </React.StrictMode>
    );

    return SDK.notifyLoadSucceeded();
  })
  .catch((err) => {
    console.error("Extension init failed:", err);
    SDK.notifyLoadFailed(err);
  });
