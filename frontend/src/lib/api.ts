export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly code: number,
    public readonly status: number,
  ) {
    super(message);
  }
}

export async function apiDownload(path: string, body: unknown, filename: string) {
  const response = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!response.ok) throw new Error('导出失败');
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}

export function jsonBody(value: unknown): RequestInit {
  return { method: 'POST', body: JSON.stringify(value) };
}

// ---- 用户态管理 ----

export interface AuthUser {
  user_id: string;
  username: string;
  email?: string;
  avatar?: string;
  level?: number;
  level_title?: string;
  stats?: Record<string, number>;
}

export interface AuthResult {
  user_id: string;
  username: string;
  token: string;
  refresh_token: string;
  expires_in: number;
}

const TOKEN_KEY = 'codeforge_token';
const REFRESH_KEY = 'codeforge_refresh_token';

export function getToken(): string | null {
  return typeof window !== 'undefined' ? localStorage.getItem(TOKEN_KEY) : null;
}

export function isLoggedIn(): boolean {
  return !!getToken();
}

export function saveAuth(result: AuthResult) {
  localStorage.setItem(TOKEN_KEY, result.token);
  localStorage.setItem(REFRESH_KEY, result.refresh_token);
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_KEY);
}

// ---- 401 自动续期:7 天免密登录 ----
// access token(24h)过期后,用 refresh token 换新 access 并重试原请求。
// 用 promise 单飞:并发请求共享同一次刷新,避免重复调用 refresh。

let refreshPromise: Promise<string> | null = null;

function refreshAccessToken(): Promise<string> {
  if (!refreshPromise) {
    const refreshToken = localStorage.getItem(REFRESH_KEY);
    if (!refreshToken) {
      refreshPromise = Promise.reject(new Error('无刷新令牌'));
    } else {
      refreshPromise = fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: refreshToken }),
      })
        .then(async (response) => {
          const payload = (await response.json()) as ApiEnvelope<{ token: string }> | { token?: string };
          const token =
            payload && typeof payload === 'object' && 'data' in payload
              ? (payload as ApiEnvelope<{ token: string }>).data?.token
              : (payload as { token?: string }).token;
          if (!response.ok || !token) throw new Error('刷新令牌失效');
          localStorage.setItem(TOKEN_KEY, token);
          return token;
        })
        .catch((err) => {
          // 刷新失败:清除登录态,下次请求以游客身份
          clearAuth();
          throw err;
        })
        .finally(() => {
          refreshPromise = null;
        });
    }
  }
  return refreshPromise;
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers = new Headers(init.headers);
  if (token) headers.set('Authorization', `Bearer ${token}`);
  if (init.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  const response = await fetch(path, { ...init, headers });
  // access token 过期(401)且有 refresh token 时,自动续期并重试一次
  if (response.status === 401 && localStorage.getItem(REFRESH_KEY)) {
    try {
      const newToken = await refreshAccessToken();
      const retryHeaders = new Headers(init.headers);
      retryHeaders.set('Authorization', `Bearer ${newToken}`);
      const retry = await fetch(path, { ...init, headers: retryHeaders });
      return handleResponse<T>(retry);
    } catch {
      // 刷新失败:fall through 到原始 401 错误
    }
  }
  return handleResponse<T>(response);
}

async function handleResponse<T>(response: Response): Promise<T> {
  const contentType = response.headers.get('content-type') || '';
  if (!contentType.includes('application/json')) {
    if (!response.ok) throw new ApiError(`请求失败：${response.status}`, response.status, response.status);
    return response as T;
  }
  const payload = (await response.json()) as ApiEnvelope<T> | T;
  if (payload && typeof payload === 'object' && 'code' in payload && 'data' in payload) {
    const envelope = payload as ApiEnvelope<T>;
    if (!response.ok || envelope.code !== 200) {
      throw new ApiError(envelope.message || '请求失败', envelope.code, response.status);
    }
    return envelope.data;
  }
  if (!response.ok) throw new ApiError('请求失败', response.status, response.status);
  return payload as T;
}

export function login(email: string, password: string) {
  return apiFetch<AuthResult>('/api/v1/auth/login', jsonBody({ email, password }));
}

export function register(username: string, email: string, password: string) {
  return apiFetch<AuthResult>('/api/v1/auth/register', jsonBody({ username, email, password }));
}

export function fetchMe() {
  return apiFetch<AuthUser>('/api/v1/users/me');
}
