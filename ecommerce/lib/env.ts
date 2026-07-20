const LOCAL_HOSTS = new Set(["localhost", "127.0.0.1", "::1"]);

function isLocalURL(value: string) {
  try {
    const url = new URL(value);
    return LOCAL_HOSTS.has(url.hostname);
  } catch {
    return value.includes("localhost") || value.includes("127.0.0.1");
  }
}

export function apiBaseUrl() {
  if (typeof window !== "undefined" && process.env.NEXT_PUBLIC_DISABLE_STOREFRONT_API_PROXY !== "true") {
    return "";
  }

  const configured = process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/$/, "");
  if (typeof window === "undefined") {
    return (process.env.RETAILPULSE_API_INTERNAL_URL || configured || "http://localhost:4005").replace(/\/$/, "");
  }

  const { protocol, hostname, origin } = window.location;
  const isLocalPage = LOCAL_HOSTS.has(hostname);

  if (!configured || isLocalURL(configured)) {
    return "";
  }

  if (configured && !(isLocalURL(configured) && !isLocalPage)) {
    return configured;
  }

  if (isLocalPage) {
    return `${protocol}//${hostname}:4005`;
  }

  return origin;
}

export const storeSlug = process.env.NEXT_PUBLIC_STORE_SLUG || "rangavali";
