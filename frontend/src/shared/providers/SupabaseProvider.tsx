'use client';

import { useEffect } from 'react';
import { supabase } from '@/features/auth/lib/supabase';

export function SupabaseProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    supabase.auth.getSession();

    const { data: listener } = supabase.auth.onAuthStateChange((event) => {
      if (event === 'SIGNED_OUT') {
        try {
          localStorage.removeItem('harbour-wave-booking');
        } catch {
          // ignore
        }
      }
    });

    return () => listener.subscription.unsubscribe();
  }, []);

  return <>{children}</>;
}
