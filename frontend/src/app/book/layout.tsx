'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { AuthGuard } from '@/features/auth/components/AuthGuard';
import { UserMenu } from '@/features/auth/components/UserMenu';
import { StepIndicator } from '@/features/booking/components/client/StepIndicator';
import { MESSAGES } from '@/features/booking/messages';
import type { Step } from '@/features/booking/components/client/StepIndicator';

const STEP_MAP: Record<string, Step> = {
    '/book/date': 1,
    '/book/time': 2,
    '/book/boats': 3,
    '/book/details': 4,
    '/book/processing': 4,
    '/book/success': 5,
    '/book/cancelled': 1,
};

// Routes that render their own full editorial chrome (brand bar, stepper, dock).
// These bypass WizardShell while still being protected by AuthGuard.
const SELF_CONTAINED = new Set<string>([
    '/book/date',
    '/book/time',
    '/book/boats',
    '/book/details',
]);

function WizardShell({ children }: { children: React.ReactNode }) {
    const pathname = usePathname();
    const currentStep: Step = STEP_MAP[pathname] ?? 1;

    return (
        <>
            <header className="header">
                <Link href="/" className="logo" style={{ textDecoration: 'none' }}>
                    <div className="logo-icon">⛵</div>
                    <div className="logo-text">
                        <h1>{MESSAGES.header.title}</h1>
                        <span>{MESSAGES.header.subtitle}</span>
                    </div>
                </Link>
                <UserMenu />
            </header>

            <StepIndicator currentStep={currentStep} />

            <div className="card-wrap">
                <div className="card">{children}</div>
            </div>
        </>
    );
}

export default function BookLayout({ children }: { children: React.ReactNode }) {
    const pathname = usePathname();

    if (SELF_CONTAINED.has(pathname)) {
        return <AuthGuard>{children}</AuthGuard>;
    }

    return (
        <AuthGuard>
            <WizardShell>{children}</WizardShell>
        </AuthGuard>
    );
}