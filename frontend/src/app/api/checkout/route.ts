import { z } from 'zod';
import { env } from '@/shared/lib/env';

const checkoutBodySchema = z.object({
  bookingId: z.string().min(1),
  resultUrl: z.string().url(),
});

export async function POST(req: Request) {
  const auth = req.headers.get('authorization');
  if (!auth?.startsWith('Bearer ')) {
    return Response.json({ message: 'NOT_AUTHENTICATED' }, { status: 401 });
  }

  let body: unknown;
  try {
    body = await req.json();
  } catch {
    return Response.json({ message: 'INVALID_JSON' }, { status: 400 });
  }

  const parsed = checkoutBodySchema.safeParse(body);
  if (!parsed.success) {
    return Response.json({ message: 'INVALID_INPUT', issues: parsed.error.issues }, { status: 400 });
  }

  try {
    const res = await fetch(`${env.BACKEND_URL}/checkout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: auth,
      },
      body: JSON.stringify(parsed.data),
    });

    if (res.status === 401) return Response.json({ message: 'SESSION_EXPIRED' }, { status: 401 });
    if (res.status === 403) return Response.json({ message: 'FORBIDDEN' }, { status: 403 });
    if (res.status === 404) return Response.json({ message: 'BOOKING_NOT_FOUND' }, { status: 404 });

    if (!res.ok) {
      return Response.json({ message: 'CHECKOUT_FAILED' }, { status: 500 });
    }

    const data = await res.json();
    return Response.json(data);
  } catch {
    return Response.json({ message: 'BACKEND_UNAVAILABLE' }, { status: 503 });
  }
}