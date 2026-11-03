import { useEffect, useRef, useState } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import Alert from '@/components/ui/Alert';
import Badge from '@/components/ui/Badge';
import type { BadgeVariant } from '@/components/ui/Badge';
import Button from '@/components/ui/Button';
import CheckboxField from '@/components/ui/CheckboxField';
import Modal from '@/components/ui/Modal';
import Skeleton from '@/components/ui/Skeleton';
import ErrorSummary from '../components/ErrorSummary';
import SectionBar from '../components/SectionBar';
import { CARGOS, CONSENTIMIENTOS, SECCIONES } from '../constants';
import { usePostulacion } from '../PostulacionContext';
import { resumenSchema } from '../schemas';
import type { ResumenValues } from '../schemas';
import { aplanarErrores, fieldId } from '../schemas/resolver';
import type { ErrorResumen } from '../schemas/resolver';
import type { EnvioResultado, EstadoSeccion, SeccionEditable, ValoresSeccion } from '../types';
import { filasDeSeccion } from './resumenDatos';
import { RequiredNote } from '@/components/ui/FieldMessages';

const EDITABLES = SECCIONES.filter((item) => item.id !== 'resumen') as Array<
  (typeof SECCIONES)[number] & { id: SeccionEditable }
>;

const BADGE: Record<EstadoSeccion, { variant: BadgeVariant; texto: string }> = {
  COMPLETA: { variant: 'complete', texto: 'Completa' },
  INCOMPLETA: { variant: 'incomplete', texto: 'Incompleta' },
  CON_ERRORES: { variant: 'error', texto: 'Con errores' },
  SIN_INICIAR: { variant: 'neutral', texto: 'Sin iniciar' },
};

const INICIALES = Object.fromEntries(CONSENTIMIENTOS.map(({ id }) => [id, false])) as unknown as ResumenValues;

