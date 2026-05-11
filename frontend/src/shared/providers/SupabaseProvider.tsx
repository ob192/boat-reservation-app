'use client';

import { useEffect } from 'react';
import { supabase } from '@/features/auth/lib/supabase';

/**
 * Kicks off Supabase's detectSessionInUrl logic on mount.
 * Also sets up global auth state listener to clear bookingStore on sign-out.
 */
export function SupabaseProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    // getSession() triggers detectSessionInUrl internally — parses #access_token from hash
    supabase.auth.getSession();

    const { data: listener } = supabase.auth.onAuthStateChange((event) => {
      if (event === 'SIGNED_OUT') {
        // Optionally clear persisted wizard state on sign-out
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
