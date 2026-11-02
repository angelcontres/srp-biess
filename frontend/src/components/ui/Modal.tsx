import { useEffect, useId, useRef } from 'react';
import type { KeyboardEvent, ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { X } from 'lucide-react';

export interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  /**
   * Action buttons. Send the non-destructive option first: on desktop it sits at the
   * left of the primary action, and on mobile the stack is reversed.
   */
  footer?: ReactNode;
  className?: string;
}

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'textarea:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

export default function Modal({
  open,
  onClose,
  title,
  children,
  footer,
  className = '',
}: ModalProps) {
  const titleId = useId();
  const panelRef = useRef<HTMLDivElement>(null);
  const titleRef = useRef<HTMLHeadingElement>(null);
  const previouslyFocused = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!open) return undefined;

    previouslyFocused.current = document.activeElement as HTMLElement | null;
    titleRef.current?.focus();

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';

    return () => {
      document.body.style.overflow = previousOverflow;
      previouslyFocused.current?.focus();
    };
  }, [open]);

  if (!open) return null;

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === 'Escape') {
      event.stopPropagation();
      onClose();
      return;
    }

    if (event.key !== 'Tab') return;

    const focusables = Array.from(
      panelRef.current?.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR) ?? [],
    );

    if (focusables.length === 0) {
      event.preventDefault();
      panelRef.current?.focus();
      return;
    }

    const first = focusables[0];
    const last = focusables[focusables.length - 1];
    const active = document.activeElement;

    if (event.shiftKey && (active === first || active === panelRef.current)) {
      event.preventDefault();
      last?.focus();
    } else if (!event.shiftKey && active === last) {
      event.preventDefault();
      first?.focus();
    }
  };

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex bg-neutral-900/60 sm:items-center sm:justify-center sm:p-4"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose();
      }}
    >
      <div
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        onKeyDown={handleKeyDown}
        className={`flex h-full w-full flex-col bg-surface outline-none sm:h-auto sm:max-w-lg sm:rounded-lg sm:border sm:border-line-strong sm:shadow-lg ${className}`}
      >
        <header className="flex items-start justify-between gap-4 border-b border-line px-4 py-4 sm:px-6">
          <h2
            id={titleId}
            ref={titleRef}
            tabIndex={-1}
            className="text-lg font-semibold text-neutral-900 outline-none"
          >
            {title}
          </h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Cerrar ventana"
            className="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-md text-neutral-700 hover:bg-neutral-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
          >
            <X size={20} strokeWidth={2} aria-hidden="true" />
          </button>
        </header>

        <div className="flex-1 overflow-y-auto px-4 py-6 text-base text-neutral-900 sm:px-6">
          {children}
        </div>

        {footer ? (
          <footer className="flex flex-col-reverse gap-3 border-t border-line px-4 py-4 sm:flex-row sm:items-center sm:justify-end sm:px-6">
            {footer}
          </footer>
        ) : null}
      </div>
    </div>,
    document.body,
  );
}
