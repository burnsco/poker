export type ApiErrorPayload = {
  code?: string;
  error?: string;
  errors?: Record<string, string[] | string>;
  message?: string;
};

export function getBackendUrl() {
  const envUrl = import.meta.env.VITE_BACKEND_URL;
  if (typeof window === "undefined") return envUrl || "";

  // If we have an env var and it's not the default localhost (or we ARE on localhost), use it
  if (envUrl && (!envUrl.includes("localhost:4000") || window.location.hostname === "localhost")) {
    return envUrl;
  }

  // Otherwise, use the current origin (production)
  return window.location.origin;
}

export class ApiRequestError extends Error {
  readonly code?: string;
  readonly payload: ApiErrorPayload | null;
  readonly status: number;

  constructor(message: string, status: number, payload: ApiErrorPayload | null) {
    super(message);
    this.name = "ApiRequestError";
    this.status = status;
    this.payload = payload;
    this.code = payload?.code;
  }
}

function flattenFieldErrors(errors: ApiErrorPayload["errors"]) {
  if (!errors) return [];

  return Object.entries(errors).flatMap(([field, value]) => {
    const messages = Array.isArray(value) ? value : [value];
    return messages.map((message) => `${field} ${message}`);
  });
}

export function getApiErrorMessage(payload: ApiErrorPayload | null, fallback = "Request failed") {
  if (!payload) return fallback;

  const fieldErrors = flattenFieldErrors(payload.errors);
  if (fieldErrors.length > 0) return fieldErrors.join(", ");
  if (typeof payload.error === "string" && payload.error.trim()) return payload.error;
  if (typeof payload.message === "string" && payload.message.trim()) {
    return payload.message;
  }

  return fallback;
}

export async function readJsonSafely(response: Response) {
  const text = await response.text();
  if (!text) return null;

  try {
    return JSON.parse(text) as ApiErrorPayload | Record<string, unknown>;
  } catch {
    return null;
  }
}

export async function requestJson<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
  fallbackMessage?: string,
): Promise<T> {
  if (process.env.NODE_ENV === "test") {
    console.log(`[TEST FETCH] ${input.toString()}`);
  }
  const response = await fetch(input, init);
  const payload = await readJsonSafely(response);

  if (!response.ok) {
    throw new ApiRequestError(
      getApiErrorMessage(payload as ApiErrorPayload | null, fallbackMessage),
      response.status,
      payload as ApiErrorPayload | null,
    );
  }

  return payload as T;
}
