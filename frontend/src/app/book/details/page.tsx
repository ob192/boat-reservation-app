'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useRouter } from 'next/navigation';
import { Instrument_Serif, Inter_Tight, JetBrains_Mono } from 'next/font/google';
import { useUser } from '@/features/auth/hooks/useUser';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useCreateBooking, useCreateCheckout } from '@/features/booking/hooks';
import { contactSchema, type ContactFormValues } from '@/features/booking/schema/booking.schema';
import { UserMenu } from '@/features/auth/components/UserMenu';
import { MESSAGES } from '@/features/booking/messages';
import { calculateBookingTotal } from '@/features/booking/pricing';
import { formatCurrency } from '@/shared/lib/currency';
import { ConsentAgreement } from '@/features/booking/components/client/ConsentAgreement';
import { CONSENT_AGREEMENT, buildAgreementText } from '@/features/booking/consent-text';
import { sha256Hex, buildConsentRecord } from '@/shared/lib/consent';
import { fbqTrack } from '@/shared/lib/fbq';

const instrument = Instrument_Serif({
    subsets: ['latin'],
    weight: ['400'],
    style: ['normal', 'italic'],
    variable: '--bk-serif',
    display: 'swap',
});
const interTight = Inter_Tight({
    subsets: ['latin', 'cyrillic'],
    weight: ['400', '500', '600'],
    variable: '--bk-sans',
    display: 'swap',
});
const jetbrains = JetBrains_Mono({
    subsets: ['latin'],
    weight: ['400', '500'],
    variable: '--bk-mono',
    display: 'swap',
});

const STEPS = [
    { n: '01', label: 'Маршрут' },
    { n: '02', label: 'Дата' },
    { n: '03', label: 'Час' },
    { n: '04', label: 'Човни' },
    { n: '05', label: 'Деталі' },
    { n: '06', label: 'Готово' },
];

function ChevronLeft() {
    return (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M15 5l-7 7 7 7" stroke="currentColor" strokeWidth="1.8"
                  strokeLinecap="round" strokeLinejoin="round" />
        </svg>
    );
}

function ShieldHeartIcon() {
    return (
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" aria-hidden="true" style={{ flexShrink: 0, marginTop: 1 }}>
            <path d="M12 22s-8-4.5-8-11V5l8-3 8 3v6c0 6.5-8 11-8 11z"
                  stroke="currentColor" strokeWidth="1.6" strokeLinejoin="round" />
            <path d="M9.5 11c0-1.4 2.5-3.5 2.5-1 0-2.5 2.5-.4 2.5 1 0 1.5-2.5 3-2.5 3s-2.5-1.5-2.5-3z"
                  stroke="currentColor" strokeWidth="1.4" strokeLinejoin="round" />
        </svg>
    );
}

