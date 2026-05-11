'use client';

import { useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useSession } from '@/features/auth/hooks/useSession';

function CallbackContent() {
  const { session, loading } = useSession();
  const router = useRouter();
  const params = useSearchParams();

  useEffect(() => {
    if (loading) return;
    const next = params.get('next') ?? '/book';
    router.replace(session ? next : '/signin?error=auth_failed');
  }, [session, loading, params, router]);

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
