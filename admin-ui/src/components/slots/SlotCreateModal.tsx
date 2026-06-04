'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal } from '@/components/Modal';
import { useUpsertSlot } from '@/hooks/useSlots';
import { toast } from '@/hooks/useToast';
import { ApiError } from '@/lib/api';
import { ROUTES, RouteName, routeLabel } from '@/lib/routes';

const schema = z.object({
  time: z.string().regex(/^\d{2}:\d{2}$/, 'Оберіть час'),
  capacityBig: z.coerce.number().int().min(0, 'Мін. 0'),
  capacityMedium: z.coerce.number().int().min(0, 'Мін. 0'),
  capacitySmall: z.coerce.number().int().min(0, 'Мін. 0'),
}).refine(d => d.capacityBig + d.capacityMedium + d.capacitySmall > 0, {
  message: 'Хоча б одна місткість має бути > 0',
  path: ['capacityBig'],
});

type FormData = z.infer<typeof schema>;

// ── Drum Picker ───────────────────────────────────────────────────
const ITEM_H = 40;
const VISIBLE_H = 200;

interface DrumProps {
  count: number;
  value: number;
  onChange: (v: number) => void;
  label: string;
}

function Drum({ count, value, onChange, label }: DrumProps) {
  const wrapRef = useRef<HTMLDivElement>(null);
  const drumRef = useRef<HTMLDivElement>(null);
  const state = useRef({ offset: 0, velocity: 0, lastY: 0, lastTime: 0, dragging: false, raf: 0 });

  const items = Array.from({ length: count }, (_, i) => String(i).padStart(2, '0'));

  const clamp = (off: number) =>
      Math.max(-(count - 1) * ITEM_H, Math.min(0, off));

  const applyOffset = useCallback((off: number, animate: boolean) => {
    if (!drumRef.current) return;
    drumRef.current.style.transition = animate
        ? 'transform .25s cubic-bezier(.25,.46,.45,.94)'
        : 'none';
    drumRef.current.style.transform =
        `translateY(${off + (VISIBLE_H / 2 - ITEM_H / 2)}px)`;
  }, []);

  const snapTo = useCallback((idx: number) => {
    idx = Math.max(0, Math.min(count - 1, idx));
    state.current.offset = -idx * ITEM_H;
    applyOffset(state.current.offset, true);
    onChange(idx);
  }, [count, applyOffset, onChange]);

  // Sync external value changes
  useEffect(() => {
    state.current.offset = -value * ITEM_H;
    applyOffset(state.current.offset, true);
  }, [value, applyOffset]);

  useEffect(() => {
    const wrap = wrapRef.current;
    if (!wrap) return;

    const onDown = (e: MouseEvent | TouchEvent) => {
      cancelAnimationFrame(state.current.raf);
      const y = 'touches' in e ? e.touches[0].clientY : e.clientY;
      state.current = { ...state.current, dragging: true, lastY: y, lastTime: Date.now(), velocity: 0 };
      if (drumRef.current) drumRef.current.style.transition = 'none';
      e.preventDefault();
    };

    const onMove = (e: MouseEvent | TouchEvent) => {
      if (!state.current.dragging) return;
      const y = 'touches' in e ? (e as TouchEvent).touches[0].clientY : (e as MouseEvent).clientY;
      const now = Date.now();
      const startY = state.current.lastY +
          (state.current.offset - (-value * ITEM_H)); // keep track via closure
      // simpler: track from initial
      state.current.velocity = (y - state.current.lastY) / Math.max(1, now - state.current.lastTime) * 16;
      state.current.offset = clamp(state.current.offset + (y - state.current.lastY));
      state.current.lastY = y;
      state.current.lastTime = now;
      applyOffset(state.current.offset, false);
      e.preventDefault();
    };

    const onUp = () => {
      if (!state.current.dragging) return;
      state.current.dragging = false;
      const fling = () => {
        state.current.velocity *= 0.94;
        if (Math.abs(state.current.velocity) < 0.5) {
          snapTo(Math.round(-state.current.offset / ITEM_H));
          return;
        }
        state.current.offset = clamp(state.current.offset + state.current.velocity);
        applyOffset(state.current.offset, false);
        state.current.raf = requestAnimationFrame(fling);
      };
      if (Math.abs(state.current.velocity) < 0.5) {
        snapTo(Math.round(-state.current.offset / ITEM_H));
      } else {
        state.current.raf = requestAnimationFrame(fling);
      }
    };

    const onWheel = (e: WheelEvent) => {
      e.preventDefault();
      snapTo(Math.round(-state.current.offset / ITEM_H) + (e.deltaY > 0 ? 1 : -1));
    };

    wrap.addEventListener('mousedown', onDown);
    wrap.addEventListener('touchstart', onDown, { passive: false });
    wrap.addEventListener('wheel', onWheel, { passive: false });
    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
    window.addEventListener('touchmove', onMove, { passive: false });
    window.addEventListener('touchend', onUp);

    return () => {
      wrap.removeEventListener('mousedown', onDown);
      wrap.removeEventListener('touchstart', onDown);
      wrap.removeEventListener('wheel', onWheel);
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
      window.removeEventListener('touchmove', onMove);
      window.removeEventListener('touchend', onUp);
      cancelAnimationFrame(state.current.raf);
    };
  }, [snapTo, applyOffset, count]);

  return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6 }}>
      <span style={{ fontSize: '0.7rem', fontWeight: 600, color: 'var(--subtle)', letterSpacing: '0.08em', textTransform: 'uppercase' }}>
        {label}
      </span>
        <div
            ref={wrapRef}
            style={{
              position: 'relative',
              width: 80,
              height: VISIBLE_H,
              overflow: 'hidden',
              cursor: 'grab',
              userSelect: 'none',
            }}
        >
          {/* fade top */}
          <div style={{
            position: 'absolute', top: 0, left: 0, right: 0, height: 72,
            background: 'linear-gradient(to bottom, white 0%, transparent 100%)',
            zIndex: 2, pointerEvents: 'none',
          }} />
          {/* selection band */}
          <div style={{
            position: 'absolute', top: '50%', left: 0, right: 0,
            height: ITEM_H, marginTop: -ITEM_H / 2,
            borderTop: '1px solid var(--mist)',
            borderBottom: '1px solid var(--mist)',
            zIndex: 1, pointerEvents: 'none',
          }} />
          {/* fade bottom */}
          <div style={{
            position: 'absolute', bottom: 0, left: 0, right: 0, height: 72,
            background: 'linear-gradient(to top, white 0%, transparent 100%)',
            zIndex: 2, pointerEvents: 'none',
          }} />
          <div ref={drumRef} style={{ position: 'absolute', left: 0, right: 0, willChange: 'transform' }}>
            {items.map((item, i) => (
                <div
                    key={i}
                    onClick={() => snapTo(i)}
                    style={{
                      height: ITEM_H,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: i === value ? '1.4rem' : '1.1rem',
                      fontWeight: i === value ? 700 : 400,
                      fontFamily: 'Playfair Display, serif',
                      color: i === value ? 'var(--navy)' : 'var(--subtle)',
                      transition: 'color .15s, font-size .15s',
                      cursor: 'pointer',
                    }}
                >
                  {item}
                </div>
            ))}
          </div>
        </div>
      </div>
  );
}

