'use client';

import { DetailsForm } from '@/features/booking/components/client/DetailsForm';
import { MESSAGES } from '@/features/booking/messages';

export default function DetailsPage() {
  return (
    <>
      <div className="card-header">
        <div className="card-header-icon">📋</div>
        <div>
          <h3>{MESSAGES.details.title}</h3>
          <p>{MESSAGES.details.subtitle}</p>
        </div>
      </div>
      <div className="card-body">
        <DetailsForm />
      </div>
    </>
  );
}
