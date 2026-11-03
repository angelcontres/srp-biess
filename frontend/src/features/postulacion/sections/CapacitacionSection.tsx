import { Documento, Fecha, Seleccion, Texto } from '../components/campos';
import ListaRepetible from '../components/ListaRepetible';
import SeccionLayout from '../components/SeccionLayout';
import { labelOf, MAX_MB_DOCUMENTO, TIPO_EVENTO } from '../constants';
import { useSeccionForm } from '../hooks/useSeccionForm';
import { capacitacionItemSchema, capacitacionSchema } from '../schemas';
import type { CapacitacionItem, CapacitacionValues } from '../schemas';
import { nuevoId } from '../schemas/comunes';
import { fechaLegible } from '../utils/fechas';

const INICIALES: CapacitacionValues = { capacitaciones: [] };

export default function CapacitacionSection() {
  const seccion = useSeccionForm('capacitacion', capacitacionSchema, INICIALES);

  return (
    <SeccionLayout
      seccion="capacitacion"
      {...seccion.layout}
      onContinuar={() => void seccion.guardarYContinuar()}
      instrucciones={
        <p>
          Registre los cursos, seminarios y talleres en los que participó. Esta sección es opcional, pero cada registro
          debe estar completo e incluir su certificado en PDF.
        </p>
      }
    >
      <ListaRepetible<CapacitacionItem>
        form={seccion.form}
        name="capacitaciones"
        schema={capacitacionItemSchema}
        singular="capacitación"
        preguntaEliminar="¿Eliminar esta capacitación?"
        vacio="Aún no ha registrado capacitaciones. Use “Agregar capacitación” para incluir la primera."
        nuevo={() => ({
          id: nuevoId(),
          tipoEvento: '',
          nombre: '',
          institucion: '',
          fechaInicio: '',
          fechaFin: '',
          horas: '',
          documento: null,
        })}
        resumen={(item) => ({
          titulo: item.nombre,
          detalles: [
            `${labelOf(TIPO_EVENTO, item.tipoEvento)}, ${item.institucion}`,
            `Del ${fechaLegible(item.fechaInicio)} al ${fechaLegible(item.fechaFin)}, ${item.horas} horas`,
          ],
          documento: item.documento,
        })}
        renderCampos={(form) => (
          <>
            <Seleccion form={form} name="tipoEvento" label="Tipo de evento" required options={TIPO_EVENTO} />
            <Texto form={form} name="nombre" label="Nombre del evento" required autoComplete="off" />
            <Texto form={form} name="institucion" label="Institución organizadora" required autoComplete="organization" />
            <div className="grid gap-6 sm:grid-cols-2">
              <Fecha form={form} name="fechaInicio" label="Fecha de inicio" required />
              <Fecha form={form} name="fechaFin" label="Fecha de fin" required />
            </div>
            <Texto form={form} name="horas" label="Número de horas" required numeric maxLength={5} autoComplete="off" />
            <Documento form={form} name="documento" label="Certificado en PDF" tipo="CAPACITACION" maxSizeMB={MAX_MB_DOCUMENTO} />
          </>
        )}
      />
    </SeccionLayout>
  );
}
