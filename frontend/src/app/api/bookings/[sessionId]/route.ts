import { env } from '@/shared/lib/env';

export async function GET(
  req: Request,
  { params }: { params: Promise<{ sessionId: string }> },
) {
  const { sessionId } = await params;
  const auth = req.headers.get('authorization');
  if (!auth?.startsWith('Bearer ')) {
    return Response.json({ message: 'NOT_AUTHENTICATED' }, { status: 401 });
  }

  try {
    const res = await fetch(`${env.BACKEND_URL}/bookings/by-session/${sessionId}`, {
      headers: { Authorization: auth },
    });

    if (res.status === 401) return Response.json({ message: 'SESSION_EXPIRED' }, { status: 401 });
    if (res.status === 404) return Response.json({ message: 'NOT_FOUND' }, { status: 404 });

    if (!res.ok) {
      return Response.json({ message: 'BACKEND_ERROR' }, { status: res.status });
    }

    const data = await res.json();
    return Response.json(data);
  } catch {
    return Response.json({ message: 'BACKEND_UNAVAILABLE' }, { status: 503 });
  }
}
