const TOKEN_KEY = "jedug_anon_token";
const CONSENT_KEY = "jedug_terms_accepted";

export function getAnonToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setAnonToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function isConsentGiven(): boolean {
  if (typeof window === "undefined") return false;
  return localStorage.getItem(CONSENT_KEY) === "true";
}

export function setConsentGiven(): void {
  localStorage.setItem(CONSENT_KEY, "true");
}
