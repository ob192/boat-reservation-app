'use client';

import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useRouter } from 'next/navigation';
import { useUser } from '@/features/auth/hooks/useUser';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useCreateBooking, useCreateCheckout } from '@/features/booking/hooks';
import { contactSchema, type ContactFormValues } from '@/features/booking/schema/booking.schema';
import { MESSAGES } from '@/features/booking/messages';

export function DetailsForm() {
  const user = useUser();
  const router = useRouter();
  const { selectedDate, selectedTime, quantities, contact, setContact, setBookingId, setSessionId } =
      useBookingStore();

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

  useEffect(() => {
    if (user?.email) setValue('email', user.email);
  }, [user, setValue]);

  const onSubmit = async (data: ContactFormValues) => {
    if (!selectedDate || !selectedTime) return;

    setContact(data);

    try {
      const booking = await createBooking.mutateAsync({
        date: selectedDate,
        time: selectedTime,
        quantities,
        contact: data,
      });

      setBookingId(booking.bookingId);

      // Build absolute redirect URLs so the payment provider can redirect back
      const origin = window.location.origin;
      const checkout = await createCheckout.mutateAsync({
        bookingId: booking.bookingId,
        resultUrl: `${origin}/book/success?session_id=${booking.bookingId}`,
      });

      setSessionId(checkout.sessionId);

      // Redirect to payment provider checkout page
      window.location.href = checkout.checkoutUrl;
    } catch (err: unknown) {
      console.error('Checkout error:', err);
    }
  };

  const isLoading = createBooking.isPending || createCheckout.isPending;
  const error = createBooking.error ?? createCheckout.error;

  return (
      <form onSubmit={handleSubmit(onSubmit)} noValidate>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label" htmlFor="firstName">
              {MESSAGES.details.firstName}
            </label>
            <input
                id="firstName"
                className="form-input"
                {...register('firstName')}
                autoComplete="given-name"
            />
            {errors.firstName && <p className="form-error">{errors.firstName.message}</p>}
          </div>
          <div className="form-group">
            <label className="form-label" htmlFor="lastName">
              {MESSAGES.details.lastName}
            </label>
            <input
                id="lastName"
                className="form-input"
                {...register('lastName')}
                autoComplete="family-name"
            />
            {errors.lastName && <p className="form-error">{errors.lastName.message}</p>}
          </div>
        </div>

        <div className="form-group">
          <label className="form-label" htmlFor="email">
            {MESSAGES.details.email}
          </label>
          <input
              id="email"
              className="form-input"
              type="email"
              readOnly
              {...register('email')}
              autoComplete="email"
          />
          <p style={{ fontSize: '0.68rem', color: 'var(--subtle)', marginTop: '0.25rem' }}>
            {MESSAGES.details.emailReadonly}
          </p>
          {errors.email && <p className="form-error">{errors.email.message}</p>}
        </div>

        <div className="form-group">
          <label className="form-label" htmlFor="phone">
            {MESSAGES.details.phone}
          </label>
          <input
              id="phone"
              className="form-input"
              type="tel"
              placeholder={MESSAGES.details.phonePlaceholder}
              {...register('phone')}
              autoComplete="tel"
          />
        </div>

        {error && (
            <div
                style={{
                  background: '#fff5f3',
                  border: '1.5px solid var(--coral)',
                  borderRadius: 10,
                  padding: '0.75rem 1rem',
                  fontSize: '0.78rem',
                  color: '#c0392b',
                  marginBottom: '1rem',
                }}
                role="alert"
            >
              {(error as { message?: string }).message === 'SLOT_TAKEN'
                  ? MESSAGES.errors.slotTaken
                  : (error as { message?: string }).message === 'BACKEND_UNAVAILABLE'
                      ? MESSAGES.errors.backendUnavailable
                      : MESSAGES.errors.bookingFailed}
            </div>
        )}

        <div className="nav-btns">
          <button className="btn-ghost" onClick={() => router.back()} type="button">
            {MESSAGES.buttons.back}
          </button>
          <button className="btn-primary" type="submit" disabled={isLoading}>
            {isLoading ? '⏳ Обробка...' : MESSAGES.details.proceedToPayment}
          </button>
        </div>
      </form>
  );
}