import { useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import { Plus } from 'lucide-react';
import { useFieldArray, useForm } from 'react-hook-form';
import type { FieldValues, UseFormReturn } from 'react-hook-form';
import type { ZodTypeAny } from 'zod';
import Alert from '@/components/ui/Alert';
import Badge from '@/components/ui/Badge';
import Button from '@/components/ui/Button';
import FieldError from '@/components/ui/FieldMessages';
import Modal from '@/components/ui/Modal';
import type { UploadedFileInfo } from '@/components/ui/FileUpload';
import { fieldId } from '../schemas/resolver';
import { RequiredNote } from '@/components/ui/FieldMessages';

export interface ResumenItem {
  titulo: string;
  detalles: string[];
  documento?: UploadedFileInfo | null;
}

export interface ListaRepetibleProps<Item extends FieldValues & { id: string }> {
  // The parent form differs per section; the array is addressed by name.
  form: UseFormReturn<any>;
  name: string;
  schema: ZodTypeAny;
  nuevo: () => Item;
  /** "familiar", "formación"... used in buttons and accessible names. */
  singular: string;
  /** Question form for the delete dialog: "¿Eliminar este familiar?" */
  preguntaEliminar: string;
  vacio: string;
  /** Shown before the list when the section requires a minimum (SDD 8.5.5). */
  minimo?: string;
  resumen: (item: Item) => ResumenItem;
  renderCampos: (itemForm: UseFormReturn<Item>) => ReactNode;
  disabled?: boolean;
}

/**
 * Repeatable items (SDD 8.5.5): cards, empty state, "Agregar" at the end, add/edit
 * in a modal (full screen on mobile), delete confirmation naming the item, focus
 * return and polite announcement of the result.
 */
export default function ListaRepetible<Item extends FieldValues & { id: string }>({
  form,
  name,
  schema,
  nuevo,
  singular,
  preguntaEliminar,
  vacio,
  minimo,
  resumen,
  renderCampos,
  disabled = false,
}: ListaRepetibleProps<Item>) {
  const { fields, append, update, remove } = useFieldArray({ control: form.control, name, keyName: '_key' });
  const items = fields as unknown as Item[];
  const [editando, setEditando] = useState<{ index: number | null; valores: Item } | null>(null);
  const [eliminando, setEliminando] = useState<number | null>(null);
  const [anuncio, setAnuncio] = useState('');
  const agregarRef = useRef<HTMLButtonElement>(null);
  const cancelarEliminarRef = useRef<HTMLButtonElement>(null);
  const listError = (form.formState.errors[name] as { message?: string; root?: { message?: string } } | undefined);
  const errorLista = listError?.root?.message ?? listError?.message;

  const enfocarItem = (id: string) => {
    window.setTimeout(() => document.getElementById(`item-${id}`)?.focus(), 0);
  };

  const guardarItem = (valores: Item) => {
    const titulo = resumen(valores).titulo;
    if (editando?.index === null || editando?.index === undefined) {
      append(valores as never);
      setAnuncio(`Se agregó: ${titulo}. Recuerde guardar la sección.`);
    } else {
      update(editando.index, valores as never);
      setAnuncio(`Se actualizó: ${titulo}. Recuerde guardar la sección.`);
    }
    void form.trigger(name);
    setEditando(null);
    enfocarItem(valores.id);
  };

  const confirmarEliminar = () => {
    if (eliminando === null) return;
    const titulo = items[eliminando] ? resumen(items[eliminando]).titulo : '';
    remove(eliminando);
    setEliminando(null);
    setAnuncio(`Se eliminó: ${titulo}. Recuerde guardar la sección.`);
    window.setTimeout(() => agregarRef.current?.focus(), 0);
  };

  const itemEliminar = eliminando !== null ? items[eliminando] : undefined;

  return (
    <div id={fieldId(name)} tabIndex={-1} className="space-y-4 outline-none">
      {minimo ? <p className="text-base text-neutral-700">{minimo}</p> : null}

      {items.length === 0 ? (
        <p className="text-base text-neutral-700">{vacio}</p>
      ) : (
        <ul className="space-y-4">
          {items.map((item, index) => {
            const info = resumen(item);
            return (
              <li key={item.id} className="rounded-lg border border-line-strong bg-surface p-4 sm:p-6">
                <h3 id={`item-${item.id}`} tabIndex={-1} className="text-base font-semibold text-neutral-900 outline-none">
                  {info.titulo}
                </h3>
                {info.detalles.length > 0 ? (
                  <ul className="mt-1 space-y-0.5 text-sm text-muted">
                    {info.detalles.filter(Boolean).map((detalle) => (
                      <li key={detalle}>{detalle}</li>
                    ))}
                  </ul>
                ) : null}
                {info.documento !== undefined ? (
                  <div className="mt-2">
                    {info.documento ? (
                      <Badge variant="complete">Documento adjunto: {info.documento.name}</Badge>
                    ) : (
                      <Badge variant="incomplete">Falta el documento PDF</Badge>
                    )}
                  </div>
                ) : null}
                {disabled ? null : (
                  <div className="mt-3 flex flex-wrap gap-x-6 gap-y-1">
                    <Button
                      variant="tertiary"
                      className="min-h-11"
                      aria-label={`Editar ${singular}: ${info.titulo}`}
                      onClick={() => setEditando({ index, valores: item })}
                    >
                      Editar
                    </Button>
                    <Button
                      variant="tertiary"
                      className="min-h-11"
                      aria-label={`Eliminar ${singular}: ${info.titulo}`}
                      onClick={() => setEliminando(index)}
                    >
                      Eliminar
                    </Button>
                  </div>
                )}
              </li>
            );
          })}
        </ul>
      )}

      <FieldError error={errorLista} errorId={`${fieldId(name)}-error`} />

      {disabled ? null : (
        <Button
          ref={agregarRef}
          variant="secondary"
          leftIcon={<Plus size={20} strokeWidth={2} aria-hidden="true" />}
          onClick={() => setEditando({ index: null, valores: nuevo() })}
        >
          Agregar {singular}
        </Button>
      )}

      <p aria-live="polite" className="sr-only">
        {anuncio}
      </p>

      {editando ? (
        <ItemModal
          titulo={editando.index === null ? `Agregar ${singular}` : `Editar ${singular}`}
          etiquetaGuardar={`Guardar ${singular}`}
          schema={schema}
          valores={editando.valores}
          onCancelar={() => setEditando(null)}
          onGuardar={guardarItem}
          renderCampos={renderCampos}
        />
      ) : null}

      <Modal
        open={eliminando !== null}
        onClose={() => setEliminando(null)}
        title={preguntaEliminar}
        initialFocusRef={cancelarEliminarRef}
        footer={
          <>
            <Button ref={cancelarEliminarRef} variant="secondary" onClick={() => setEliminando(null)}>
              Cancelar
            </Button>
            <Button variant="danger" onClick={confirmarEliminar}>
              Eliminar
            </Button>
          </>
        }
      >
        <p>
          Se eliminará: {itemEliminar ? resumen(itemEliminar).titulo : ''}. No podrá recuperarlo después de guardar la
          sección.
        </p>
      </Modal>
    </div>
  );
}

interface ItemModalProps<Item extends FieldValues> {
  titulo: string;
  etiquetaGuardar: string;
  schema: ZodTypeAny;
  valores: Item;
  onCancelar: () => void;
  onGuardar: (valores: Item) => void;
  renderCampos: (itemForm: UseFormReturn<Item>) => ReactNode;
}

function ItemModal<Item extends FieldValues>({
  titulo,
  etiquetaGuardar,
  schema,
  valores,
  onCancelar,
  onGuardar,
  renderCampos,
}: ItemModalProps<Item>) {
  const itemForm = useForm<Item>({
    resolver: zodResolver(schema),
    mode: 'onBlur',
    reValidateMode: 'onBlur',
    defaultValues: valores as never,
  });
  const [confirmarDescartar, setConfirmarDescartar] = useState(false);
  const formId = `form-${titulo.replace(/\s+/g, '-')}`;

  // Closing with typed data asks first (SDD 8.8: Esc closes unless data would be lost).
  const cerrar = () => {
    if (itemForm.formState.isDirty) setConfirmarDescartar(true);
    else onCancelar();
  };

  return (
    <Modal
      open
      onClose={cerrar}
      title={titulo}
      footer={
        confirmarDescartar ? null : (
          <>
            <Button variant="secondary" onClick={cerrar}>
              Cancelar
            </Button>
            <Button type="submit" form={formId}>
              {etiquetaGuardar}
            </Button>
          </>
        )
      }
    >
      {confirmarDescartar ? (
        <Alert variant="warning" title="Tiene datos sin guardar en este formulario.">
          <p>Si cierra ahora, se perderán.</p>
          <div className="mt-3 flex flex-col gap-3 sm:flex-row">
            <Button variant="secondary" onClick={() => setConfirmarDescartar(false)}>
              Seguir editando
            </Button>
            <Button variant="danger" onClick={onCancelar}>
              Descartar datos
            </Button>
          </div>
        </Alert>
      ) : (
        <form
          id={formId}
          noValidate
          className="space-y-6"
          onSubmit={(event) => {
            event.stopPropagation();
            void itemForm.handleSubmit(onGuardar)(event);
          }}
        >
          <RequiredNote />
          {renderCampos(itemForm)}
        </form>
      )}
    </Modal>
  );
}
