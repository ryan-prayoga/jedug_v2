import { PUBLIC_API_BASE_URL } from "$env/static/public";
import type { ApiResponse } from "./types";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function parseResponse<T>(res: Response): Promise<ApiResponse<T>> {
  const json: ApiResponse<T> = await res.json();
  if (!json.success) {
    throw new ApiError(json.message || "Request failed", res.status);
  }
  return json;
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
  const res = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
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
  const res = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
    method: "POST",
    headers: getHeaders(token),
    body: body ? JSON.stringify(body) : undefined,
  });
  return parseResponse<T>(res);
}

export async function apiUploadBinary(
  path: string,
  file: Blob,
  contentType: string,
): Promise<ApiResponse<undefined>> {
  const res = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
    method: "POST",
    headers: {
      "Content-Type": contentType,
    },
    body: file,
  });
  return parseResponse<undefined>(res);
}
