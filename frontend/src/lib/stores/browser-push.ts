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
import { getOrCreateIssueFollowerId } from "$lib/utils/storage";

export type BrowserPushStatus =
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
  enabled: boolean;
  subscribed: boolean;
  subscriptionCount: number;
  followerID: string | null;
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
  enabled: false,
  subscribed: false,
  subscriptionCount: 0,
  followerID: null,
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

function deriveStatus(
  supported: boolean,
  permission: PushPermission,
  subscribed: boolean,
): BrowserPushStatus {
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
  const normalized = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
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

async function loadSnapshot(followerID: string) {
  const supported = isSupported();
  const permission: PushPermission = supported
    ? Notification.permission
    : "unsupported";

  if (!supported) {
    return {
      supported,
      permission,
      enabled: false,
      subscriptionCount: 0,
      vapidPublicKey: null,
      localSubscription: null as PushSubscription | null,
    };
  }

  let backendStatus: BrowserPushStatusResponse | null = null;
  try {
    const result = await getBrowserPushStatus(followerID);
    backendStatus = result.data ?? null;
  } catch {
    backendStatus = null;
  }

  let localSubscription: PushSubscription | null = null;
  if (permission === "granted") {
    const registration = await ensureServiceWorkerRegistration();
    localSubscription = await registration.pushManager.getSubscription();
  }

  return {
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
  snapshot: Awaited<ReturnType<typeof loadSnapshot>>,
) {
  if (!snapshot.enabled || !snapshot.localSubscription) {
    return snapshot;
  }

  if (snapshot.subscriptionCount > 0) {
    return snapshot;
  }

  await subscribeBrowserPush(
    followerID,
    toSubscriptionPayload(snapshot.localSubscription),
  );
  return loadSnapshot(followerID);
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
      followerID: null,
      loading: false,
      busy: false,
    }));
    return;
  }

  if (showLoading) {
    state.update((prev) => ({
      ...prev,
      loading: true,
      error: null,
      success: null,
      followerID,
    }));
  }

  try {
    let snapshot = await loadSnapshot(followerID);
    snapshot = await syncExistingSubscription(followerID, snapshot);
    bindMessageBridge();

    const subscribed = Boolean(snapshot.localSubscription);

    state.update((prev) => ({
      ...prev,
      supported: snapshot.supported,
      permission: snapshot.permission,
      enabled: snapshot.enabled,
      subscribed,
      subscriptionCount: snapshot.subscriptionCount,
      vapidPublicKey: snapshot.vapidPublicKey,
      followerID,
      loading: false,
      busy: false,
      initialized: true,
      status: deriveStatus(snapshot.supported, snapshot.permission, subscribed),
      error: null,
    }));
  } catch {
    state.update((prev) => ({
      ...prev,
      loading: false,
      busy: false,
      initialized: true,
      followerID,
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
        supported: false,
        permission: "unsupported",
        status: "unsupported",
        error: "Browser ini belum mendukung notifikasi browser.",
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

    state.update((prev) => ({
      ...prev,
      busy: true,
      error: null,
      success: null,
      followerID,
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
          status: deriveStatus(true, permission, false),
          error:
            permission === "denied"
              ? "Izin notifikasi ditolak. Ubah pengaturan browser jika ingin mengaktifkannya lagi."
              : null,
        }));
        return false;
      }

      const statusResult = await getBrowserPushStatus(followerID);
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
        followerID,
        toSubscriptionPayload(subscription),
      );

      state.update((prev) => ({
        ...prev,
        busy: false,
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

    let followerID: string | null = null;
    state.update((prev) => {
      followerID = prev.followerID;
      return {
        ...prev,
        busy: true,
        error: null,
        success: null,
      };
    });

    if (!followerID) {
      followerID = getOrCreateIssueFollowerId();
    }

    try {
      const registration = await ensureServiceWorkerRegistration();
      const subscription = await registration.pushManager.getSubscription();

      if (subscription && followerID) {
        await unsubscribeBrowserPush(followerID, subscription.endpoint);
        await subscription.unsubscribe();
      }

      const permission: PushPermission = Notification.permission;
      state.update((prev) => ({
        ...prev,
        busy: false,
        subscribed: false,
        subscriptionCount: 0,
        permission,
        status: deriveStatus(true, permission, false),
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
