import { useEffect, useRef } from 'react';

export default function Modal({
  onClose,
  children,
  ariaLabel,
  ariaDescribedby,
}) {
  const overlayRef = useRef(null);
  const contentRef = useRef(null);
  const previousFocusRef = useRef(null);
  const dialogRef = useRef(null);

  useEffect(() => {
    previousFocusRef.current = document.activeElement;
    const dialog = dialogRef.current;
    if (dialog && typeof dialog.showModal === 'function') {
      dialog.showModal();
    }

    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        e.stopPropagation();
        onClose();
        return;
      }

      if (e.key === 'Tab' && contentRef.current) {
        const focusable = contentRef.current.querySelectorAll(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        );
        if (focusable.length === 0) return;

        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        const isFirst = document.activeElement === first;
        const isLast = document.activeElement === last;

        if (e.shiftKey && isFirst) {
          e.preventDefault();
          last.focus();
        } else if (!e.shiftKey && isLast) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    const firstFocusable = contentRef.current?.querySelector(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    firstFocusable?.focus();

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      previousFocusRef.current?.focus();
    };
  }, [onClose]);

  const handleOverlayClick = (e) => {
    if (e.target === overlayRef.current) {
      onClose();
    }
  };

  return (
    <dialog
      ref={dialogRef}
      className='modal'
      aria-label={ariaLabel}
      aria-describedby={ariaDescribedby}
    >
      <button
        ref={overlayRef}
        type='button'
        className='modal-overlay'
        onClick={handleOverlayClick}
        aria-hidden='true'
        tabIndex={-1}
      />
      <div ref={contentRef} className='modal-content'>
        {children}
      </div>
    </dialog>
  );
}
