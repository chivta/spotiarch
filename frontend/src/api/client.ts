const DEFAULT_API_BASE_URL = "/api";
const JSON_CONTENT_TYPE = "application/json";

const baseURL = (import.meta.env.VITE_API_BASE_URL || DEFAULT_API_BASE_URL).replace(/\/$/, "");

interface ErrorResponse {
  code?: string;
}

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(code: string, status: number) {
    super(code);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) headers.set("Content-Type", JSON_CONTENT_TYPE);

  let response: Response;
  try {
    response = await fetch(`${baseURL}${path}`, { ...init, headers, credentials: "include" });
  } catch {
    throw new ApiError("NETWORK_ERROR", 0);
  }

  if (!response.ok) {
    const body = await response.json().catch(() => ({})) as ErrorResponse;
    throw new ApiError(body.code || "INTERNAL_ERROR", response.status);
  }

  if (response.status === 204 || response.headers.get("content-length") === "0") {
    return undefined as T;
  }
  return response.json() as Promise<T>;
}
