import { useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { CircleCheck, Info } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import Modal from '@/components/ui/Modal';
import { formatDateTime } from '@/utils/format';
import { Documento, PreguntaSiNo, Texto } from '../components/campos';
import SeccionLayout from '../components/SeccionLayout';
import SectionBar from '../components/SectionBar';
import { CARGOS, MAX_MB_RESPALDO } from '../constants';
import { useSeccionForm } from '../hooks/useSeccionForm';
import { usePostulacion } from '../PostulacionContext';
import { respaldoSchema } from '../schemas';
import type { RespaldoValues } from '../schemas';
import { fieldId } from '../schemas/resolver';
import { RequiredMark, RequiredNote } from '@/components/ui/FieldMessages';

const INICIALES: RespaldoValues = { tieneRespaldo: '', nombreOrganizacion: '', documento: null };

function AvisoInstitucional({ titulo, children }: { titulo: string; children: ReactNode }) {
  return (
    <div className="flex gap-3 rounded-md border border-primary border-l-4 bg-info-50 p-4">
      <Info size={20} strokeWidth={2} aria-hidden="true" className="shrink-0 text-primary" />
      <div className="max-w-prose space-y-1 text-base text-neutral-900">
        <p className="font-semibold">{titulo}</p>
        {children}
      </div>
    </div>
  );
}

/** Before confirming: three selectable cards and a definitive-choice confirmation (SDD 8.5.6). */
function SeleccionCargo() {
  const { gateway, postulacion, actualizarPostulacion, mostrarElegibilidad } = usePostulacion();
  const [seleccion, setSeleccion] = useState<number | null>(null);
  const [error, setError] = useState('');
  const [confirmando, setConfirmando] = useState(false);
  const [enviando, setEnviando] = useState(false);
  const [errorServidor, setErrorServidor] = useState(false);
  const revisarRef = useRef<HTMLButtonElement>(null);
  const cargo = CARGOS.find((item) => item.id === seleccion);

  const confirmar = async () => {
    if (!cargo) return;
    setEnviando(true);
    setErrorServidor(false);
    try {
      const { postulacion: actualizada, elegibilidad } = await gateway.confirmarCargo(postulacion.id, cargo.id, postulacion.version);
      actualizarPostulacion(actualizada);
      setConfirmando(false);
      mostrarElegibilidad(elegibilidad);
    } catch {
      // RF-EF12-10: external service down. Not marked NO_ELEGIBLE; retry allowed.
      setConfirmando(false);
      setErrorServidor(true);
    } finally {
      setEnviando(false);
    }
  };

  return (
    <>
      <SectionBar seccion="cargo" />
      <div className="mx-auto max-w-5xl space-y-8 px-4 py-8 sm:px-6">
        {errorServidor ? (
          <Alert variant="warning" title="No es posible validar su calidad laboral en este momento.">
            Su selección no se guardó. Inténtelo de nuevo en unos minutos.
          </Alert>
        ) : null}
        <AvisoInstitucional titulo="Selección definitiva">
          <p>La selección del cargo es definitiva para esta postulación.</p>
          <p className="text-sm text-muted">[PENDIENTE: requisitos clave de cada cargo según las bases del concurso]</p>
        </AvisoInstitucional>

        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault();
            if (!seleccion) {
              setError('Seleccione el cargo al que desea postular.');
              document.getElementById(fieldId('cargo'))?.focus();
              return;
            }
            setConfirmando(true);
          }}
          className="space-y-8"
        >
          <RequiredNote />
          <fieldset id={fieldId('cargo')} tabIndex={-1} aria-describedby={error ? 'cargo-error' : undefined} className="outline-none">
            <legend className="text-base font-medium text-neutral-900">
              Cargo al que postula<RequiredMark />
            </legend>
            <div className="mt-4 grid gap-4 lg:grid-cols-3">
              {CARGOS.map((item) => {
                const activo = seleccion === item.id;
                return (
                  <label
                    key={item.id}
                    className={`flex cursor-pointer flex-col gap-3 rounded-lg border-2 p-4 sm:p-6 ${
                      activo ? 'border-primary bg-info-50' : 'border-line-strong bg-surface'
                    }`}
                  >
                    <span className="flex items-start gap-3">
                      <input
                        type="radio"
                        name="cargo"
                        value={item.id}
                        checked={activo}
                        onChange={() => {
                          setSeleccion(item.id);
                          setError('');
                        }}
                        className="mt-0.5 h-5 w-5 shrink-0 accent-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
                      />
                      <span className="text-base font-semibold text-neutral-900">{item.nombre}</span>
                    </span>
                    {activo ? (
                      <span className="inline-flex items-center gap-1 text-sm font-medium text-primary">
                        <CircleCheck size={16} strokeWidth={2} aria-hidden="true" />
                        Seleccionado
                      </span>
                    ) : null}
                  </label>
                );
              })}
            </div>
            {error ? (
              <p id="cargo-error" role="alert" className="mt-2 text-base text-danger">
                {error}
              </p>
            ) : null}
          </fieldset>

          <div className="border-t border-line pt-6">
            <Button type="submit" className="w-full sm:w-auto">
              Confirmar cargo
            </Button>
          </div>
        </form>
      </div>

      <Modal
        open={confirmando}
        onClose={() => setConfirmando(false)}
        title="¿Confirmar el cargo seleccionado?"
        initialFocusRef={revisarRef}
        footer={
          <>
            <Button ref={revisarRef} variant="secondary" onClick={() => setConfirmando(false)}>
              Revisar otra vez
            </Button>
            <Button isLoading={enviando} loadingText="Validando…" onClick={() => void confirmar()}>
              Confirmar cargo
            </Button>
          </>
        }
      >
        <p>Cargo: {cargo?.nombre}.</p>
        <p className="mt-2">
          La selección es definitiva para esta postulación. Después de confirmar, se validará su elegibilidad y no podrá
          cambiar de cargo.
        </p>
      </Modal>
    </>
  );
}

