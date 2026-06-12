export const FB_PIXEL_ID = '1016024664696948';

type FbqFn = (...args: unknown[]) => void;

declare global {
    interface Window {
        fbq?: FbqFn;
        _fbq?: FbqFn;
    }
}

export function fbqTrack(event: string, params?: Record<string, unknown>) {
    if (typeof window === 'undefined' || !window.fbq) return;
    window.fbq('track', event, params);
}

export function fbqTrackCustom(event: string, params?: Record<string, unknown>) {
    if (typeof window === 'undefined' || !window.fbq) return;
    window.fbq('trackCustom', event, params);
}