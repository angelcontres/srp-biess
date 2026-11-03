import { Documento, Fecha, Seleccion, Texto } from '../components/campos';
import ListaRepetible from '../components/ListaRepetible';
import SeccionLayout from '../components/SeccionLayout';
import { labelOf, MAX_MB_DOCUMENTO, NIVEL_INSTRUCCION, TIPO_INSTITUCION } from '../constants';
import { useSeccionForm } from '../hooks/useSeccionForm';
import { formacionItemSchema, formacionSchema } from '../schemas';
import type { FormacionItem, FormacionValues } from '../schemas';
import { nuevoId } from '../schemas/comunes';
import { fechaLegible } from '../utils/fechas';

const INICIALES: FormacionValues = { formaciones: [] };

export default function FormacionSection() {
  const seccion = useSeccionForm('formacion', formacionSchema, INICIALES);

  return (
    <SeccionLayout
      seccion="formacion"
      {...seccion.layout}
      onContinuar={() => void seccion.guardarYContinuar()}
      instrucciones={<p>Registre sus títulos académicos. Cada formación debe incluir el título en PDF.</p>}
    >
      <ListaRepetible<FormacionItem>
        form={seccion.form}
        name="formaciones"
        schema={formacionItemSchema}
        singular="formación"
        preguntaEliminar="¿Eliminar esta formación académica?"
        minimo="Debe registrar al menos una formación académica para enviar su postulación."
        vacio="Aún no ha registrado formación académica. Use “Agregar formación” para incluir la primera."
        nuevo={() => ({
          id: nuevoId(),
          nivel: '',
          institucion: '',
          tipoInstitucion: '',
          titulo: '',
          area: '',
          numeroSenescyt: '',
          fechaSenescyt: '',
          documento: null,
        })}
        resumen={(item) => ({
          titulo: item.titulo,
          detalles: [
            `${labelOf(NIVEL_INSTRUCCION, item.nivel)}, ${item.institucion} (${labelOf(TIPO_INSTITUCION, item.tipoInstitucion).toLowerCase()})`,
            `Registro SENESCYT ${item.numeroSenescyt}, del ${fechaLegible(item.fechaSenescyt)}`,
          ],
          documento: item.documento,
        })}
        renderCampos={(form) => (
          <>
            <Seleccion form={form} name="nivel" label="Nivel de instrucción" required options={NIVEL_INSTRUCCION} />
            <Texto form={form} name="institucion" label="Institución educativa" required autoComplete="organization" />
            <Seleccion form={form} name="tipoInstitucion" label="Tipo de institución" required options={TIPO_INSTITUCION} />
            <Texto form={form} name="titulo" label="Título obtenido" required autoComplete="off" />
            {/* TODO(SRP-013): replace with the study-area catalog when it is published (RF-EF12-12). */}
            <Texto form={form} name="area" label="Área de estudios" required autoComplete="off" />
            <Texto form={form} name="numeroSenescyt" label="Número de registro de la SENESCYT" required autoComplete="off" />
            <Fecha form={form} name="fechaSenescyt" label="Fecha de registro de la SENESCYT" required />
            <Documento form={form} name="documento" label="Título en PDF" tipo="TITULO" maxSizeMB={MAX_MB_DOCUMENTO} />
          </>
        )}
      />
    </SeccionLayout>
  );
}
