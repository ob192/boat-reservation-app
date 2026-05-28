import { bookingSchema } from '@/features/booking/schema/booking.schema';
import { env } from '@/shared/lib/env';

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

  const parsed = bookingSchema.safeParse(body);
  if (!parsed.success) {
    return Response.json(
      { message: 'INVALID_INPUT', issues: parsed.error.issues },
      { status: 400 },
    );
  }

  const idempotencyKey = req.headers.get('x-idempotency-key') ?? crypto.randomUUID();

  try {
    const res = await fetch(`${env.BACKEND_URL}/api/bookings`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: auth,
        'X-Idempotency-Key': idempotencyKey,
      },
      body: JSON.stringify(parsed.data),
    });

    const data = await res.json();

    if (!res.ok) {
      if (res.status === 409) return Response.json({ message: 'SLOT_TAKEN' }, { status: 409 });
      if (res.status === 401) return Response.json({ message: 'SESSION_EXPIRED' }, { status: 401 });
      if (res.status === 422) return Response.json({ message: 'VALIDATION_FAILED', detail: data }, { status: 422 });
      return Response.json({ message: 'BOOKING_FAILED' }, { status: 500 });
    }

    return Response.json(data, { status: 201 });
  } catch {
    return Response.json({ message: 'BACKEND_UNAVAILABLE' }, { status: 503 });
  }
}
