import type { ChangeEvent, FocusEvent, KeyboardEvent } from 'react';

export interface DateInputProps {
  /** Current value in dd/mm/aaaa format */
  value: string;
  /** Called when the complete date changes */
  onChange: (value: string) => void;
  /** Input mode for the fields */
  inputMode?: 'numeric' | 'text';
  /** Whether the field is disabled */
  disabled?: boolean;
  /** Error message to display */
  error?: string;
  /** Whether the field is required */
  required?: boolean;
  /** Custom id for the container */
  id?: string;
  /** Aria label for accessibility */
  'aria-label'?: string;
  /** Reference to the container element */
  ref?: React.Ref<HTMLDivElement>;
}

/**
 * Three-input date component (dd/mm/aaaa) with auto-padding and auto-focus.
 * Each input auto-pads with leading zeros and moves focus to the next field when complete.
 */
export default function DateInput({
  value,
  onChange,
  inputMode = 'numeric',
  disabled = false,
  error,
  required = false,
  id,
  'aria-label': ariaLabel,
  ref,
}: DateInputProps) {
  const [dd, mm, aaaa] = value.split('/');

  const handleDdChange = (e: ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value.replace(/\D/g, '').slice(0, 2);
    let newDd = raw;

    // Auto-pad with leading zero
    if (raw.length === 2 && !isNaN(Number(raw))) {
      newDd = raw.padStart(2, '0');
    }

    const newParts = [newDd, mm, aaaa];
    const newValue = newParts.filter(Boolean).join('/');

    onChange(newValue);

    // Auto-focus next field if dd is complete (2 digits)
    if (newDd.length === 2) {
      const mmInput = document.getElementById(`${id}-mm`);
      if (mmInput) (mmInput as HTMLInputElement).focus();
    }
  };

  const handleMmChange = (e: ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value.replace(/\D/g, '').slice(0, 2);
    let newMm = raw;

    // Auto-pad with leading zero
    if (raw.length === 2 && !isNaN(Number(raw))) {
      newMm = raw.padStart(2, '0');
    }

    const newParts = [dd, newMm, aaaa];
    const newValue = newParts.filter(Boolean).join('/');

    onChange(newValue);

    // Auto-focus next field if mm is complete (2 digits)
    if (newMm.length === 2) {
      const aaaaInput = document.getElementById(`${id}-aaaa`);
      if (aaaaInput) (aaaaInput as HTMLInputElement).focus();
    }
  };

  const handleAaaaChange = (e: ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value.replace(/\D/g, '').slice(0, 4);
    const newParts = [dd, mm, raw];
    const newValue = newParts.filter(Boolean).join('/');
    onChange(newValue);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>, nextFieldId?: string) => {
    if (e.key === 'Backspace' && e.currentTarget.value === '' && nextFieldId) {
      const prevInput = document.getElementById(nextFieldId);
      if (prevInput) (prevInput as HTMLInputElement).focus();
    }
  };

  const handleBlur = (e: FocusEvent<HTMLInputElement>, field: 'dd' | 'mm' | 'aaaa') => {
    const raw = e.target.value.replace(/\D/g, '');
    if (raw.length === 1 && !isNaN(Number(raw))) {
      const padded = raw.padStart(2, '0');
      const newParts = field === 'dd' ? [padded, mm, aaaa] : field === 'mm' ? [dd, padded, aaaa] : [dd, mm, padded];
      onChange(newParts.filter(Boolean).join('/'));
    }
  };

  return (
    <div
      id={id}
      ref={ref}
      className="flex items-center gap-2"
      aria-label={ariaLabel || 'Fecha'}
      aria-invalid={!!error}
    >
      {/* Day input */}
      <div className="relative flex-1">
        <input
          id={`${id}-dd`}
          type="text"
          inputMode={inputMode}
          value={dd}
          onChange={handleDdChange}
          onKeyDown={(e) => handleKeyDown(e, undefined)}
          onBlur={(e) => handleBlur(e, 'dd')}
          disabled={disabled}
          placeholder="dd"
          maxLength={2}
          className="w-full rounded-md border border-line-input bg-surface px-3 py-2 text-center text-base tabular-nums focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark disabled:border-line disabled:bg-neutral-100 disabled:text-muted aria-invalid:border-danger"
          aria-label="Día"
          aria-required={required}
        />
      </div>

      {/* Separator */}
      <span className="text-lg font-semibold text-muted" aria-hidden="true">
        /
      </span>

      {/* Month input */}
      <div className="relative flex-1">
        <input
          id={`${id}-mm`}
          type="text"
          inputMode={inputMode}
          value={mm}
          onChange={handleMmChange}
          onKeyDown={(e) => handleKeyDown(e, `${id}-dd`)}
          onBlur={(e) => handleBlur(e, 'mm')}
          disabled={disabled}
          placeholder="mm"
          maxLength={2}
          className="w-full rounded-md border border-line-input bg-surface px-3 py-2 text-center text-base tabular-nums focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark disabled:border-line disabled:bg-neutral-100 disabled:text-muted aria-invalid:border-danger"
          aria-label="Mes"
          aria-required={required}
        />
      </div>

      {/* Separator */}
      <span className="text-lg font-semibold text-muted" aria-hidden="true">
        /
      </span>

      {/* Year input */}
      <div className="relative flex-1">
        <input
          id={`${id}-aaaa`}
          type="text"
          inputMode={inputMode}
          value={aaaa}
          onChange={handleAaaaChange}
          onKeyDown={(e) => handleKeyDown(e, `${id}-mm`)}
          disabled={disabled}
          placeholder="aaaa"
          maxLength={4}
          className="w-full rounded-md border border-line-input bg-surface px-3 py-2 text-center text-base tabular-nums focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark disabled:border-line disabled:bg-neutral-100 disabled:text-muted aria-invalid:border-danger"
          aria-label="Año"
          aria-required={required}
        />
      </div>

      {error && (
        <p className="absolute -bottom-5 left-0 text-sm text-danger" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
