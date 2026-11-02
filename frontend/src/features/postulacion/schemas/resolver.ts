import { zodResolver } from '@hookform/resolvers/zod';
import type { FieldErrors, FieldValues, Resolver } from 'react-hook-form';
import type { MutableRefObject } from 'react';
import type { ZodTypeAny } from 'zod';

/**
 * "progreso": Guardar progreso. Incomplete sections may be saved, but no invalid
 *   value is persisted (RF-EF12-02): errors on empty values are ignored.
 * "completo": Guardar y continuar. Every rule applies.
 */
export type ModoValidacion = 'progreso' | 'completo';

function isEmpty(value: unknown): boolean {
  return (
    value === undefined ||
    value === null ||
    value === false ||
    (typeof value === 'string' && value.trim() === '') ||
    (Array.isArray(value) && value.length === 0)
  );
}

function getAt(source: unknown, path: (string | number)[]): unknown {
  return path.reduce<unknown>(
    (current, key) => (current && typeof current === 'object' ? (current as Record<string, unknown>)[key] : undefined),
    source,
  );
}

function isLeafError(node: unknown): boolean {
  return Boolean(node && typeof node === 'object' && 'message' in node && 'type' in node);
}

/** Removes errors whose field is still empty. Returns null when nothing is left. */
function pruneEmptyErrors(node: unknown, values: unknown, path: (string | number)[] = []): unknown {
  if (isLeafError(node)) return isEmpty(getAt(values, path)) ? null : node;
  if (!node || typeof node !== 'object') return null;

  // `root` holds array-level errors (useFieldArray): it is checked against the array itself.
  const entries = Object.entries(node as Record<string, unknown>)
    .filter(([key]) => key !== 'ref')
    .map(([key, child]) => {
      const childPath = key === 'root' ? path : [...path, Array.isArray(node) ? Number(key) : key];
      return [key, pruneEmptyErrors(child, values, childPath)] as const;
    })
    .filter(([, child]) => child !== null);

  if (entries.length === 0) return null;
  if (Array.isArray(node)) {
    const pruned: unknown[] = [];
    entries.forEach(([key, child]) => {
      pruned[Number(key)] = child;
    });
    return pruned;
  }
  return Object.fromEntries(entries);
}

/** zodResolver that relaxes "required" rules while saving progress. */
export function seccionResolver<T extends FieldValues>(
  schema: ZodTypeAny,
  modoRef: MutableRefObject<ModoValidacion>,
): Resolver<T> {
  const strict = zodResolver(schema) as unknown as Resolver<T>;
  return async (values, context, options) => {
    const result = await strict(values, context, options);
    if (modoRef.current === 'completo' || Object.keys(result.errors).length === 0) return result;
    const errors = (pruneEmptyErrors(result.errors, values) ?? {}) as FieldErrors<T>;
    return Object.keys(errors).length > 0 ? { values: {}, errors } : { values, errors: {} };
  };
}

/** Same rule for restoring a local draft: only format errors discard it. */
export function borradorValido(schema: ZodTypeAny, values: unknown): boolean {
  const result = schema.safeParse(values);
  if (result.success) return true;
  return result.error.issues.every((issue) => isEmpty(getAt(values, issue.path)));
}

export interface ErrorResumen {
  id: string;
  message: string;
}

/** DOM id for a form path: "familiares.0.cargo" -> "f-familiares-0-cargo". */
export function fieldId(path: string): string {
  return `f-${path.replace(/\./g, '-')}`;
}

/** Flattens RHF errors into the error summary list, in field order. */
export function aplanarErrores(errors: FieldErrors, prefix = ''): ErrorResumen[] {
  return Object.entries(errors).flatMap(([key, node]) => {
    if (!node || key === 'ref') return [];
    const path = key === 'root' ? prefix : prefix ? `${prefix}.${key}` : key;
    if (isLeafError(node)) {
      const message = (node as { message?: unknown }).message;
      return typeof message === 'string' && message ? [{ id: fieldId(path), message }] : [];
    }
    return aplanarErrores(node as FieldErrors, path);
  });
}
