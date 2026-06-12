'use client';

import { useEffect, useRef, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { supabase } from '@/features/auth/lib/supabase';
import { fbqTrack } from '@/shared/lib/fbq';

const REGISTERED_FLAG = 'sup-fb-registered';

function trackRegistrationOnce() {
  try {
    if (localStorage.getItem(REGISTERED_FLAG)) return;
    fbqTrack('CompleteRegistration', { content_name: 'Google Sign-In' });
    localStorage.setItem(REGISTERED_FLAG, '1');
  } catch {
    fbqTrack('CompleteRegistration', { content_name: 'Google Sign-In' });
  }
}

function CallbackContent() {
  const router = useRouter();
  const params = useSearchParams();
  const handled = useRef(false);

  useEffect(() => {
    if (handled.current) return;
    handled.current = true;

    const next = params.get('next') ?? '/book';

    async function exchange() {
      const code = new URLSearchParams(window.location.search).get('code');
      if (code) {
        const { error } = await supabase.auth.exchangeCodeForSession(code);
        if (error) {
          router.replace('/signin?error=auth_failed');
          return;
        }
        trackRegistrationOnce();
        router.replace(next);
        return;
      }

      const { data: { session } } = await supabase.auth.getSession();
      if (session) {
        trackRegistrationOnce();
        router.replace(next);
        return;
      }

      const { data: listener } = supabase.auth.onAuthStateChange((_event, session) => {
        listener.subscription.unsubscribe();
        if (session) {
          trackRegistrationOnce();
          router.replace(next);
        } else {
          router.replace('/signin?error=auth_failed');
        }
      });

      const timeout = setTimeout(() => {
        listener.subscription.unsubscribe();
        router.replace('/signin?error=auth_failed');
      }, 5000);

      return () => {
        clearTimeout(timeout);
        listener.subscription.unsubscribe();
      };
    }

    exchange();
  }, [params, router]);

  return (
      <div className="full-page-spinner">
        <div className="spinner" />
      </div>
  );
}

export default function AuthCallbackPage() {
  return (
      <Suspense fallback={<div className="full-page-spinner"><div className="spinner" /></div>}>
        <CallbackContent />
      </Suspense>
  );
}