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

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = typeof window !== 'undefined' ? localStorage.getItem('codeforge_token') : null;
  const headers = new Headers(init.headers);
  if (token) headers.set('Authorization', `Bearer ${token}`);
  if (init.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  const response = await fetch(path, { ...init, headers });
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
