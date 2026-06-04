import { env } from '@/shared/lib/env';

export async function GET(
    req: Request,
    { params }: { params: Promise<{ date: string; route: string }> },
) {
    const { date, route } = await params;
    const auth = req.headers.get('authorization');
    if (!auth?.startsWith('Bearer ')) {
        return Response.json({ message: 'NOT_AUTHENTICATED' }, { status: 401 });
    }

    try {
        const res = await fetch(`${env.BACKEND_URL}/api/slots/${date}/${route}`, {
            headers: { Authorization: auth },
        });

        if (!res.ok) {
            // Forward the backend's message (INVALID_ROUTE / INVALID_DATE) when present
            const data = await res.json().catch(() => ({}));
            return Response.json(
                { message: (data as { message?: string }).message ?? 'BACKEND_ERROR' },
                { status: res.status },
            );
        }

        const data = await res.json();
        return Response.json(data);
    } catch {
        return Response.json({ message: 'BACKEND_UNAVAILABLE' }, { status: 503 });
    }
}