import { supabase } from './supabase';

const BASE = process.env.NEXT_PUBLIC_API_URL;

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

export async function adminFetch<T>(
    path: string,
    init: RequestInit & { parseJson?: boolean } = {},
): Promise<T> {
  const { parseJson = true, ...rest } = init;
  const { data: { session } } = await supabase.auth.getSession();
  if (!session) throw new ApiError(401, 'NOT_AUTHENTICATED');

  const res = await fetch(`/api${path}`, {
    ...rest,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${session.access_token}`,
      ...(rest.headers as Record<string, string> ?? {}),
    },
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({})) as { message?: string };
    throw new ApiError(res.status, err.message ?? 'Request failed');
  }

  if (!parseJson || res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export function formatDate(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

export function formatMonth(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  return `${y}-${m}`;
}

export function todayStr(): string {
  return formatDate(new Date());
}

export function currentMonthStr(): string {
  return formatMonth(new Date());
}
