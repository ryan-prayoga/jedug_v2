import { PUBLIC_API_BASE_URL } from "$env/static/public";
import type { ApiResponse } from "./types";

export class ApiError extends Error {
  status: number;
  errorCode?: string;
  constructor(message: string, status: number, errorCode?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.errorCode = errorCode;
  }
}

async function parseResponse<T>(res: Response): Promise<ApiResponse<T>> {
  const json: ApiResponse<T> = await res.json();
  if (!json.success) {
    throw new ApiError(
      json.message || "Request failed",
      res.status,
      json.error_code,
    );
  }
  return json;
}

function resolveUrl(path: string): string {
  if (/^https?:\/\//i.test(path)) {
    return path;
  }
  return `${PUBLIC_API_BASE_URL}${path}`;
}

interface ApiRequestOptions {
  deviceToken?: string;
  headers?: HeadersInit;
}

function normalizeRequestOptions(
  options?: string | ApiRequestOptions,
): ApiRequestOptions {
  if (typeof options === "string") {
    return { deviceToken: options };
  }
  return options ?? {};
}

function getHeaders(
  options?: string | ApiRequestOptions,
  includeJSON = true,
): Headers {
  const resolved = normalizeRequestOptions(options);
  const headers = new Headers(resolved.headers);
  if (includeJSON && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (resolved.deviceToken) {
    headers.set("X-Device-Token", resolved.deviceToken);
  }
  return headers;
}

export async function apiGet<T>(
  path: string,
  options?: string | ApiRequestOptions,
): Promise<ApiResponse<T>> {
  const res = await fetch(resolveUrl(path), {
    method: "GET",
    headers: getHeaders(options),
  });
  return parseResponse<T>(res);
}

export async function apiPost<T>(
  path: string,
  body?: unknown,
  options?: string | ApiRequestOptions,
): Promise<ApiResponse<T>> {
  const res = await fetch(resolveUrl(path), {
    method: "POST",
    headers: getHeaders(options),
    body: body ? JSON.stringify(body) : undefined,
  });
  return parseResponse<T>(res);
}

export async function apiDelete<T>(
  path: string,
  body?: unknown,
  options?: string | ApiRequestOptions,
): Promise<ApiResponse<T>> {
  const res = await fetch(resolveUrl(path), {
    method: "DELETE",
    headers: getHeaders(options),
    body: body ? JSON.stringify(body) : undefined,
  });
  return parseResponse<T>(res);
}

export async function apiPatch<T>(
  path: string,
  body?: unknown,
  options?: string | ApiRequestOptions,
): Promise<ApiResponse<T>> {
  const res = await fetch(resolveUrl(path), {
    method: "PATCH",
    headers: getHeaders(options),
    body: body ? JSON.stringify(body) : undefined,
  });
  return parseResponse<T>(res);
}

export async function apiUploadBinary(
  path: string,
  file: Blob,
  contentType: string,
  method = "POST",
  extraHeaders: Record<string, string> = {},
): Promise<ApiResponse<undefined>> {
  const headers: HeadersInit = {
    ...extraHeaders,
  };
  if (!("Content-Type" in headers) && !("content-type" in headers)) {
    headers["Content-Type"] = contentType;
  }

  const res = await fetch(resolveUrl(path), {
    method,
    headers,
    body: file,
  });

  if (!res.ok) {
    const responseType = res.headers.get("content-type") || "";
    if (responseType.includes("application/json")) {
      const json: ApiResponse<undefined> = await res.json();
      throw new ApiError(json.message || "Upload failed", res.status);
    }

    const message = (await res.text()).trim() || "Upload failed";
    throw new ApiError(message, res.status);
  }

  const responseType = res.headers.get("content-type") || "";
  if (responseType.includes("application/json")) {
    return parseResponse<undefined>(res);
  }

  return { success: true };
}
