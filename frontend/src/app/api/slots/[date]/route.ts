import { env } from '@/shared/lib/env';

export async function GET(
  req: Request,
  { params }: { params: Promise<{ date: string }> },
) {
  const { date } = await params;
  const auth = req.headers.get('authorization');
  if (!auth?.startsWith('Bearer ')) {
    return Response.json({ message: 'NOT_AUTHENTICATED' }, { status: 401 });
  }

  try {
    const res = await fetch(`${env.BACKEND_URL}/slots/${date}`, {
      headers: { Authorization: auth },
      next: { revalidate: 15 },
    });

    if (!res.ok) {
      return Response.json({ message: 'BACKEND_ERROR' }, { status: res.status });
    }

    const data = await res.json();
    return Response.json(data);
  } catch {
    return Response.json({ message: 'BACKEND_UNAVAILABLE' }, { status: 503 });
  }
}
