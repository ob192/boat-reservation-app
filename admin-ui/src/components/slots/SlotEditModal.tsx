'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Modal } from '@/components/Modal';
import { useUpsertSlot } from '@/hooks/useSlots';
import { toast } from '@/hooks/useToast';
import { SlotInfo } from '@/lib/types';
import { ApiError } from '@/lib/api';

const schema = z.object({
  capacityBig: z.coerce.number().int().min(0, 'Мін. 0'),
  capacityMedium: z.coerce.number().int().min(0, 'Мін. 0'),
}).refine(d => d.capacityBig + d.capacityMedium > 0, {
  message: 'Хоча б одна місткість має бути > 0',
  path: ['capacityBig'],
});

type FormData = z.infer<typeof schema>;

interface SlotEditModalProps {
  open: boolean;
  onClose: () => void;
  date: string;
  slot: SlotInfo;
}

export function SlotEditModal({ open, onClose, date, slot }: SlotEditModalProps) {
  const { mutateAsync, isPending } = useUpsertSlot(date);

  const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      capacityBig: slot.totalBig,
      capacityMedium: slot.totalMedium,
    },
  });

  const onSubmit = async (data: FormData) => {
    try {
      await mutateAsync({ time: slot.time, ...data });
      toast('Збережено', 'success');
      onClose();
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Помилка сервера', 'error');
    }
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Редагувати слот ${slot.time}`}
      footer={
        <>
          <button className="btn btn-secondary" onClick={onClose} disabled={isPending}>
            Відміна
          </button>
          <button className="btn btn-primary" onClick={handleSubmit(onSubmit)} disabled={isPending}>
            {isPending ? 'Збереження…' : 'Зберегти'}
          </button>
        </>
      }
    >
      <div className="form-group">
        <label className="form-label">Місткість (великі човни)</label>
        <input
          type="number"
          className={`form-input${errors.capacityBig ? ' error' : ''}`}
          {...register('capacityBig')}
          min={0}
        />
        {errors.capacityBig && <span className="form-error">{errors.capacityBig.message}</span>}
      </div>

      <div className="form-group">
        <label className="form-label">Місткість (середні човни)</label>
        <input
          type="number"
          className={`form-input${errors.capacityMedium ? ' error' : ''}`}
          {...register('capacityMedium')}
          min={0}
        />
        {errors.capacityMedium && <span className="form-error">{errors.capacityMedium.message}</span>}
      </div>
    </Modal>
  );
}
