import { browser } from "$app/environment";
import { goto } from "$app/navigation";
import { writable } from "svelte/store";
import {
  getBrowserPushStatus,
  subscribeBrowserPush,
  unsubscribeBrowserPush,
  type BrowserPushStatusResponse,
} from "$lib/api/push";
import { requestIssueDetailRefresh } from "$lib/utils/issue-detail-refresh";
import {
  ensureFollowerAuthToken,
  getFollowerAuthProblem,
} from "$lib/utils/follower-auth";
import { getOrCreateIssueFollowerId } from "$lib/utils/storage";

export type BrowserPushStatus =
  | "ios_browser_tab"
  | "unsupported"
  | "default"
  | "granted"
  | "denied"
  | "subscribed";

type PushPermission = NotificationPermission | "unsupported";

interface BrowserPushState {
  status: BrowserPushStatus;
  supported: boolean;
  permission: PushPermission;
  isIOS: boolean;
  isStandalone: boolean;
  requiresHomeScreen: boolean;
  needsFollowerRebind: boolean;
  followerAuthMessage: string | null;
  enabled: boolean;
  subscribed: boolean;
  subscriptionCount: number;
  followerID: string | null;
  followerToken: string | null;
  loading: boolean;
  busy: boolean;
  initialized: boolean;
  error: string | null;
  success: string | null;
  vapidPublicKey: string | null;
}

type PushClientMessage =
  | {
      type: "jedug:push-open-issue";
      issue_id?: string;
      url?: string;
    }
  | {
      type: "jedug:push-received";
      payload?: {
        issue_id?: string;
        url?: string;
      };
    };

const initialState: BrowserPushState = {
  status: "default",
  supported: false,
  permission: "default",
  isIOS: false,
  isStandalone: false,
  requiresHomeScreen: false,
  needsFollowerRebind: false,
  followerAuthMessage: null,
  enabled: false,
  subscribed: false,
  subscriptionCount: 0,
  followerID: null,
  followerToken: null,
  loading: false,
  busy: false,
  initialized: false,
  error: null,
  success: null,
  vapidPublicKey: null,
};

const state = writable<BrowserPushState>(initialState);

let _registrationPromise: Promise<ServiceWorkerRegistration> | null = null;
let _messageBridgeBound = false;

function isSupported(): boolean {
  return (
    browser &&
    "Notification" in window &&
    "serviceWorker" in navigator &&
    "PushManager" in window
  );
}

function detectIOS(): boolean {
  if (!browser) return false;

  const userAgent = navigator.userAgent ?? "";
  const platform = navigator.platform ?? "";

  return (
    /iPad|iPhone|iPod/.test(userAgent) ||
    (platform === "MacIntel" && navigator.maxTouchPoints > 1)
  );
}

function detectStandaloneMode(): boolean {
  if (!browser) return false;

  const navigatorStandalone =
    "standalone" in navigator &&
    Boolean((navigator as Navigator & { standalone?: boolean }).standalone);
  const displayModeStandalone =
    window.matchMedia("(display-mode: standalone)").matches ||
    window.matchMedia("(display-mode: fullscreen)").matches ||
    window.matchMedia("(display-mode: minimal-ui)").matches;

  return navigatorStandalone || displayModeStandalone;
}

function deriveStatus(
  requiresHomeScreen: boolean,
  supported: boolean,
  permission: PushPermission,
  subscribed: boolean,
): BrowserPushStatus {
  if (requiresHomeScreen) return "ios_browser_tab";
  if (!supported) return "unsupported";
  if (permission === "denied") return "denied";
  if (subscribed) return "subscribed";
  if (permission === "granted") return "granted";
  return "default";
}

async function ensureServiceWorkerRegistration() {
  if (!browser) {
    throw new Error("service worker only available in browser");
  }

  if (_registrationPromise) {
    return _registrationPromise;
  }

  _registrationPromise = navigator.serviceWorker.register("/sw.js", {
    scope: "/",
  });
  return _registrationPromise;
}

