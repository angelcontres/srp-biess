import type { ReactNode } from 'react';
import { Controller } from 'react-hook-form';
import type { Control, FieldPath, FieldValues, UseFormReturn } from 'react-hook-form';
import FileUpload from '@/components/ui/FileUpload';
import RadioGroup from '@/components/ui/RadioGroup';
import SelectField from '@/components/ui/SelectField';
import type { SelectOption } from '@/components/ui/SelectField';
import TextAreaField from '@/components/ui/TextAreaField';
import TextField from '@/components/ui/TextField';
import { SI_NO } from '../constants';
import { usePostulacion } from '../PostulacionContext';
import { fieldId } from '../schemas/resolver';
import type { TipoDocumento } from '../types';
import { maskDmy } from '../utils/fechas';

/* Thin bindings between React Hook Form and the UI kit. Ids follow fieldId(),
   which the error summary links to. */

function errorAt<T extends FieldValues>(form: UseFormReturn<T>, name: FieldPath<T>): string | undefined {
  const error = form.getFieldState(name, form.formState).error;
  return error?.message ?? (error as { root?: { message?: string } } | undefined)?.root?.message;
}

interface BaseProps<T extends FieldValues> {
  form: UseFormReturn<T>;
  name: FieldPath<T>;
  label: string;
  required?: boolean;
  helperText?: string;
  className?: string;
}

export function Texto<T extends FieldValues>({
  form,
  name,
  label,
  required,
  helperText,
  className,
  ...input
}: BaseProps<T> & {
  autoComplete?: string;
  inputMode?: 'text' | 'numeric' | 'tel' | 'email';
  type?: 'text' | 'tel' | 'email';
  maxLength?: number;
  numeric?: boolean;
}) {
  const { numeric, ...rest } = input;
  return (
    <TextField
      {...rest}
      {...form.register(name)}
      inputMode={numeric ? 'numeric' : rest.inputMode}
      id={fieldId(name)}
      label={label}
      required={required}
      helperText={helperText}
      error={errorAt(form, name)}
      containerClassName={className}
      className={numeric ? 'tabular-nums' : undefined}
    />
  );
}

export function Fecha<T extends FieldValues>({ form, name, label, required, helperText, className }: BaseProps<T>) {
  return (
    <Controller
      control={form.control as Control<T>}
      name={name}
      render={({ field, fieldState }) => (
        <TextField
          id={fieldId(name)}
          label={label}
          required={required}
          helperText={helperText ?? 'Formato: dd/mm/aaaa, por ejemplo 15/01/2020.'}
          inputMode="numeric"
          autoComplete="off"
          maxLength={10}
          className="tabular-nums"
          containerClassName={className}
          value={field.value ?? ''}
          onChange={(event) => field.onChange(maskDmy(event.target.value))}
          onBlur={field.onBlur}
          name={field.name}
          ref={field.ref}
          error={fieldState.error?.message}
        />
      )}
    />
  );
}

export function Seleccion<T extends FieldValues>({
  form,
  name,
  label,
  required,
  helperText,
  className,
  options,
  autoComplete,
  disabled,
  onValueChange,
}: BaseProps<T> & {
  options: SelectOption[];
  autoComplete?: string;
  disabled?: boolean;
  onValueChange?: (value: string) => void;
}) {
  const registration = form.register(name, {
    onChange: onValueChange ? (event) => onValueChange(String(event.target.value)) : undefined,
  });
  return (
    <SelectField
      {...registration}
      id={fieldId(name)}
      label={label}
      required={required}
      helperText={helperText}
      options={options}
      autoComplete={autoComplete}
      disabled={disabled}
      error={errorAt(form, name)}
      containerClassName={className}
    />
  );
}

export function AreaTexto<T extends FieldValues>({
  form,
  name,
  label,
  required,
  helperText,
  className,
  maxLength,
}: BaseProps<T> & { maxLength?: number }) {
  const value = form.watch(name) as string | undefined;
  return (
    <TextAreaField
      {...form.register(name)}
      id={fieldId(name)}
      label={label}
      required={required}
      helperText={helperText}
      maxLength={maxLength}
      currentLength={value?.length ?? 0}
      error={errorAt(form, name)}
      containerClassName={className}
    />
  );
}

export function PreguntaSiNo<T extends FieldValues>({
  form,
  name,
  legend,
  helperText,
  required = true,
  onValueChange,
}: {
  form: UseFormReturn<T>;
  name: FieldPath<T>;
  legend: ReactNode;
  helperText?: ReactNode;
  required?: boolean;
  onValueChange?: (value: string) => void;
}) {
  return (
    <Controller
      control={form.control as Control<T>}
      name={name}
      render={({ field, fieldState }) => (
        <RadioGroup
          id={fieldId(name)}
          name={field.name}
          legend={legend}
          helperText={helperText}
          required={required}
          options={SI_NO}
          value={(field.value as string) ?? ''}
          onChange={(value) => {
            field.onChange(value);
            onValueChange?.(value);
          }}
          onBlur={field.onBlur}
          error={fieldState.error?.message}
        />
      )}
    />
  );
}

export function Documento<T extends FieldValues>({
  form,
  name,
  label,
  required = true,
  tipo,
  maxSizeMB,
  contentHint,
}: Omit<BaseProps<T>, 'helperText' | 'className'> & { tipo: TipoDocumento; maxSizeMB: number; contentHint?: string }) {
  const { gateway, postulacion } = usePostulacion();
  return (
    <Controller
      control={form.control as Control<T>}
      name={name}
      render={({ field, fieldState }) => (
        <FileUpload
          id={fieldId(name)}
          label={label}
          required={required}
          maxSizeMB={maxSizeMB}
          contentHint={contentHint}
          value={field.value ?? null}
          onChange={(value) => {
            field.onChange(value);
            field.onBlur();
          }}
          onUpload={(file, onProgress, signal) => gateway.subirDocumento(postulacion.id, tipo, file, onProgress, signal)}
          error={fieldState.error?.message}
        />
      )}
    />
  );
}