// ── Modal ─────────────────────────────────────────────────────────
interface SlotCreateModalProps {
  open: boolean;
  onClose: () => void;
  date: string;
  route: string;            // NEW
}

export function SlotCreateModal({ open, onClose, date, route }: SlotCreateModalProps) {
  const { mutateAsync, isPending } = useUpsertSlot(date);
  const [hour, setHour] = useState(0);
  const [minute, setMinute] = useState(0);

  const { register, handleSubmit, reset, setValue, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { time: '00:00', capacityBig: 1, capacityMedium: 1, capacitySmall: 1 },
  });

  useEffect(() => {
    const hh = String(hour).padStart(2, '0');
    const mm = String(minute).padStart(2, '0');
    setValue('time', `${hh}:${mm}`, { shouldValidate: true });
  }, [hour, minute, setValue]);

  const onSubmit = async (data: FormData) => {
    try {
      await mutateAsync({ ...data, route });               // route added
      toast('Слот створено', 'success');
      reset();
      setHour(0);
      setMinute(0);
      setRoute('Desna');                                   // NEW
      onClose();
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) {
        toast('Вже існує', 'error');
      } else {
        toast('Помилка сервера', 'error');
      }
    }
  };

  const { ref: timeRef, ...timeRest } = register('time');

  return (
      <Modal
          open={open}
          onClose={onClose}
          title="Додати новий слот"
          footer={
            <>
              <button className="btn btn-secondary" onClick={onClose} disabled={isPending}>
                Відміна
              </button>
              <button className="btn btn-primary" onClick={handleSubmit(onSubmit)} disabled={isPending}>
                {isPending ? 'Створення…' : 'Створити'}
              </button>
            </>
          }
      >
        <input type="hidden" {...timeRest} ref={timeRef} />

        {/* before the Час form-group */}
        <div className="form-group">
          <label className="form-label">Маршрут</label>
          <div style={{ display: 'flex', gap: 8 }}>
            {ROUTES.map(r => (
                <button
                    key={r}
                    type="button"
                    className={`btn btn-sm ${route === r ? 'btn-primary' : 'btn-secondary'}`}
                    style={{ flex: 1, justifyContent: 'center' }}
                    onClick={() => setRoute(r)}
                >
                  {routeLabel(r)}
                </button>
            ))}
          </div>
        </div>
        
        <div className="form-group">
          <label className="form-label">Час</label>

          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 8,
            padding: '8px 0',
            borderRadius: 'var(--radius)',
            border: errors.time ? '1.5px solid var(--coral)' : '1.5px solid var(--mist)',
            background: 'var(--cream)',
          }}>
            <Drum count={24} value={hour} onChange={setHour} label="година" />
            <span style={{
              fontSize: '2rem', fontWeight: 700,
              fontFamily: 'Playfair Display, serif',
              color: 'var(--navy)', alignSelf: 'center',
              marginTop: 24,
            }}>:</span>
            <Drum count={60} value={minute} onChange={setMinute} label="хвилина" />
          </div>

          <div style={{
            textAlign: 'center',
            fontFamily: 'Playfair Display, serif',
            fontSize: '1.1rem',
            fontWeight: 700,
            color: 'var(--navy)',
            marginTop: 8,
          }}>
            {String(hour).padStart(2, '0')}:{String(minute).padStart(2, '0')}
          </div>

          {errors.time && <span className="form-error">{errors.time.message}</span>}
        </div>

        <div className="form-group">
          <label className="form-label">Місткість (великі човни)</label>
          <input
              type="number"
              className={`form-input${errors.capacityBig ? ' error' : ''}`}
              {...register('capacityBig')}
              min={1}
          />
          {errors.capacityBig && <span className="form-error">{errors.capacityBig.message}</span>}
        </div>

        <div className="form-group">
          <label className="form-label">Місткість (середні човни)</label>
          <input
              type="number"
              className={`form-input${errors.capacityMedium ? ' error' : ''}`}
              {...register('capacityMedium')}
              min={1}
          />
          {errors.capacityMedium && <span className="form-error">{errors.capacityMedium.message}</span>}
        </div>

        {/* after the "середні човни" form-group */}
        <div className="form-group">
          <label className="form-label">Місткість (малі човни)</label>
          <input
              type="number"
              className={`form-input${errors.capacitySmall ? ' error' : ''}`}
              {...register('capacitySmall')}
              min={0}
          />
          {errors.capacitySmall && <span className="form-error">{errors.capacitySmall.message}</span>}
        </div>
      </Modal>
  );
}