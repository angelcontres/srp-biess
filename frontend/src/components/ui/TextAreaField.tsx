import { forwardRef, useId } from 'react';
import type { TextareaHTMLAttributes } from 'react';
import FieldError, { controlBorderClasses, controlFocusClasses, joinIds, RequiredMark } from './FieldMessages';

export interface TextAreaFieldProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: string;
  error?: string;
  helperText?: string;
  /** Current length, used for the counter when `maxLength` is set. */
  currentLength?: number;
  containerClassName?: string;
}

const TextAreaField = forwardRef<HTMLTextAreaElement, TextAreaFieldProps>(function TextAreaField(
  {
    label,
    error,
    helperText,
    currentLength,
    maxLength,
    containerClassName = '',
    className = '',
    id,
    required,
    rows = 5,
    ...rest
  },
  ref,
) {
  const generatedId = useId();
  const fieldId = id ?? `textarea-${generatedId}`;
  const helperId = `${fieldId}-helper`;
  const counterId = `${fieldId}-counter`;
  const errorId = `${fieldId}-error`;
  const hasError = Boolean(error);
  const showCounter = maxLength !== undefined && currentLength !== undefined;

  return (
    <div className={containerClassName}>
      <label htmlFor={fieldId} className="block text-base font-medium text-neutral-900">
        {label}
        {required ? <RequiredMark /> : null}
      </label>
      {helperText ? (
        <p id={helperId} className="mt-1 text-sm text-muted">
          {helperText}
        </p>
      ) : null}
      <textarea
        {...rest}
        id={fieldId}
        ref={ref}
        rows={rows}
        maxLength={maxLength}
        aria-required={required || undefined}
        aria-invalid={hasError || undefined}
        aria-describedby={joinIds(helperText && helperId, showCounter && counterId, hasError && errorId)}
        className={[
          'mt-2 block w-full rounded-md bg-surface px-3 py-2 text-base text-neutral-900',
          controlBorderClasses(hasError),
          controlFocusClasses,
          className,
        ].join(' ')}
      />
      {showCounter ? (
        <p id={counterId} className="mt-1 text-sm tabular-nums text-muted">
          {currentLength} de {maxLength} caracteres
        </p>
      ) : null}
      <FieldError error={error} errorId={errorId} />
    </div>
  );
});

export default TextAreaField;
