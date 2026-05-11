'use client';

import { useSession } from './useSession';

export function useUser() {
  const { user } = useSession();
  return user;
}
