import { useId } from 'react';
import type { ReactNode } from 'react';
import FieldError, { controlFocusClasses, joinIds, RequiredMark } from './FieldMessages';
import type { SelectOption } from './SelectField';

export interface RadioGroupProps {
  legend: ReactNode;
  name: string;
  options: SelectOption[];
  /** Current value. An empty string means no option is selected (no preselection). */
  value: string;
  onChange: (value: string) => void;
  onBlur?: () => void;
  error?: string;
  helperText?: ReactNode;
  /** Id of the fieldset, target of the error summary links. */
  id?: string;
  required?: boolean;
  disabled?: boolean;
  inline?: boolean;
  containerClassName?: string;
}

/**
 * Native radio group inside fieldset/legend. Used for every Yes/No question:
 * toggles are prohibited and nothing is preselected (DESIGN prohibited.md).
 */
export default function RadioGroup({
  legend,
  name,
  options,
  value,
  onChange,
  onBlur,
  error,
  helperText,
  id,
  required,
  disabled,
  inline = true,
  containerClassName = '',
}: RadioGroupProps) {
  const generatedId = useId();
  const groupId = id ?? `radio-${generatedId}`;
  const helperId = `${groupId}-helper`;
  const errorId = `${groupId}-error`;
  const hasError = Boolean(error);

  return (
    <fieldset
      id={groupId}
      tabIndex={-1}
      aria-describedby={joinIds(helperText ? helperId : null, hasError && errorId)}
      className={`outline-none ${containerClassName}`}
    >
      <legend className="text-base font-medium text-neutral-900">
        {legend}
        {required ? <RequiredMark /> : null}
      </legend>
      {helperText ? (
        <div id={helperId} className="mt-1 text-sm text-muted">
          {helperText}
        </div>
      ) : null}
      <div className={`mt-2 flex gap-x-6 gap-y-1 ${inline ? 'flex-row flex-wrap' : 'flex-col'}`}>
        {options.map((option) => {
          const optionId = `${groupId}-${option.value}`;
          return (
            <label
              key={option.value}
              htmlFor={optionId}
              className="flex min-h-11 cursor-pointer items-center gap-3 text-base text-neutral-900"
            >
              <input
                id={optionId}
                type="radio"
                name={name}
                value={option.value}
                checked={value === option.value}
                onChange={() => onChange(option.value)}
                onBlur={onBlur}
                disabled={disabled}
                aria-required={required || undefined}
                aria-invalid={hasError || undefined}
                className={`h-5 w-5 shrink-0 accent-primary ${controlFocusClasses}`}
              />
              {option.label}
            </label>
          );
        })}
      </div>
      <FieldError error={error} errorId={errorId} />
    </fieldset>
  );
}
