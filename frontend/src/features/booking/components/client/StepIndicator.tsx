'use client';

import { MESSAGES } from '@/features/booking/messages';

export type Step = 1 | 2 | 3 | 4 | 5;

interface StepIndicatorProps {
  currentStep: Step;
}

const STEPS = [
  { n: 1 as Step, label: MESSAGES.nav.date },
  { n: 2 as Step, label: MESSAGES.nav.time },
  { n: 3 as Step, label: MESSAGES.nav.boats },
  { n: 4 as Step, label: MESSAGES.nav.details },
  { n: 5 as Step, label: MESSAGES.nav.done },
];

export function StepIndicator({ currentStep }: StepIndicatorProps) {
  return (
    <nav className="steps-nav" aria-label="Кроки бронювання">
      {STEPS.map((step, i) => {
        const isDone = step.n < currentStep;
        const isActive = step.n === currentStep;

        return (
          <div key={step.n} style={{ display: 'contents' }}>
            <div
              className={`step-dot ${isActive ? 'active' : ''} ${isDone ? 'done' : ''}`}
              aria-current={isActive ? 'step' : undefined}
            >
              <div className="dot">
                {isDone ? '✓' : step.n}
              </div>
              <div className="label">{step.label}</div>
            </div>
            {i < STEPS.length - 1 && (
              <div className={`step-line ${isDone ? 'done' : ''}`} aria-hidden="true" />
            )}
          </div>
        );
      })}
    </nav>
  );
}
