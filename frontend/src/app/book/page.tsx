'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function BookPage() {
  const router = useRouter();
  useEffect(() => { router.replace('/book/route'); }, [router]);
  return null;
}
