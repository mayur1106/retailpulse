export type AuthResponse = {
  user: {
    userId: string;
    organizationId: string;
    role: string;
    email: string;
  };
  tokens: {
    accessToken: string;
    refreshToken: string;
    expiresAt: string;
  };
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:4005";
const TOKEN_STORAGE_KEY = "retailpulse.tokens";

type TokenPair = AuthResponse["tokens"];

export async function apiRequest<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const payload = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error(payload?.error?.message ?? "Request failed");
  }
  return payload as T;
}

export function getAccessToken(): string | null {
  if (typeof window === "undefined") {
    return null;
  }
  const raw = localStorage.getItem(TOKEN_STORAGE_KEY);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw).accessToken ?? null;
  } catch {
    return null;
  }
}

export function getRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  try {
    return JSON.parse(localStorage.getItem(TOKEN_STORAGE_KEY) ?? "null")?.refreshToken ?? null;
  } catch {
    return null;
  }
}

export function saveTokens(tokens: TokenPair): void {
  localStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(tokens));
}

export function clearTokens(): void {
  localStorage.removeItem(TOKEN_STORAGE_KEY);
}

let refreshPromise: Promise<TokenPair> | null = null;

async function refreshSession(): Promise<TokenPair> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) throw new Error("Your session has expired. Please sign in again.");
  if (!refreshPromise) {
    refreshPromise = apiRequest<TokenPair>("/v1/auth/refresh", { refreshToken })
      .then((tokens) => {
        saveTokens(tokens);
        return tokens;
      })
      .finally(() => {
        refreshPromise = null;
      });
  }
  return refreshPromise;
}

export async function authedRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getAccessToken();
  if (!token) {
    throw new Error("Please login first");
  }
  let response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...(options.headers ?? {}),
    },
  });
  if (response.status === 401) {
    try {
      const tokens = await refreshSession();
      response = await fetch(`${API_BASE_URL}${path}`, {
        ...options,
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${tokens.accessToken}`,
          ...(options.headers ?? {}),
        },
      });
    } catch (error) {
      clearTokens();
      throw error;
    }
  }
  const payload = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error(payload?.error?.message ?? "Request failed");
  }
  return payload as T;
}

export async function logout(): Promise<void> {
  const refreshToken = getRefreshToken();
  try {
    if (refreshToken) await apiRequest<void>("/v1/auth/logout", { refreshToken });
  } finally {
    clearTokens();
  }
}
