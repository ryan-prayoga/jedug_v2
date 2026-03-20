const TOKEN_KEY = "jedug_anon_token";
const CONSENT_KEY = "jedug_terms_accepted";
const ISSUE_FOLLOWER_ID_KEY = "jedug_issue_follower_id";
const ISSUE_FOLLOWER_TOKEN_KEY = "jedug_issue_follower_token";
const ISSUE_FOLLOWER_TOKEN_EXP_KEY = "jedug_issue_follower_token_exp";
const ISSUE_FOLLOWER_STREAM_TOKEN_KEY = "jedug_issue_follower_stream_token";
const ISSUE_FOLLOWER_STREAM_TOKEN_EXP_KEY =
  "jedug_issue_follower_stream_token_exp";
const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export function getAnonToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setAnonToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearAnonToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

export function isConsentGiven(): boolean {
  if (typeof window === "undefined") return false;
  return localStorage.getItem(CONSENT_KEY) === "true";
}

export function setConsentGiven(): void {
  localStorage.setItem(CONSENT_KEY, "true");
}

export function clearConsentGiven(): void {
  localStorage.removeItem(CONSENT_KEY);
}

export function getIssueFollowerId(): string | null {
  if (typeof window === "undefined") return null;

  const followerId = localStorage.getItem(ISSUE_FOLLOWER_ID_KEY);
  if (!followerId || !UUID_PATTERN.test(followerId)) {
    return null;
  }

  return followerId;
}

export function getOrCreateIssueFollowerId(): string | null {
  if (typeof window === "undefined") return null;

  const existing = getIssueFollowerId();
  if (existing) {
    return existing;
  }

  const followerId = generateBrowserUUID();
  clearIssueFollowerAuthToken();
  localStorage.setItem(ISSUE_FOLLOWER_ID_KEY, followerId);
  return followerId;
}

export function getIssueFollowerAuthToken(): string | null {
  if (typeof window === "undefined") return null;

  const token = localStorage.getItem(ISSUE_FOLLOWER_TOKEN_KEY);
  const expiresAt = localStorage.getItem(ISSUE_FOLLOWER_TOKEN_EXP_KEY);
  if (!token || !expiresAt) {
    return null;
  }

  const expiresAtMs = Date.parse(expiresAt);
  if (Number.isNaN(expiresAtMs) || expiresAtMs <= Date.now() + 30_000) {
    clearIssueFollowerAuthToken();
    return null;
  }

  return token;
}

export function getIssueFollowerStreamAuthToken(): string | null {
  if (typeof window === "undefined") return null;

  const token = localStorage.getItem(ISSUE_FOLLOWER_STREAM_TOKEN_KEY);
  const expiresAt = localStorage.getItem(ISSUE_FOLLOWER_STREAM_TOKEN_EXP_KEY);
  if (!token || !expiresAt) {
    return null;
  }

  const expiresAtMs = Date.parse(expiresAt);
  if (Number.isNaN(expiresAtMs) || expiresAtMs <= Date.now() + 30_000) {
    clearIssueFollowerStreamAuthToken();
    return null;
  }

  return token;
}

export function setIssueFollowerAuthToken(
  token: string,
  expiresAt: string,
): void {
  localStorage.setItem(ISSUE_FOLLOWER_TOKEN_KEY, token);
  localStorage.setItem(ISSUE_FOLLOWER_TOKEN_EXP_KEY, expiresAt);
}

export function setIssueFollowerStreamAuthToken(
  token: string,
  expiresAt: string,
): void {
  localStorage.setItem(ISSUE_FOLLOWER_STREAM_TOKEN_KEY, token);
  localStorage.setItem(ISSUE_FOLLOWER_STREAM_TOKEN_EXP_KEY, expiresAt);
}

export function clearIssueFollowerAuthToken(): void {
  localStorage.removeItem(ISSUE_FOLLOWER_TOKEN_KEY);
  localStorage.removeItem(ISSUE_FOLLOWER_TOKEN_EXP_KEY);
  clearIssueFollowerStreamAuthToken();
}

export function clearIssueFollowerStreamAuthToken(): void {
  localStorage.removeItem(ISSUE_FOLLOWER_STREAM_TOKEN_KEY);
  localStorage.removeItem(ISSUE_FOLLOWER_STREAM_TOKEN_EXP_KEY);
}

export function clearIssueFollowerIdentity(): void {
  localStorage.removeItem(ISSUE_FOLLOWER_ID_KEY);
  clearIssueFollowerAuthToken();
}

export function resetAnonymousBrowserIdentity(): void {
  clearAnonToken();
  clearConsentGiven();
  clearIssueFollowerIdentity();
}

function generateBrowserUUID(): string {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return crypto.randomUUID();
  }

  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (char) => {
    const random = Math.floor(Math.random() * 16);
    const value = char === "x" ? random : (random & 0x3) | 0x8;
    return value.toString(16);
  });
}