export default function ResumenSection({ onEnviada }: { onEnviada: (resultado: EnvioResultado) => void }) {
  const { gateway, postulacion, irA } = usePostulacion();
  const form = useForm<ResumenValues>({
    resolver: zodResolver(resumenSchema),
    mode: 'onSubmit',
    defaultValues: INICIALES,
  });
  const [datos, setDatos] = useState<Partial<ValoresSeccion> | null>(null);
  const [errores, setErrores] = useState<ErrorResumen[]>([]);
  const [intento, setIntento] = useState(0);
  const [confirmando, setConfirmando] = useState(false);
  const [enviando, setEnviando] = useState(false);
  const [errorEnvio, setErrorEnvio] = useState(false);
  const seguirRevisandoRef = useRef<HTMLButtonElement>(null);
  const cargo = CARGOS.find((item) => item.id === postulacion.elegibilidad?.cargoId);

  useEffect(() => {
    let activo = true;
    Promise.all(EDITABLES.map(async ({ id }) => [id, await gateway.obtenerSeccion(postulacion.id, id)] as const))
      .then((entradas) => activo && setDatos(Object.fromEntries(entradas) as Partial<ValoresSeccion>))
      .catch(() => activo && setDatos({}));
    return () => {
      activo = false;
    };
  }, [gateway, postulacion.id]);

  const pendientes = EDITABLES.filter(({ id }) => postulacion.secciones[id] !== 'COMPLETA' && id !== 'capacitacion');

  const intentarEnviar = form.handleSubmit(
    () => {
      const faltan = pendientes.map(({ id, titulo }) => ({
        id: `corregir-${id}`,
        message: `Complete la sección ${titulo}.`,
      }));
      if (faltan.length > 0) {
        setErrores(faltan);
        setIntento((value) => value + 1);
        return;
      }
      setErrores([]);
      setConfirmando(true);
    },
    (fieldErrors) => {
      const faltan = pendientes.map(({ id, titulo }) => ({ id: `corregir-${id}`, message: `Complete la sección ${titulo}.` }));
      setErrores([...faltan, ...aplanarErrores(fieldErrors)]);
      setIntento((value) => value + 1);
    },
  );

  const enviar = async () => {
    setEnviando(true);
    setErrorEnvio(false);
    try {
      const resultado = await gateway.enviar(postulacion.id, postulacion.version);
      setConfirmando(false);
      onEnviada(resultado);
    } catch {
      setConfirmando(false);
      setErrorEnvio(true);
    } finally {
      setEnviando(false);
    }
  };

  return (
    <>
      <SectionBar seccion="resumen" />
      <div className="mx-auto max-w-5xl space-y-8 px-4 py-8 sm:px-6">
        {errorEnvio ? (
          <Alert variant="danger" title="La postulación no fue enviada.">
            Ocurrió un problema en el sistema. Su información no se perdió. Vuelva a presionar “Enviar postulación”. Si
            persiste, contacte a soporte.
          </Alert>
        ) : null}

        <ErrorSummary errores={errores} intento={intento} />

        <p className="max-w-prose text-base text-neutral-700">Revise su información. Después del envío no podrá modificarla.</p>

        <section aria-labelledby="estado-secciones" className="space-y-4">
          <h2 id="estado-secciones" className="text-lg font-semibold text-neutral-900">
            Estado de las secciones
          </h2>
          {cargo ? <p className="text-base text-neutral-900">Cargo seleccionado: {cargo.nombre}.</p> : null}
          {datos === null ? (
            <Skeleton fields={3} label="Cargando el resumen" />
          ) : (
            <ul className="divide-y divide-line rounded-lg border border-line-strong bg-surface">
              {EDITABLES.map(({ id, titulo }) => {
                const badge = BADGE[postulacion.secciones[id]];
                return (
                  <li key={id} className="p-4 sm:p-6">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <h3 className="text-base font-semibold text-neutral-900">{titulo}</h3>
                      <div className="flex flex-wrap items-center gap-4">
                        <Badge variant={badge.variant}>{badge.texto}</Badge>
                        <Button
                          id={`corregir-${id}`}
                          variant="tertiary"
                          className="min-h-11"
                          aria-label={`Corregir la sección ${titulo}`}
                          onClick={() => irA(id)}
                        >
                          Corregir
                        </Button>
                      </div>
                    </div>
                    <details className="mt-2">
                      <summary className="flex min-h-11 cursor-pointer items-center text-base text-primary underline underline-offset-4">
                        Ver los datos de {titulo.toLowerCase()}
                      </summary>
                      <dl className="mt-2 space-y-2">
                        {filasDeSeccion(id, datos[id] ?? null).map(([etiqueta, valor], index) => (
                          <div key={`${etiqueta}-${index}`}>
                            <dt className="text-sm text-muted">{etiqueta}</dt>
                            <dd className="text-base text-neutral-900">{valor}</dd>
                          </div>
                        ))}
                      </dl>
                    </details>
                  </li>
                );
              })}
            </ul>
          )}
        </section>

        <form
          noValidate
          className="space-y-8"
          onSubmit={(event) => {
            event.preventDefault();
            void intentarEnviar();
          }}
        >
          <fieldset className="space-y-2">
            <legend className="text-lg font-semibold text-neutral-900">Consentimiento informado</legend>
            <p className="text-base text-neutral-700">Para enviar su postulación debe aceptar los cinco consentimientos.</p>
            <RequiredNote />
            {CONSENTIMIENTOS.map(({ id, texto }) => (
              <CheckboxField
                key={id}
                {...form.register(id)}
                id={fieldId(id)}
                required
                label={texto}
                helperText={<a href="#bases-legales">Leer el texto completo</a>}
                error={form.formState.errors[id]?.message}
              />
            ))}
          </fieldset>

          <div className="border-t border-line pt-6">
            <Button type="submit" className="w-full sm:w-auto">
              Enviar postulación
            </Button>
          </div>
        </form>
      </div>

      <Modal
        open={confirmando}
        onClose={() => setConfirmando(false)}
        title="¿Enviar su postulación?"
        initialFocusRef={seguirRevisandoRef}
        footer={
          <>
            <Button ref={seguirRevisandoRef} variant="secondary" onClick={() => setConfirmando(false)}>
              Seguir revisando
            </Button>
            <Button isLoading={enviando} loadingText="Enviando…" onClick={() => void enviar()}>
              Enviar postulación
            </Button>
          </>
        }
      >
        <p>Después del envío no podrá modificar la información ni los documentos.</p>
        {cargo ? <p className="mt-2">Cargo: {cargo.nombre}.</p> : null}
      </Modal>
    </>
  );
}
