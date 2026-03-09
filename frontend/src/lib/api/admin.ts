import { PUBLIC_API_BASE_URL } from "$env/static/public";
import type {
  ApiResponse,
  AdminIssue,
  AdminIssueDetail,
  AdminLoginResponse,
  AdminMeResponse,
} from "./types";
import { ApiError } from "./client";
import { getAdminToken } from "$lib/utils/storage";

async function parseResponse<T>(res: Response): Promise<ApiResponse<T>> {
  const json: ApiResponse<T> = await res.json();
  if (!json.success) {
    throw new ApiError(json.message || "Request failed", res.status);
  }
  return json;
}

function adminHeaders(withAuth: boolean): HeadersInit {
  const headers: HeadersInit = { "Content-Type": "application/json" };
  if (withAuth) {
    const token = getAdminToken();
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
  }
  return headers;
}

async function adminGet<T>(path: string): Promise<ApiResponse<T>> {
  const res = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
    method: "GET",
    headers: adminHeaders(true),
  });
  return parseResponse<T>(res);
}

async function adminPost<T>(
  path: string,
  body?: unknown,
): Promise<ApiResponse<T>> {
  const res = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
    method: "POST",
    headers: adminHeaders(true),
    body: body ? JSON.stringify(body) : undefined,
  });
  return parseResponse<T>(res);
}

export async function adminLogin(username: string, password: string) {
  // Login does NOT send Bearer auth (no token yet)
  const res = await fetch(`${PUBLIC_API_BASE_URL}/api/v1/admin/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  return parseResponse<AdminLoginResponse>(res);
}

export async function adminMe() {
  return adminGet<AdminMeResponse>("/api/v1/admin/me");
}

export async function adminListIssues(params?: {
  limit?: number;
  offset?: number;
  status?: string;
}) {
  const query = new URLSearchParams();
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.offset) query.set("offset", String(params.offset));
  if (params?.status) query.set("status", params.status);
  const qs = query.toString();
  return adminGet<AdminIssue[]>(`/api/v1/admin/issues${qs ? `?${qs}` : ""}`);
}

export async function adminGetIssue(id: string) {
  return adminGet<AdminIssueDetail>(
    `/api/v1/admin/issues/${encodeURIComponent(id)}`,
  );
}

export async function adminHideIssue(id: string, reason?: string) {
  return adminPost<undefined>(
    `/api/v1/admin/issues/${encodeURIComponent(id)}/hide`,
    reason ? { reason } : undefined,
  );
}

export async function adminFixIssue(id: string, reason?: string) {
  return adminPost<undefined>(
    `/api/v1/admin/issues/${encodeURIComponent(id)}/fix`,
    reason ? { reason } : undefined,
  );
}

export async function adminRejectIssue(id: string, reason?: string) {
  return adminPost<undefined>(
    `/api/v1/admin/issues/${encodeURIComponent(id)}/reject`,
    reason ? { reason } : undefined,
  );
}

export async function adminUnhideIssue(id: string, reason?: string) {
  return adminPost<undefined>(
    `/api/v1/admin/issues/${encodeURIComponent(id)}/unhide`,
    reason ? { reason } : undefined,
  );
}

export async function adminBanDevice(id: string, reason?: string) {
  return adminPost<undefined>(
    `/api/v1/admin/devices/${encodeURIComponent(id)}/ban`,
    reason ? { reason } : undefined,
  );
}
