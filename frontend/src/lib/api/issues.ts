import { apiGet } from "./client";
import type { Issue, IssueDetail } from "./types";

export async function listIssues(params?: {
  limit?: number;
  offset?: number;
  status?: string;
}) {
  const query = new URLSearchParams();
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.offset) query.set("offset", String(params.offset));
  if (params?.status) query.set("status", params.status);
  const qs = query.toString();
  return apiGet<Issue[]>(`/api/v1/issues${qs ? `?${qs}` : ""}`);
}

export async function getIssue(id: string) {
  return apiGet<IssueDetail>(`/api/v1/issues/${encodeURIComponent(id)}`);
}
