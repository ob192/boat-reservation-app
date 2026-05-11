'use client';

import { Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { SuccessPoller, ProcessingScreen } from '@/features/booking/components/client/ProcessingScreen';

function SuccessContent() {
  const params = useSearchParams();
  const sessionId = params.get('session_id');

  if (!sessionId) {
    return (
      <div className="confirm-screen">
        <p style={{ color: 'var(--subtle)' }}>Відсутній ідентифікатор сесії.</p>
      </div>
    );
  }

  return <SuccessPoller sessionId={sessionId} />;
}

export default function SuccessPage() {
  return (
    <Suspense fallback={<ProcessingScreen />}>
      <SuccessContent />
    </Suspense>
  );
}
