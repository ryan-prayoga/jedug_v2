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

function getHeaders(token?: string): HeadersInit {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["X-Device-Token"] = token;
  }
  return headers;
}

export async function apiGet<T>(
  path: string,
  token?: string,
): Promise<ApiResponse<T>> {
  const res = await fetch(resolveUrl(path), {
    method: "GET",
    headers: getHeaders(token),
  });
  return parseResponse<T>(res);
}

export async function apiPost<T>(
  path: string,
  body?: unknown,
  token?: string,
): Promise<ApiResponse<T>> {
  const res = await fetch(resolveUrl(path), {
    method: "POST",
    headers: getHeaders(token),
    body: body ? JSON.stringify(body) : undefined,
  });
  return parseResponse<T>(res);
}

export async function apiDelete<T>(
  path: string,
  body?: unknown,
  token?: string,
): Promise<ApiResponse<T>> {
  const res = await fetch(resolveUrl(path), {
    method: "DELETE",
    headers: getHeaders(token),
    body: body ? JSON.stringify(body) : undefined,
  });
  return parseResponse<T>(res);
}

export async function apiPatch<T>(
  path: string,
  body?: unknown,
  token?: string,
): Promise<ApiResponse<T>> {
  const res = await fetch(resolveUrl(path), {
    method: "PATCH",
    headers: getHeaders(token),
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
