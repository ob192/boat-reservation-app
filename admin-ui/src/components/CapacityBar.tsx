interface CapacityBarProps {
  available: number;
  total: number;
}

export function CapacityBar({ available, total }: CapacityBarProps) {
  if (total === 0) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, flex: 1 }}>
        <div className="slots-bar" style={{ flex: 1 }}>
          <div className="slots-fill fill-full" style={{ width: '100%' }} />
        </div>
        <span className="text-subtle" style={{ fontSize: '0.75rem', minWidth: 36 }}>0/0</span>
      </div>
    );
  }

  const used = total - available;
  const pct = Math.round((used / total) * 100);
  const cls = pct < 50 ? 'fill-good' : pct < 85 ? 'fill-warn' : 'fill-full';

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, flex: 1 }}>
      <div className="slots-bar" style={{ flex: 1 }}>
        <div className={`slots-fill ${cls}`} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-subtle" style={{ fontSize: '0.75rem', minWidth: 36 }}>
        {available}/{total}
      </span>
    </div>
  );
}
