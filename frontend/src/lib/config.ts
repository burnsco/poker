function getBackendUrl() {
  const envUrl = import.meta.env.VITE_BACKEND_URL;
  if (typeof window === "undefined") return envUrl || "";
  if (envUrl && (!envUrl.includes("localhost:4000") || window.location.hostname === "localhost")) {
    return envUrl;
  }
  return window.location.origin;
}

function getWebSocketBase() {
  const envWs = import.meta.env.VITE_BACKEND_WS_URL;
  if (typeof window === "undefined") return envWs || "ws://localhost:4000/socket";

  if (envWs && (!envWs.includes("localhost:4000") || window.location.hostname === "localhost")) {
    return envWs;
  }

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/socket`;
}

export const BACKEND_URL = getBackendUrl();
export const WEBSOCKET_BASE = getWebSocketBase();