/** After eligibility: organizational support for the sector (RF-EF12-11). */
function Respaldo() {
  const { postulacion } = usePostulacion();
  const seccion = useSeccionForm('cargo', respaldoSchema, INICIALES);
  const { form } = seccion;
  const elegibilidad = postulacion.elegibilidad;
  const cargo = CARGOS.find((item) => item.id === elegibilidad?.cargoId);
  const elegible = elegibilidad?.elegible === true;
  const tieneRespaldo = form.watch('tieneRespaldo');

  return (
    <SeccionLayout
      seccion="cargo"
      {...seccion.layout}
      soloLectura={!elegible}
      onContinuar={() => void seccion.guardarYContinuar()}
    >
      <AvisoInstitucional titulo="Cargo seleccionado">
        <p>{cargo?.nombre}</p>
        {elegibilidad ? (
          <>
            <p>{elegibilidad.mensaje}</p>
            <p className="text-sm tabular-nums text-muted">
              Resultado de elegibilidad del {formatDateTime(elegibilidad.fecha)} (hora de Ecuador).
            </p>
          </>
        ) : null}
      </AvisoInstitucional>

      {elegible && cargo ? (
        <section aria-labelledby="respaldo-titulo" className="space-y-6">
          <h2 id="respaldo-titulo" className="text-lg font-semibold text-neutral-900">
            Respaldo organizacional
          </h2>
          <PreguntaSiNo
            form={form}
            name="tieneRespaldo"
            legend={`¿Está respaldado por ${cargo.respaldo}?`}
            onValueChange={(value) => {
              if (value === 'NO') {
                form.setValue('nombreOrganizacion', '', { shouldDirty: true });
                form.setValue('documento', null, { shouldDirty: true });
              }
            }}
          />
          {tieneRespaldo === 'SI' ? (
            <div className="space-y-6 border-l-4 border-line pl-4">
              <Texto form={form} name="nombreOrganizacion" label="Nombre de la organización" required autoComplete="organization" />
              <Documento
                form={form}
                name="documento"
                label="Documento de respaldo"
                tipo="RESPALDO_ORGANIZACIONAL"
                maxSizeMB={MAX_MB_RESPALDO}
                contentHint="[PENDIENTE: contenido exigido del documento de respaldo según las bases]"
              />
            </div>
          ) : null}
        </section>
      ) : null}
    </SeccionLayout>
  );
}

export default function CargoSection() {
  const { postulacion } = usePostulacion();
  // SDD 8.3: once confirmed, the position cannot be changed from this section.
  return postulacion.elegibilidad ? <Respaldo /> : <SeleccionCargo />;
}
