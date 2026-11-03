import { forwardRef, useId } from 'react';
import type { SelectHTMLAttributes } from 'react';
import FieldError, { controlBorderClasses, controlFocusClasses, joinIds, RequiredMark } from './FieldMessages';

export interface SelectOption {
  value: string;
  label: string;
}

export interface SelectFieldProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  label: string;
  options: SelectOption[];
  error?: string;
  helperText?: string;
  /** First, empty option. It is a prompt to choose, never a label replacement. */
  emptyOptionLabel?: string;
  containerClassName?: string;
}

const SelectField = forwardRef<HTMLSelectElement, SelectFieldProps>(function SelectField(
  {
    label,
    options,
    error,
    helperText,
    emptyOptionLabel = 'Seleccione una opción',
    containerClassName = '',
    className = '',
    id,
    required,
    ...rest
  },
  ref,
) {
  const generatedId = useId();
  const selectId = id ?? `select-${generatedId}`;
  const helperId = `${selectId}-helper`;
  const errorId = `${selectId}-error`;
  const hasError = Boolean(error);

  return (
    <div className={containerClassName}>
      <label htmlFor={selectId} className="block text-base font-medium text-neutral-900">
        {label}
        {required ? <RequiredMark /> : null}
      </label>
      {helperText ? (
        <p id={helperId} className="mt-1 text-sm text-muted">
          {helperText}
        </p>
      ) : null}
      <select
        {...rest}
        id={selectId}
        ref={ref}
        aria-required={required || undefined}
        aria-invalid={hasError || undefined}
        aria-describedby={joinIds(helperText && helperId, hasError && errorId)}
        className={[
          'mt-2 block h-11 w-full rounded-md bg-surface px-3 text-base text-neutral-900',
          'disabled:cursor-not-allowed disabled:bg-neutral-100 disabled:text-muted',
          controlBorderClasses(hasError),
          controlFocusClasses,
          className,
        ].join(' ')}
      >
        <option value="">{emptyOptionLabel}</option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <FieldError error={error} errorId={errorId} />
    </div>
  );
});

export default SelectField;
