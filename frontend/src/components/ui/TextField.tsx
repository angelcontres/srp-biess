import { forwardRef, useId } from 'react';
import type { InputHTMLAttributes, ReactNode } from 'react';
import { CircleAlert, ShieldCheck } from 'lucide-react';
import { RequiredMark } from './FieldMessages';

export interface TextFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label: string;
  /** Validation message; also marks the field as invalid (DESIGN §7.4). */
  error?: string;
  /** Help text rendered between the label and the input. */
  helperText?: string;
  /** Verified value: read-only, neutral background and "Verificado" badge. */
  verified?: boolean;
  containerClassName?: string;
  /** Optional element rendered at the end of the label row, such as a badge. */
  labelAdornment?: ReactNode;
  /** Control rendered inside the input box, aligned to its vertical center (for example the password toggle). */
  trailingAdornment?: ReactNode;
}

const TextField = forwardRef<HTMLInputElement, TextFieldProps>(function TextField(
  {
    label,
    error,
    helperText,
    verified = false,
    containerClassName = '',
    labelAdornment,
    trailingAdornment,
    className = '',
    id,
    required,
    disabled,
    readOnly,
    ...rest
  },
  ref,
) {
  const generatedId = useId();
  const inputId = id ?? `field-${generatedId}`;
  const errorId = `${inputId}-error`;
  const helperId = `${inputId}-helper`;
  const hasError = Boolean(error);
  const isReadOnly = readOnly === true || verified;

  const describedBy =
    [helperText ? helperId : null, hasError ? errorId : null].filter(Boolean).join(' ') || undefined;

  return (
    <div className={containerClassName}>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <label htmlFor={inputId} className="block text-base font-medium text-neutral-900">
          {label}
          {required ? <RequiredMark /> : null}
        </label>
        {labelAdornment ??
          (verified ? (
            <span className="inline-flex items-center gap-1 rounded bg-neutral-100 px-2 py-1 text-sm font-medium text-neutral-700">
              <ShieldCheck size={16} strokeWidth={2} aria-hidden="true" />
              Verificado
            </span>
          ) : null)}
      </div>

      {helperText ? (
        <p id={helperId} className="mt-1 text-sm text-muted">
          {helperText}
        </p>
      ) : null}

      <div className="relative mt-2">
        <input
          {...rest}
          id={inputId}
          ref={ref}
          required={required}
          disabled={disabled}
          readOnly={isReadOnly}
          aria-invalid={hasError || undefined}
          aria-describedby={describedBy}
          className={[
            'block h-11 w-full rounded-md border px-3 text-base text-neutral-900',
            'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark',
            'disabled:cursor-not-allowed disabled:bg-neutral-100 disabled:text-muted',
            trailingAdornment ? 'pr-12' : '',
            hasError ? 'border-2 border-danger' : 'border border-line-input',
            isReadOnly ? 'bg-neutral-100' : 'bg-surface',
            className,
          ]
            .filter(Boolean)
            .join(' ')}
        />

        {trailingAdornment ? (
          <div className="absolute inset-y-0 right-0 flex items-center pr-1">{trailingAdornment}</div>
        ) : null}
      </div>

      {hasError ? (
        <p id={errorId} role="alert" className="mt-2 flex gap-2 text-base text-danger">
          <CircleAlert size={20} strokeWidth={2} aria-hidden="true" className="shrink-0" />
          <span>{error}</span>
        </p>
      ) : null}

      {verified ? (
        <p className="mt-1 text-sm text-muted">Dato verificado. No se puede modificar.</p>
      ) : null}
    </div>
  );
});

export default TextField;
