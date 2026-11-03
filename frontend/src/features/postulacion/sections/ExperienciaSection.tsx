import type { UseFormReturn } from 'react-hook-form';
import Alert from '@/components/ui/Alert';
import CheckboxField from '@/components/ui/CheckboxField';
import { AreaTexto, Documento, Fecha, Seleccion, Texto } from '../components/campos';
import ListaRepetible from '../components/ListaRepetible';
import SeccionLayout from '../components/SeccionLayout';
import TiempoCalculado, { periodoDe, tiempoTexto } from '../components/TiempoCalculado';
import { labelOf, MAX_MB_DOCUMENTO, TIPO_EXPERIENCIA } from '../constants';
import { useSeccionForm } from '../hooks/useSeccionForm';
import { experienciaItemSchema, experienciaSchema } from '../schemas';
import type { ExperienciaItem, ExperienciaValues } from '../schemas';
import { nuevoId } from '../schemas/comunes';
import { fieldId } from '../schemas/resolver';
import { formatTiempo, haySuperposicion, totalReferencial } from '../utils/experiencia';
import type { Periodo } from '../utils/experiencia';
import { fechaLegible } from '../utils/fechas';

const INICIALES: ExperienciaValues = { experiencias: [] };

function CamposExperiencia({ form }: { form: UseFormReturn<ExperienciaItem> }) {
  const [fechaInicio, fechaFin, actualmente] = form.watch(['fechaInicio', 'fechaFin', 'actualmente']);
  return (
    <>
      <Seleccion form={form} name="tipo" label="Tipo de experiencia" required options={TIPO_EXPERIENCIA} />
      <Texto form={form} name="cargo" label="Cargo o puesto" required autoComplete="organization-title" />
      <Texto form={form} name="institucion" label="Institución o empresa" required autoComplete="organization" />
      <Fecha form={form} name="fechaInicio" label="Fecha de inicio" required />
      <CheckboxField
        {...form.register('actualmente', {
          onChange: (event) => {
            if (event.target.checked) form.setValue('fechaFin', '');
          },
        })}
        id={fieldId('actualmente')}
        label="Trabajo aquí actualmente"
        helperText="Si la marca, el tiempo se calcula hasta hoy."
      />
      {actualmente ? null : <Fecha form={form} name="fechaFin" label="Fecha de fin" required />}
      <TiempoCalculado
        etiqueta="Tiempo de esta experiencia"
        valor={tiempoTexto({ fechaInicio, fechaFin, actualmente })}
      />
      <AreaTexto form={form} name="actividades" label="Actividades relacionadas" required maxLength={500} />
      <AreaTexto form={form} name="funciones" label="Descripción de funciones" required maxLength={1000} />
      <Documento form={form} name="documento" label="Certificado de experiencia en PDF" tipo="EXPERIENCIA" maxSizeMB={MAX_MB_DOCUMENTO} />
    </>
  );
}

export default function ExperienciaSection() {
  const seccion = useSeccionForm('experiencia', experienciaSchema, INICIALES);
  const experiencias = seccion.form.watch('experiencias') ?? [];
  const periodos = experiencias.map(periodoDe).filter((periodo): periodo is Periodo => periodo !== null);

  return (
    <SeccionLayout
      seccion="experiencia"
      {...seccion.layout}
      onContinuar={() => void seccion.guardarYContinuar()}
      instrucciones={
        <p>
          Registre su experiencia profesional. El sistema calcula el tiempo de cada experiencia en años, meses y días.
        </p>
      }
    >
      {periodos.length > 0 ? (
        <div className="rounded-lg border border-line-strong bg-surface p-4 sm:p-6">
          <TiempoCalculado
            etiqueta="Tiempo total de experiencia"
            valor={formatTiempo(totalReferencial(periodos))}
            ayuda="Es un valor referencial; la validación oficial la realiza el BIESS."
          />
        </div>
      ) : null}

      {haySuperposicion(periodos) ? (
        <Alert variant="warning" title="Existen períodos que se superponen.">
          Verifique la información. Puede continuar, pero los períodos superpuestos se revisarán.
        </Alert>
      ) : null}

      <ListaRepetible<ExperienciaItem>
        form={seccion.form}
        name="experiencias"
        schema={experienciaItemSchema}
        singular="experiencia"
        preguntaEliminar="¿Eliminar esta experiencia laboral?"
        vacio="Aún no ha registrado experiencia profesional. Use “Agregar experiencia” para incluir la primera."
        nuevo={() => ({
          id: nuevoId(),
          tipo: '',
          cargo: '',
          institucion: '',
          fechaInicio: '',
          actualmente: false,
          fechaFin: '',
          actividades: '',
          funciones: '',
          documento: null,
        })}
        resumen={(item) => ({
          titulo: `${item.cargo}, ${item.institucion}`,
          detalles: [
            `Experiencia ${labelOf(TIPO_EXPERIENCIA, item.tipo).toLowerCase()}`,
            `Del ${fechaLegible(item.fechaInicio)} ${item.actualmente ? 'hasta la actualidad' : `al ${fechaLegible(item.fechaFin)}`}`,
            `Tiempo: ${tiempoTexto(item) ?? 'sin calcular'}`,
          ],
          documento: item.documento,
        })}
        renderCampos={(form) => <CamposExperiencia form={form} />}
      />
    </SeccionLayout>
  );
}