function urlBase64ToUint8Array(base64String: string): ArrayBuffer {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const normalized = (base64String + padding)
    .replace(/-/g, "+")
    .replace(/_/g, "/");
  const rawData = window.atob(normalized);
  const output = new Uint8Array(rawData.length);

  for (let index = 0; index < rawData.length; index += 1) {
    output[index] = rawData.charCodeAt(index);
  }

  return output.buffer.slice(0);
}

function toSubscriptionPayload(subscription: PushSubscription) {
  const p256dh = subscription.getKey("p256dh");
  const auth = subscription.getKey("auth");
  if (!p256dh || !auth) {
    throw new Error("Kunci subscription browser tidak lengkap.");
  }

  return {
    endpoint: subscription.endpoint,
    keys: {
      p256dh: arrayBufferToBase64(p256dh),
      auth: arrayBufferToBase64(auth),
    },
  };
}

function arrayBufferToBase64(buffer: ArrayBuffer): string {
  let binary = "";
  const bytes = new Uint8Array(buffer);
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary);
}

async function loadSnapshot(followerID: string, followerToken: string | null) {
  const isIOS = detectIOS();
  const isStandalone = detectStandaloneMode();
  const requiresHomeScreen = isIOS && !isStandalone;
  const supported = isSupported();
  const permission: PushPermission = supported
    ? Notification.permission
    : "unsupported";

  if (requiresHomeScreen || !supported) {
    return {
      isIOS,
      isStandalone,
      requiresHomeScreen,
      supported,
      permission,
      enabled: false,
      subscriptionCount: 0,
      vapidPublicKey: null,
      localSubscription: null as PushSubscription | null,
    };
  }

  let backendStatus: BrowserPushStatusResponse | null = null;
  if (followerToken) {
    try {
      const result = await getBrowserPushStatus(followerToken);
      backendStatus = result.data ?? null;
    } catch {
      backendStatus = null;
    }
  }

  let localSubscription: PushSubscription | null = null;
  if (permission === "granted") {
    const registration = await ensureServiceWorkerRegistration();
    localSubscription = await registration.pushManager.getSubscription();
  }

  return {
    isIOS,
    isStandalone,
    requiresHomeScreen,
    supported,
    permission,
    enabled: backendStatus?.enabled ?? false,
    subscriptionCount: backendStatus?.subscription_count ?? 0,
    vapidPublicKey: backendStatus?.vapid_public_key ?? null,
    localSubscription,
  };
}

function bindMessageBridge() {
  if (!browser || _messageBridgeBound || !("serviceWorker" in navigator)) {
    return;
  }

  navigator.serviceWorker.addEventListener("message", (event) => {
    const data = event.data as PushClientMessage | undefined;
    if (!data || typeof data !== "object") return;

    if (data.type === "jedug:push-received") {
      const issueID = data.payload?.issue_id;
      const url = data.payload?.url;
      if (!issueID || !url) return;

      const target = new URL(url, window.location.origin);
      if (target.pathname === window.location.pathname) {
        requestIssueDetailRefresh({ issueID, source: "notification" });
      }
      return;
    }

    if (data.type === "jedug:push-open-issue") {
      const issueID = data.issue_id;
      const url = data.url;
      if (!issueID || !url) return;

      const target = new URL(url, window.location.origin);
      const targetPath = `${target.pathname}${target.search}${target.hash}`;
      if (target.pathname === window.location.pathname) {
        requestIssueDetailRefresh({ issueID, source: "notification" });
        return;
      }

      void goto(targetPath);
    }
  });

  _messageBridgeBound = true;
}

async function syncExistingSubscription(
  followerID: string,
  followerToken: string | null,
  snapshot: Awaited<ReturnType<typeof loadSnapshot>>,
) {
  if (!snapshot.enabled || !snapshot.localSubscription || !followerToken) {
    return snapshot;
  }

  if (snapshot.subscriptionCount > 0) {
    return snapshot;
  }

  await subscribeBrowserPush(
    followerToken,
    toSubscriptionPayload(snapshot.localSubscription),
  );
  return loadSnapshot(followerID, followerToken);
}

