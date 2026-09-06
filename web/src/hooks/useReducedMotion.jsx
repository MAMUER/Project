import { useEffect, useId, useState } from 'react';

export function useReducedMotion() {
  const [reduced, setReduced] = useState(
    () =>
      typeof window !== 'undefined' &&
      window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
  );

  useEffect(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return;

    const media = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handler = (e) => setReduced(e.matches);
    media.addEventListener('change', handler);
    return () => media.removeEventListener('change', handler);
  }, []);

  return reduced;
}

export function usePauseState(defaultPaused = false) {
  const [paused, setPaused] = useState(defaultPaused);
  const reducedMotion = useReducedMotion();
  const effectivePaused = reducedMotion || paused;

  return { paused, setPaused, effectivePaused, reducedMotion };
}

export function useLivePauseId() {
  return useId();
}

export function PauseOverlay({ onToggle, paused }) {
  return (
    <button
      type='button'
      onClick={onToggle}
      aria-pressed={paused}
      aria-label={paused ? 'Возобновить автоматическое обновление' : 'Остановить автоматическое обновление'}
      style={{
        position: 'absolute',
        top: 8,
        right: 8,
        zIndex: 10,
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--radius-sm)',
        padding: '6px 12px',
        color: 'var(--text-primary)',
        fontSize: 13,
        cursor: 'pointer',
      }}
    >
      {paused ? '▶ Обновление остановлено' : '⏸ Остановить обновление'}
    </button>
  );
}
