'use client';

import { supabase } from '@/features/auth/lib/supabase';

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session) throw new ApiError(401, 'NOT_AUTHENTICATED');

  const res = await fetch(`/api${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${session.access_token}`,
      ...init.headers,
    },
  });

  if (res.status === 401) {
    await supabase.auth.signOut();
    window.location.href = '/signin?error=session_expired';
    throw new ApiError(401, 'SESSION_EXPIRED');
  }

  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new ApiError(res.status, (error as { message?: string }).message ?? 'Request failed');
  }

  return res.json() as Promise<T>;
}