async function refreshState(showLoading: boolean) {
  if (!browser) return;

  const followerID = getOrCreateIssueFollowerId();
  if (!followerID) {
    state.update((prev) => ({
      ...prev,
      supported: false,
      status: "unsupported",
      permission: "unsupported",
      isIOS: false,
      isStandalone: false,
      requiresHomeScreen: false,
      needsFollowerRebind: false,
      followerAuthMessage: null,
      followerID: null,
      loading: false,
      busy: false,
    }));
    return;
  }

  const followerToken = await ensureFollowerAuthToken();
  const authProblem = getFollowerAuthProblem();

  if (showLoading) {
    state.update((prev) => ({
      ...prev,
      loading: true,
      error: null,
      success: null,
      followerID,
      followerToken,
      needsFollowerRebind: authProblem.code === "binding_reset_required",
      followerAuthMessage: authProblem.message,
    }));
  }

  try {
    let snapshot = await loadSnapshot(followerID, followerToken);
    snapshot = await syncExistingSubscription(
      followerID,
      followerToken,
      snapshot,
    );
    bindMessageBridge();

    const subscribed = Boolean(snapshot.localSubscription);

    state.update((prev) => ({
      ...prev,
      supported: snapshot.supported,
      permission: snapshot.permission,
      isIOS: snapshot.isIOS,
      isStandalone: snapshot.isStandalone,
      requiresHomeScreen: snapshot.requiresHomeScreen,
      needsFollowerRebind: authProblem.code === "binding_reset_required",
      followerAuthMessage: authProblem.message,
      enabled: snapshot.enabled,
      subscribed,
      subscriptionCount: snapshot.subscriptionCount,
      vapidPublicKey: snapshot.vapidPublicKey,
      followerID,
      followerToken,
      loading: false,
      busy: false,
      initialized: true,
      status: deriveStatus(
        snapshot.requiresHomeScreen,
        snapshot.supported,
        snapshot.permission,
        subscribed,
      ),
      error: null,
    }));
  } catch {
    state.update((prev) => ({
      ...prev,
      loading: false,
      busy: false,
      initialized: true,
      followerID,
      followerToken,
      error: "Belum bisa memeriksa status notifikasi browser.",
    }));
  }
}

