'use client';

export async function sha256Hex(str: string): Promise<string> {
    const buf = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(str));
    return [...new Uint8Array(buf)].map((b) => b.toString(16).padStart(2, '0')).join('');
}

function canvasSignal(): string {
    try {
        const c = document.createElement('canvas');
        const ctx = c.getContext('2d');
        if (!ctx) return 'no-canvas';
        ctx.textBaseline = 'top';
        ctx.font = "14px 'Arial'";
        ctx.fillStyle = '#069';
        ctx.fillText('Consent\u2705fingerprint', 2, 2);
        ctx.fillStyle = 'rgba(102,200,0,.7)';
        ctx.fillRect(20, 5, 60, 18);
        return c.toDataURL();
    } catch {
        return 'no-canvas';
    }
}

function webglSignal(): string {
    try {
        const gl = document.createElement('canvas').getContext('webgl') as WebGLRenderingContext | null;
        if (!gl) return 'no-webgl';
        const dbg = gl.getExtension('WEBGL_debug_renderer_info');
        return dbg ? String(gl.getParameter(dbg.UNMASKED_RENDERER_WEBGL)) : 'no-webgl';
    } catch {
        return 'no-webgl';
    }
}

export interface FingerprintComponents {
    userAgent: string;
    language: string;
    languages: string;
    platform: string;
    timezone: string;
    tzOffset: number;
    screen: string;
    cores: number;
    memory: number;
    touch: number;
    canvas: string;
    webgl: string;
}

export async function getDeviceFingerprint(): Promise<{ hash: string; components: FingerprintComponents }> {
    const n = navigator as Navigator & { deviceMemory?: number };
    const components: FingerprintComponents = {
        userAgent: n.userAgent,
        language: n.language,
        languages: (n.languages || []).join(','),
        platform: n.platform,
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        tzOffset: new Date().getTimezoneOffset(),
        screen: [screen.width, screen.height, screen.colorDepth, window.devicePixelRatio].join('x'),
        cores: n.hardwareConcurrency || 0,
        memory: n.deviceMemory || 0,
        touch: n.maxTouchPoints || 0,
        canvas: canvasSignal(),
        webgl: webglSignal(),
    };
    const hash = await sha256Hex(JSON.stringify(components));
    return { hash, components };
}

export interface ConsentRecord {
    event: 'consent_agreed';
    consentId: string;
    agreementId: string;
    agreementVersion: string;
    agreementHash: string;
    user: { id?: string; email: string; name: string };
    device: {
        fingerprint: string;
        userAgent: string;
        platform: string;
        timezone: string;
        language: string;
        screen: string;
    };
    clientTimestamp: string;
}

export async function buildConsentRecord(opts: {
    agreementId: string;
    agreementVersion: string;
    agreementHash: string;
    user: { id?: string; email: string; name: string };
}): Promise<ConsentRecord> {
    const fp = await getDeviceFingerprint();
    return {
        event: 'consent_agreed',
        consentId: 'cns_' + crypto.randomUUID(),
        agreementId: opts.agreementId,
        agreementVersion: opts.agreementVersion,
        agreementHash: opts.agreementHash, // proves *what* was agreed to
        user: opts.user,
        device: {
            fingerprint: fp.hash, // proves *what device*
            userAgent: fp.components.userAgent,
            platform: fp.components.platform,
            timezone: fp.components.timezone,
            language: fp.components.language,
            screen: fp.components.screen,
        },
        clientTimestamp: new Date().toISOString(),
    };
}