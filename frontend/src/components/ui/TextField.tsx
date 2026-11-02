import { forwardRef, useId } from 'react';
import type { InputHTMLAttributes } from 'react';

export interface TextFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label: string;
  /** Validation message; also marks the field as invalid. */
  error?: string;
  helperText?: string;
  /** Autocompleted or verified values are shown read-only (SDD 6.4). */
  readOnly?: boolean;
  containerClassName?: string;
}

const TextField = forwardRef<HTMLInputElement, TextFieldProps>(function TextField(
  {
    label,
    error,
    helperText,
    readOnly = false,
    containerClassName = '',
    className = '',
    id,
    required,
    disabled,
    ...rest
  },
  ref,
) {
  const generatedId = useId();
  const inputId = id ?? `field-${generatedId}`;
  const errorId = `${inputId}-error`;
  const helperId = `${inputId}-helper`;
  const hasError = Boolean(error);

  const describedBy =
    [hasError ? errorId : null, helperText ? helperId : null].filter(Boolean).join(' ') || undefined;

  return (
    <div className={`flex flex-col gap-1 ${containerClassName}`}>
      <label htmlFor={inputId} className="text-sm font-medium text-neutral-900">
        {label}
        {required ? (
          <span aria-hidden="true" className="ml-0.5 text-danger">
            *
          </span>
        ) : null}
      </label>

      <input
        {...rest}
        id={inputId}
        ref={ref}
        required={required}
        disabled={disabled}
        readOnly={readOnly}
        aria-invalid={hasError || undefined}
        aria-describedby={describedBy}
        className={[
          'w-full rounded-lg border px-3 py-2 text-base text-neutral-900 placeholder:text-neutral-900/40',
          'focus:outline-none focus:ring-2 focus:ring-offset-0',
          'disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-neutral-900/60',
          readOnly ? 'bg-slate-100' : 'bg-white',
          hasError
            ? 'border-danger focus:border-danger focus:ring-danger/30'
            : 'border-slate-300 focus:border-primary focus:ring-primary/30',
          className,
        ]
          .filter(Boolean)
          .join(' ')}
      />

      {helperText && !hasError ? (
        <p id={helperId} className="text-xs text-neutral-900/60">
          {helperText}
        </p>
      ) : null}

      {hasError ? (
        <p id={errorId} role="alert" className="text-xs font-medium text-danger">
          {error}
        </p>
      ) : null}
    </div>
  );
});

export default TextField;