export const browserPushState = {
  subscribe: state.subscribe,

  async init() {
    if (!browser) return;

    let shouldInit = false;
    state.update((prev) => {
      if (prev.initialized) {
        return prev;
      }
      shouldInit = true;
      return { ...prev, initialized: true };
    });

    if (!shouldInit) return;
    await refreshState(true);
  },

  async refresh() {
    await refreshState(false);
  },

  async enable() {
    if (!browser) return false;

    if (!isSupported()) {
      state.update((prev) => ({
        ...prev,
        isIOS: detectIOS(),
        isStandalone: detectStandaloneMode(),
        requiresHomeScreen: false,
        needsFollowerRebind: false,
        followerAuthMessage: null,
        supported: false,
        permission: "unsupported",
        status: "unsupported",
        error: "Browser ini belum mendukung notifikasi browser.",
      }));
      return false;
    }

    const isIOS = detectIOS();
    const isStandalone = detectStandaloneMode();
    if (isIOS && !isStandalone) {
      state.update((prev) => ({
        ...prev,
        isIOS,
        isStandalone,
        requiresHomeScreen: true,
        busy: false,
        status: "ios_browser_tab",
        error:
          "Di iPhone, notifikasi browser JEDUG hanya bisa aktif jika web app ini dibuka dari Home Screen.",
      }));
      return false;
    }

    const followerID = getOrCreateIssueFollowerId();
    if (!followerID) {
      state.update((prev) => ({
        ...prev,
        error: "Identitas browser belum siap. Coba muat ulang halaman.",
      }));
      return false;
    }

    const followerToken = await ensureFollowerAuthToken({ forceRefresh: true });
    if (!followerToken) {
      const authProblem = getFollowerAuthProblem();
      state.update((prev) => ({
        ...prev,
        busy: false,
        needsFollowerRebind: authProblem.code === "binding_reset_required",
        followerAuthMessage: authProblem.message,
        error:
          authProblem.message ??
          "Ikuti setidaknya satu laporan dari browser ini dulu agar notifikasi browser bisa diamankan.",
      }));
      return false;
    }

    state.update((prev) => ({
      ...prev,
      busy: true,
      needsFollowerRebind: false,
      followerAuthMessage: null,
      error: null,
      success: null,
      followerID,
      followerToken,
    }));

    try {
      let permission = Notification.permission;
      if (permission === "default") {
        permission = await Notification.requestPermission();
      }

      if (permission !== "granted") {
        state.update((prev) => ({
          ...prev,
          busy: false,
          permission,
          status: deriveStatus(false, true, permission, false),
          error:
            permission === "denied"
              ? "Izin notifikasi ditolak. Ubah pengaturan browser jika ingin mengaktifkannya lagi."
              : null,
        }));
        return false;
      }

      const statusResult = await getBrowserPushStatus(followerToken);
      const backendStatus = statusResult.data;
      if (!backendStatus?.enabled || !backendStatus.vapid_public_key) {
        state.update((prev) => ({
          ...prev,
          busy: false,
          permission,
          enabled: backendStatus?.enabled ?? false,
          status: "granted",
          error: "Browser push belum siap di server. Coba lagi beberapa saat.",
        }));
        return false;
      }

      const registration = await ensureServiceWorkerRegistration();
      let subscription = await registration.pushManager.getSubscription();

      if (!subscription) {
        subscription = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(
            backendStatus.vapid_public_key,
          ),
        });
      }

      await subscribeBrowserPush(
        followerToken,
        toSubscriptionPayload(subscription),
      );

      state.update((prev) => ({
        ...prev,
        busy: false,
        requiresHomeScreen: false,
        enabled: true,
        subscribed: true,
        subscriptionCount: Math.max(prev.subscriptionCount, 1),
        permission,
        status: "subscribed",
        vapidPublicKey: backendStatus.vapid_public_key ?? null,
        error: null,
        success: "Notifikasi browser berhasil diaktifkan di perangkat ini.",
      }));

      return true;
    } catch {
      state.update((prev) => ({
        ...prev,
        busy: false,
        error: "Subscription notifikasi browser gagal dibuat. Coba lagi.",
      }));
      await refreshState(false);
      return false;
    }
  },

  async disable() {
    if (!browser || !isSupported()) return false;

    let followerToken: string | null = null;
    state.update((prev) => {
      followerToken = prev.followerToken;
      return {
        ...prev,
        busy: true,
        error: null,
        success: null,
      };
    });

    if (!followerToken) {
      followerToken = await ensureFollowerAuthToken({ forceRefresh: true });
    }

    try {
      const registration = await ensureServiceWorkerRegistration();
      const subscription = await registration.pushManager.getSubscription();

      if (subscription && followerToken) {
        await unsubscribeBrowserPush(followerToken, subscription.endpoint);
        await subscription.unsubscribe();
      }

      const permission: PushPermission = Notification.permission;
      state.update((prev) => ({
        ...prev,
        busy: false,
        subscribed: false,
        subscriptionCount: 0,
        permission,
        status: deriveStatus(false, true, permission, false),
        success: "Notifikasi browser dimatikan di perangkat ini.",
      }));

      await refreshState(false);
      return true;
    } catch {
      state.update((prev) => ({
        ...prev,
        busy: false,
        error: "Belum bisa mematikan notifikasi browser. Coba lagi.",
      }));
      return false;
    }
  },
};
