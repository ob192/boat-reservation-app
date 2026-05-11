'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useSession } from '@/features/auth/hooks/useSession';

function FullPageSpinner() {
  return (
    <div className="full-page-spinner">
      <div className="spinner" />
    </div>
  );
}

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { session, loading } = useSession();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!loading && !session) {
      const next = encodeURIComponent(pathname);
      router.replace(`/signin?next=${next}`);
    }
  }, [session, loading, router, pathname]);

  if (loading) return <FullPageSpinner />;
  if (!session) return null;
  return <>{children}</>;
}
