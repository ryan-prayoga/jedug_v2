import { bootstrapDevice } from "$lib/api/device";
import { ApiError } from "$lib/api/client";
import { getAnonToken, setAnonToken } from "$lib/utils/storage";

type EnsureDeviceOptions = {
  forceRefresh?: boolean;
  retry?: number;
};

let bootstrapPromise: Promise<string> | null = null;

function isUsableToken(token: string | null): token is string {
  return typeof token === "string" && token.trim().length > 0;
}

async function bootstrapOnce(forceRefresh: boolean): Promise<string> {
  const existingToken = forceRefresh ? null : getAnonToken();
  const response = await bootstrapDevice(existingToken || undefined);
  const nextToken = response.data?.anon_token || existingToken;

  if (!isUsableToken(nextToken)) {
    throw new Error("bootstrap token missing");
  }

  if (nextToken !== existingToken) {
    setAnonToken(nextToken);
  }

  return nextToken;
}

export async function ensureDeviceBootstrap(
  options: EnsureDeviceOptions = {},
): Promise<string> {
  const { forceRefresh = false, retry = 1 } = options;

  if (!forceRefresh) {
    const token = getAnonToken();
    if (isUsableToken(token)) {
      return token;
    }
  }

  if (!bootstrapPromise) {
    bootstrapPromise = (async () => {
      let attempts = 0;
      let shouldForceRefresh = forceRefresh;
      let lastError: unknown;

      while (attempts <= retry) {
        try {
          return await bootstrapOnce(shouldForceRefresh);
        } catch (error) {
          lastError = error;
          attempts += 1;
          shouldForceRefresh = true;
          if (attempts > retry) {
            throw error;
          }
        }
      }

      throw lastError instanceof Error
        ? lastError
        : new Error("failed to bootstrap device");
    })();
  }

  try {
    return await bootstrapPromise;
  } finally {
    bootstrapPromise = null;
  }
}

export function isBootstrapMissingError(error: unknown): boolean {
  if (!(error instanceof ApiError)) return false;
  if (error.status !== 401) return false;
  return error.message.toLowerCase().includes("bootstrap");
}
