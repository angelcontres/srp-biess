import { Info } from 'lucide-react';
import { AreaTexto, Fecha, PreguntaSiNo, Texto } from '../components/campos';
import SeccionLayout from '../components/SeccionLayout';
import { INHABILIDADES } from '../constants';
import { useSeccionForm } from '../hooks/useSeccionForm';
import { inhabilidadesSchema } from '../schemas';
import type { InhabilidadesValues } from '../schemas';
import { fieldId } from '../schemas/resolver';

const INICIALES: InhabilidadesValues = {
  respuestas: INHABILIDADES.map(() => '' as const),
  remocionInstitucion: '',
  remocionCargo: '',
  remocionFechaFin: '',
  remocionMotivo: '',
};

const INDICE_BLOQUEANTE = INHABILIDADES.findIndex((pregunta) => pregunta.efecto === 'BLOQUEANTE');
const BLOQUEANTE = INHABILIDADES[INDICE_BLOQUEANTE];
const MENSAJE_BLOQUEO =
  'Su postulación no puede continuar con esta respuesta. [PENDIENTE: texto jurídico del efecto de la pregunta 1.9.1]';

export default function InhabilidadesSection() {
  const seccion = useSeccionForm('inhabilidades', inhabilidadesSchema, INICIALES);
  const { form } = seccion;
  const respuestas = form.watch('respuestas');
  const bloqueada = BLOQUEANTE !== undefined && respuestas?.[INDICE_BLOQUEANTE] === BLOQUEANTE.respuestaBloqueante;

  return (
    <SeccionLayout
      seccion="inhabilidades"
      {...seccion.layout}
      onContinuar={() =>
        void seccion.guardarYContinuar(() =>
          // SRS 7.4: the blocking answer prevents moving forward; it is not a format error.
          bloqueada ? [{ id: fieldId(`respuestas.${INDICE_BLOQUEANTE}`), message: MENSAJE_BLOQUEO }] : [],
        )
      }
      instrucciones={
        <p>
          [PENDIENTE: texto jurídico que explique qué es una declaración de inhabilidad y su carácter de declaración].
          Responda todas las preguntas. Ninguna respuesta está marcada de antemano.
        </p>
      }
    >
      <ol className="space-y-8">
        {INHABILIDADES.map((pregunta, index) => {
          const respuesta = respuestas?.[index];
          const esBloqueo = pregunta.efecto === 'BLOQUEANTE' && respuesta === pregunta.respuestaBloqueante;
          const esAlerta = pregunta.efecto === 'ALERTA' && respuesta === 'SI';
          return (
            <li key={pregunta.id} className="space-y-4 border-b border-line pb-8 last:border-b-0">
              <PreguntaSiNo
                form={form}
                name={`respuestas.${index}`}
                onValueChange={
                  pregunta.pideRemocion
                    ? (value) => {
                        if (value !== 'NO') return;
                        // Hidden conditional data is not sent (SDD 8.6).
                        (['remocionInstitucion', 'remocionCargo', 'remocionFechaFin', 'remocionMotivo'] as const).forEach(
                          (campo) => form.setValue(campo, '', { shouldDirty: true }),
                        );
                      }
                    : undefined
                }
                legend={
                  <>
                    <span className="tabular-nums">{pregunta.codigo}</span> {pregunta.texto}
                  </>
                }
              />
              {esBloqueo ? (
                <div className="flex gap-3 rounded-md border border-primary border-l-4 bg-info-50 p-4">
                  <Info size={20} strokeWidth={2} aria-hidden="true" className="shrink-0 text-primary" />
                  <p className="text-base text-neutral-900">{MENSAJE_BLOQUEO}</p>
                </div>
              ) : null}
              {esAlerta ? (
                <p className="flex gap-2 text-base text-neutral-700">
                  <Info size={20} strokeWidth={2} aria-hidden="true" className="shrink-0 text-primary" />
                  Puede continuar. Esta respuesta quedará registrada para revisión.
                </p>
              ) : null}
              {pregunta.pideRemocion && respuesta === 'SI' ? (
                <div className="space-y-6 border-l-4 border-line pl-4" aria-live="polite">
                  <p className="text-base text-neutral-700">Complete los datos de la remoción.</p>
                  <Texto form={form} name="remocionInstitucion" label="Institución" required autoComplete="organization" />
                  <Texto form={form} name="remocionCargo" label="Cargo" required autoComplete="organization-title" />
                  <Fecha form={form} name="remocionFechaFin" label="Fecha de finalización" required />
                  <AreaTexto form={form} name="remocionMotivo" label="Motivo de la remoción" required maxLength={500} />
                </div>
              ) : null}
            </li>
          );
        })}
      </ol>
    </SeccionLayout>
  );
}
