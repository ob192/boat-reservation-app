import { env } from '@/shared/lib/env';

export async function GET(
    req: Request,
    { params }: { params: Promise<{ code: string }> },
) {
    const { code } = await params;
    const auth = req.headers.get('authorization');
    if (!auth?.startsWith('Bearer ')) {
        return Response.json({ message: 'NOT_AUTHENTICATED' }, { status: 401 });
    }

    try {
        const res = await fetch(`${env.BACKEND_URL}/api/promocodes/${encodeURIComponent(code)}`, {
            headers: { Authorization: auth },
        });

        const data = await res.json().catch(() => ({}));

        if (res.status === 401) return Response.json({ message: 'SESSION_EXPIRED' }, { status: 401 });

        if (!res.ok) {
            // Forward the promo error code (PROMO_NOT_FOUND / PROMO_INACTIVE /
            // PROMO_EXHAUSTED) so the UI can show a specific message.
            const detail = data as { code?: string; message?: string };
            const msg = detail.code ?? detail.message ?? 'PROMO_INVALID';
            return Response.json({ message: msg }, { status: res.status });
        }

        return Response.json(data);
    } catch {
        return Response.json({ message: 'BACKEND_UNAVAILABLE' }, { status: 503 });
    }
}
