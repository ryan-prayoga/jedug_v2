import { apiGet } from "./client";
import type { Issue, IssueDetail, IssueTimelineEvent } from "./types";

export interface ListIssuesParams {
  limit?: number;
  offset?: number;
  status?: string;
  severity?: number;
  bbox?: [number, number, number, number]; // [minLng, minLat, maxLng, maxLat]
}

export async function listIssues(params?: ListIssuesParams) {
  const query = new URLSearchParams();
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.offset) query.set("offset", String(params.offset));
  if (params?.status) query.set("status", params.status);
  if (params?.severity) query.set("severity", String(params.severity));
  if (params?.bbox) query.set("bbox", params.bbox.join(","));
  const qs = query.toString();
  return apiGet<Issue[]>(`/api/v1/issues${qs ? `?${qs}` : ""}`);
}

export async function getIssue(id: string) {
  return apiGet<IssueDetail>(`/api/v1/issues/${encodeURIComponent(id)}`);
}

export interface IssueTimelineParams {
  limit?: number;
  offset?: number;
}

export async function getIssueTimeline(
  id: string,
  params?: IssueTimelineParams,
) {
  const query = new URLSearchParams();
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.offset) query.set("offset", String(params.offset));
  const qs = query.toString();

  return apiGet<IssueTimelineEvent[]>(
    `/api/v1/issues/${encodeURIComponent(id)}/timeline${qs ? `?${qs}` : ""}`,
  );
}