export default function DetailsPage() {
    const user = useUser();
    const isGuest = !user?.email;

    const router = useRouter();
    const {
        selectedRoute, selectedDate, selectedTime, quantities,
        contact, setContact, setBookingId, setSessionId,
    } = useBookingStore();

    const createBooking = useCreateBooking();
    const createCheckout = useCreateCheckout();

    const {
        register,
        handleSubmit,
        setValue,
        formState: { errors },
    } = useForm<ContactFormValues>({
        defaultValues: {
            email: user?.email ?? contact.email,
            firstName: contact.firstName || (user?.user_metadata?.full_name?.split(' ')[0] ?? ''),
            lastName: contact.lastName || (user?.user_metadata?.full_name?.split(' ').slice(1).join(' ') ?? ''),
            phone: contact.phone ?? '',
        },
        resolver: zodResolver(contactSchema),
    });

    const [agreed, setAgreed] = useState(false);
    const [docHash, setDocHash] = useState('');

    useEffect(() => {
        sha256Hex(buildAgreementText()).then(setDocHash);
    }, []);

    useEffect(() => {
        if (user?.email) setValue('email', user.email);
    }, [user, setValue]);

    const onSubmit = async (data: ContactFormValues) => {
        if (!selectedRoute || !selectedDate || !selectedTime) return;
        setContact(data);
        try {
            const consent = await buildConsentRecord({
                agreementId: CONSENT_AGREEMENT.id,
                agreementVersion: CONSENT_AGREEMENT.version,
                agreementHash: docHash,
                user: {
                    id: user?.id,
                    email: data.email,
                    name: `${data.firstName} ${data.lastName}`.trim(),
                },
            });

            const booking = await createBooking.mutateAsync({
                routeName: selectedRoute,
                date: selectedDate,
                time: selectedTime,
                quantities,
                contact: data,
                consent,
            });
            setBookingId(booking.bookingId);
            const origin = window.location.origin;
            const checkout = await createCheckout.mutateAsync({
                bookingId: booking.bookingId,
                resultUrl: `${origin}/book/success?session_id=${booking.bookingId}`,
            });
            setSessionId(checkout.sessionId);

            fbqTrack('InitiateCheckout', {
                value: total,
                currency: 'UAH',
                content_name: selectedRoute,
                content_category: 'sup_booking',
                num_items: quantities.big + quantities.medium + quantities.small,
            });

            window.location.href = checkout.checkoutUrl;
        } catch (err: unknown) {
            console.error('Checkout error:', err);
        }
    };

    const isLoading = createBooking.isPending || createCheckout.isPending;
    const apiError = createBooking.error ?? createCheckout.error;

    const total = calculateBookingTotal(selectedRoute, quantities);

    const dateLabel = selectedDate
        ? new Date(selectedDate + 'T00:00:00').toLocaleDateString('uk-UA', {
            weekday: 'long', day: 'numeric', month: 'long',
        })
        : '—';

    const boardParts: string[] = [];
    if (quantities.big > 0) boardParts.push(`${quantities.big}× великий`);
    if (quantities.medium > 0) boardParts.push(`${quantities.medium}× середній`);
    if (quantities.child > 0) boardParts.push(`${quantities.child}× дитина`);

    return (
        <div className={`${instrument.variable} ${interTight.variable} ${jetbrains.variable} bk-root`}>
            <div className="bk-shell">

                {/* ── Brand bar ── */}
                <header className="bk-topbar">
                    <Link href="/" className="bk-brand">
                        <span className="bk-monogram">S</span>
                        <span className="bk-wordmark">
                            <span className="bk-wordmark-name">SUP Chernihiv</span>
                            <span className="bk-wordmark-sub">оренда SUP-бордів</span>
                        </span>
                    </Link>
                    <UserMenu />
                </header>

                {/* ── Stepper ── */}
                <nav className="bk-stepper" aria-label="Прогрес бронювання">
                    {STEPS.map((s, i) => (
                        <div
                            key={s.n}
                            className={`bk-seg ${i < 3 ? 'done' : ''} ${i === 3 ? 'active' : ''}`}
                            aria-current={i === 3 ? 'step' : undefined}
                        >
                            <span className="bk-seg-bar" />
                            <span className="bk-seg-label">
                                <span className="bk-seg-num">{s.n}</span> {s.label}
                            </span>
                        </div>
                    ))}
                </nav>

                {/* ── Intro ── */}
                <section className="bk-intro">
                    <p className="bk-eyebrow bk-eyebrow--lead">
                        <span className="bk-eyebrow-rule" />
                        Крок 04 — Деталі
                    </p>
                    <h1 className="bk-headline">Ваші дані.</h1>
                    <p className="bk-details-hint">
                        Підтвердження бронювання та квиток надійдуть на вашу електронну пошту
                        одразу після оплати.
                    </p>
                </section>

                {/* ── Form ── */}
                <form onSubmit={handleSubmit(onSubmit)} noValidate>

                    {/* Name row */}
                    <div className="bk-field-row">
                        <div className="bk-field">
                            <label className="bk-field-label" htmlFor="firstName">
                                Ім'я
                            </label>
                            <input
                                id="firstName"
                                className={`bk-field-input ${errors.firstName ? 'bk-field-input--error' : ''}`}
                                {...register('firstName')}
                                autoComplete="given-name"
                                placeholder="Олена"
                            />
                            {errors.firstName && (
                                <span className="bk-field-error">{errors.firstName.message}</span>
                            )}
                        </div>
                        <div className="bk-field">
                            <label className="bk-field-label" htmlFor="lastName">
                                Прізвище
                            </label>
                            <input
                                id="lastName"
                                className={`bk-field-input ${errors.lastName ? 'bk-field-input--error' : ''}`}
                                {...register('lastName')}
                                autoComplete="family-name"
                                placeholder="Іваненко"
                            />
                            {errors.lastName && (
                                <span className="bk-field-error">{errors.lastName.message}</span>
                            )}
                        </div>
                    </div>

                    {/* Email */}
                    <div className="bk-field">
                        <label className="bk-field-label" htmlFor="email">
                            Електронна пошта
                        </label>
                        <input
                            id="email"
                            className={`bk-field-input ${isGuest ? '' : 'bk-field-input--readonly'} ${
                                isGuest && errors.email ? 'bk-field-input--error' : ''
                            }`}
                            type="email"
                            readOnly={!isGuest}
                            placeholder="you@example.com"
                            {...register('email')}
                            autoComplete="email"
                        />
                        {isGuest && errors.email && (
                            <span className="bk-field-error">{errors.email.message}</span>
                        )}
                        <span className="bk-field-hint">
                            {isGuest
                                ? 'Надішлемо підтвердження та квиток на цю адресу'
                                : "Пов'язано з вашим обліковим записом"}
                        </span>
                    </div>

                    {/* Phone */}
                    <div className="bk-field">
                        <label className="bk-field-label" htmlFor="phone">
                            Телефон
                        </label>
                        <input
                            id="phone"
                            className={`bk-field-input ${errors.phone ? 'bk-field-input--error' : ''}`}
                            type="tel"
                            placeholder="+380 50 123 45 67"
                            {...register('phone')}
                            autoComplete="tel"
                        />
                        {errors.phone && (
                            <span className="bk-field-error">{errors.phone.message}</span>
                        )}
                    </div>

                    {/* ── Order summary ── */}
                    <div className="bk-rule" style={{ margin: '22px 0' }} />

                    <div className="bk-order-card">
                        <div className="bk-order-heading">Ваше замовлення</div>

                        <div className="bk-order-rows">
                            <div className="bk-order-row">
                                <span className="bk-order-key">Дата</span>
                                <span className="bk-order-val">{dateLabel}</span>
                            </div>
                            <div className="bk-order-row">
                                <span className="bk-order-key">Час</span>
                                <span className="bk-order-val">{selectedTime ?? '—'}</span>
                            </div>
                            <div className="bk-order-row">
                                <span className="bk-order-key">Борди</span>
                                <span className="bk-order-val">
                                    {boardParts.length ? boardParts.join(', ') : '—'}
                                </span>
                            </div>
                        </div>

                        <div className="bk-order-total-row">
                            <span className="bk-order-total-label">До сплати</span>
                            <span className="bk-order-total-amount">{formatCurrency(total)}</span>
                        </div>
                    </div>

                    <ConsentAgreement agreed={agreed} onAgreedChange={setAgreed} docHash={docHash} />

                    {/* ── API error ── */}
                    {apiError && (
                        <div className="bk-banner bk-banner--error" role="alert" style={{ marginTop: 16 }}>
                            {(apiError as { message?: string }).message === 'SLOT_TAKEN'
                                ? MESSAGES.errors.slotTaken
                                : (apiError as { message?: string }).message === 'SLOT_CANCELLED'
                                    ? MESSAGES.errors.slotCancelled
                                    : (apiError as { message?: string }).message === 'BACKEND_UNAVAILABLE'
                                        ? MESSAGES.errors.backendUnavailable
                                        : MESSAGES.errors.bookingFailed}
                        </div>
                    )}

                    {/* ── Legal hint ── */}
                    <div className="bk-legal">
                        <ShieldHeartIcon />
                        <p className="bk-legal-text">
                            Натискаючи «Перейти до оплати», ви погоджуєтеся з{' '}
                            <span className="bk-legal-underline">умовами використання</span>{' '}
                            та{' '}
                            <span className="bk-legal-underline">політикою конфіденційності</span>.
                        </p>
                    </div>

                    <div className="bk-dock-spacer" />

                    {/* ── Sticky dock ── */}
                    <div className="bk-dock">
                        <div className="bk-dock-inner">
                            <button
                                className="bk-back"
                                onClick={() => router.push('/book/boats')}
                                type="button"
                                aria-label="Назад"
                            >
                                <ChevronLeft />
                            </button>
                            <button
                                className="bk-cta"
                                type="submit"
                                disabled={isLoading || !agreed}
                            >
                                {isLoading ? 'Обробка…' : 'Перейти до оплати'}
                            </button>
                        </div>
                    </div>
                </form>
            </div>
        </div>
    );
}