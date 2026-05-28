import type {
  Policy,
  Run,
  SimulateResult,
  ExecuteResult,
  PocketBaseListResult,
} from "../models/types";

const DEFAULT_BASE_URL = "http://localhost:8090";

export class PocketBaseClient {
  private baseUrl: string;
  private token: string;

  constructor(baseUrl?: string) {
    this.baseUrl = baseUrl || DEFAULT_BASE_URL;
    this.token = "";
  }

  setToken(token: string) {
    this.token = token;
  }

  setBaseUrl(url: string) {
    this.baseUrl = url;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (this.token) {
      headers["Authorization"] = this.token;
    }

    const resp = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!resp.ok) {
      const text = await resp.text();
      throw new Error(`API error ${resp.status}: ${text}`);
    }

    if (resp.status === 204 || resp.headers.get("content-length") === "0") {
      return undefined as T;
    }

    return resp.json();
  }

  // Policy CRUD

  async getPolicies(
    page = 1,
    perPage = 50
  ): Promise<PocketBaseListResult<Policy>> {
    return this.request<PocketBaseListResult<Policy>>(
      "GET",
      `/api/collections/policies/records?page=${page}&perPage=${perPage}`
    );
  }

  async getPolicy(id: string): Promise<Policy> {
    return this.request<Policy>("GET", `/api/collections/policies/records/${id}`);
  }

  async createPolicy(data: Partial<Policy>): Promise<Policy> {
    return this.request<Policy>("POST", "/api/collections/policies/records", data);
  }

  async updatePolicy(id: string, data: Partial<Policy>): Promise<Policy> {
    return this.request<Policy>(
      "PATCH",
      `/api/collections/policies/records/${id}`,
      data
    );
  }

  async deletePolicy(id: string): Promise<void> {
    return this.request<void>(
      "DELETE",
      `/api/collections/policies/records/${id}`
    );
  }

  // Runs

  async getRuns(
    policyId?: string,
    page = 1,
    perPage = 50
  ): Promise<PocketBaseListResult<Run>> {
    let filter = "";
    if (policyId) {
      filter = `&filter=${encodeURIComponent(`policy='${policyId}'`)}`;
    }
    return this.request<PocketBaseListResult<Run>>(
      "GET",
      `/api/collections/runs/records?page=${page}&perPage=${perPage}&sort=-created${filter}`
    );
  }

  async getRun(id: string): Promise<Run> {
    return this.request<Run>("GET", `/api/collections/runs/records/${id}`);
  }

  // Custom endpoints

  async simulate(policyId: string): Promise<SimulateResult> {
    return this.request<SimulateResult>("POST", "/api/pr-governor/simulate", {
      policy_id: policyId,
    });
  }

  async execute(policyId: string, dryRun = false): Promise<ExecuteResult> {
    return this.request<ExecuteResult>("POST", "/api/pr-governor/execute", {
      policy_id: policyId,
      dry_run: dryRun,
    });
  }
}

// Singleton instance shared across the extension
let client: PocketBaseClient | null = null;

export function getClient(): PocketBaseClient {
  if (!client) {
    client = new PocketBaseClient();
  }
  return client;
}

export function initClient(baseUrl: string, token?: string): PocketBaseClient {
  client = new PocketBaseClient(baseUrl);
  if (token) {
    client.setToken(token);
  }
  return client;
}
