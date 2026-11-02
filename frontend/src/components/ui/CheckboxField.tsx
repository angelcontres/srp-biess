import { forwardRef, useId } from 'react';
import type { InputHTMLAttributes, ReactNode } from 'react';
import FieldError, { controlFocusClasses, joinIds, RequiredMark } from './FieldMessages';

export interface CheckboxFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label: ReactNode;
  error?: string;
  helperText?: ReactNode;
  containerClassName?: string;
}

/** Single checkbox with visible label. Never preselected in consents. */
const CheckboxField = forwardRef<HTMLInputElement, CheckboxFieldProps>(function CheckboxField(
  { label, error, helperText, containerClassName = '', id, required, ...rest },
  ref,
) {
  const generatedId = useId();
  const fieldId = id ?? `checkbox-${generatedId}`;
  const helperId = `${fieldId}-helper`;
  const errorId = `${fieldId}-error`;
  const hasError = Boolean(error);

  return (
    <div className={containerClassName}>
      <label htmlFor={fieldId} className="flex min-h-11 cursor-pointer items-start gap-3 py-2">
        <input
          {...rest}
          ref={ref}
          id={fieldId}
          type="checkbox"
          aria-required={required || undefined}
          aria-invalid={hasError || undefined}
          aria-describedby={joinIds(helperText ? helperId : null, hasError && errorId)}
          className={`mt-0.5 h-5 w-5 shrink-0 accent-primary ${controlFocusClasses}`}
        />
        <span className="text-base text-neutral-900">
          {label}
          {required ? <RequiredMark /> : null}
        </span>
      </label>
      {helperText ? (
        <div id={helperId} className="ml-8 text-sm text-muted">
          {helperText}
        </div>
      ) : null}
      <FieldError error={error} errorId={errorId} />
    </div>
  );
});

export default CheckboxField;
