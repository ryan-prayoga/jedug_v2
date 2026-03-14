import { ApiError, apiDelete, apiGet, apiPost } from "./client";
import type { ApiResponse } from "./types";
import type {
  Issue,
  IssueDetail,
  IssueFollowState,
  IssueFollowersCount,
  IssueTimelineEvent,
} from "./types";

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

async function with404Fallback<T>(
  primary: () => Promise<ApiResponse<T>>,
  fallback: () => Promise<ApiResponse<T>>,
) {
  try {
    return await primary();
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return fallback();
    }

    throw error;
  }
}

export async function followIssue(id: string, followerId: string) {
  const encodedId = encodeURIComponent(id);

  return with404Fallback(
    () =>
      apiPost<IssueFollowState>(`/api/v1/issues/${encodedId}/follow`, {
        follower_id: followerId,
      }),
    () =>
      apiPost<IssueFollowState>(`/api/v1/issues/${encodedId}/followers`, {
        follower_id: followerId,
      }),
  );
}

export async function unfollowIssue(id: string, followerId: string) {
  const encodedId = encodeURIComponent(id);

  return with404Fallback(
    () =>
      apiDelete<IssueFollowState>(`/api/v1/issues/${encodedId}/follow`, {
        follower_id: followerId,
      }),
    () =>
      apiDelete<IssueFollowState>(
        `/api/v1/issues/${encodedId}/followers?follower_id=${encodeURIComponent(followerId)}`,
      ),
  );
}

export async function getIssueFollowerCount(id: string) {
  const encodedId = encodeURIComponent(id);

  return with404Fallback(
    () =>
      apiGet<IssueFollowersCount>(
        `/api/v1/issues/${encodedId}/followers/count`,
      ),
    () => apiGet<IssueFollowersCount>(`/api/v1/issues/${encodedId}/count`),
  );
}

export async function getIssueFollowStatus(id: string, followerId: string) {
  const query = new URLSearchParams({ follower_id: followerId });
  const encodedId = encodeURIComponent(id);

  return with404Fallback(
    () =>
      apiGet<IssueFollowState>(
        `/api/v1/issues/${encodedId}/follow-status?${query.toString()}`,
      ),
    () =>
      apiGet<IssueFollowState>(
        `/api/v1/issues/${encodedId}/follow/status?${query.toString()}`,
      ),
  );
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
