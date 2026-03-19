import { ApiError } from "$lib/api/client";
import { issueFollowerAuthToken } from "$lib/api/follower-auth";
import {
  clearIssueFollowerAuthToken,
  getAnonToken,
  getIssueFollowerAuthToken,
  getIssueFollowerStreamAuthToken,
  getOrCreateIssueFollowerId,
  setIssueFollowerAuthToken,
  setIssueFollowerStreamAuthToken,
} from "$lib/utils/storage";

type RefreshedFollowerAuthTokens = {
  accessToken: string | null;
  streamToken: string | null;
};

let refreshPromise: Promise<RefreshedFollowerAuthTokens | null> | null = null;
let followerAuthProblem: {
  code: "binding_reset_required" | null;
  message: string | null;
} = {
  code: null,
  message: null,
};

function clearFollowerAuthProblem() {
  followerAuthProblem = {
    code: null,
    message: null,
  };
}

function setFollowerAuthProblem(
  code: "binding_reset_required",
  message: string,
) {
  followerAuthProblem = { code, message };
}

export function getFollowerAuthProblem() {
  return followerAuthProblem;
}

export function getFollowerAuthUnavailableMessage(fallback: string): string {
  return followerAuthProblem.message ?? fallback;
}

export async function ensureFollowerAuthToken(
  options: { forceRefresh?: boolean } = {},
): Promise<string | null> {
  const { forceRefresh = false } = options;
  const followerID = getOrCreateIssueFollowerId();
  if (!followerID) return null;

  if (!forceRefresh) {
    const existing = getIssueFollowerAuthToken();
    if (existing) {
      return existing;
    }
  }

  if (refreshPromise) {
    const refreshed = await refreshPromise;
    return refreshed?.accessToken ?? null;
  }

  refreshPromise = refreshFollowerAuthTokens(followerID);

  try {
    const refreshed = await refreshPromise;
    return refreshed?.accessToken ?? null;
  } finally {
    refreshPromise = null;
  }
}

export async function ensureFollowerStreamToken(
  options: { forceRefresh?: boolean } = {},
): Promise<string | null> {
  const { forceRefresh = false } = options;
  const followerID = getOrCreateIssueFollowerId();
  if (!followerID) return null;

  if (!forceRefresh) {
    const existing = getIssueFollowerStreamAuthToken();
    if (existing) {
      return existing;
    }
  }

  if (refreshPromise) {
    const refreshed = await refreshPromise;
    return refreshed?.streamToken ?? null;
  }

  refreshPromise = refreshFollowerAuthTokens(followerID);

  try {
    const refreshed = await refreshPromise;
    return refreshed?.streamToken ?? null;
  } finally {
    refreshPromise = null;
  }
}

async function refreshFollowerAuthTokens(
  followerID: string,
): Promise<RefreshedFollowerAuthTokens | null> {
  const anonToken = getAnonToken();
  if (!anonToken) {
    clearIssueFollowerAuthToken();
    return null;
  }

  try {
    const result = await issueFollowerAuthToken(followerID, anonToken);
    const token = result.data?.follower_token ?? null;
    const expiresAt = result.data?.expires_at ?? null;
    const streamToken = result.data?.stream_token ?? null;
    const streamExpiresAt = result.data?.stream_expires_at ?? null;
    if (!token || !expiresAt || !streamToken || !streamExpiresAt) {
      clearIssueFollowerAuthToken();
      return null;
    }

    clearFollowerAuthProblem();
    setIssueFollowerAuthToken(token, expiresAt);
    setIssueFollowerStreamAuthToken(streamToken, streamExpiresAt);
    return {
      accessToken: token,
      streamToken,
    };
  } catch (error) {
    if (
      error instanceof ApiError &&
      error.errorCode === "follower_binding_not_found"
    ) {
      clearIssueFollowerAuthToken();
      setFollowerAuthProblem(
        "binding_reset_required",
        "Sesi notifikasi browser ini sudah tidak cocok dengan data server. Reset browser ini, setujui ulang JEDUG, lalu aktifkan notifikasi lagi.",
      );
      return null;
    }

    if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
      clearIssueFollowerAuthToken();
      return null;
    }
    throw error;
  }
}

export function persistFollowerAuthFromIssueState(
  state:
    | {
        follower_token?: string;
        follower_token_expires_at?: string;
        follower_stream_token?: string;
        follower_stream_token_expires_at?: string;
      }
    | null
    | undefined,
): void {
  const token = state?.follower_token;
  const expiresAt = state?.follower_token_expires_at;
  const streamToken = state?.follower_stream_token;
  const streamExpiresAt = state?.follower_stream_token_expires_at;
  if (token && expiresAt) {
    setIssueFollowerAuthToken(token, expiresAt);
  }
  if (streamToken && streamExpiresAt) {
    setIssueFollowerStreamAuthToken(streamToken, streamExpiresAt);
  }
}
