import { useState } from 'react';
import { PreguntaSiNo, Seleccion, Texto } from '../components/campos';
import ListaRepetible from '../components/ListaRepetible';
import SeccionLayout from '../components/SeccionLayout';
import { INSTITUCION_FAMILIAR, labelOf, PARENTESCO } from '../constants';
import { useSeccionForm } from '../hooks/useSeccionForm';
import { familiaresSchema, familiarSchema } from '../schemas';
import type { FamiliaresValues, FamiliarValues } from '../schemas';
import { nuevoId } from '../schemas/comunes';

const INICIALES: FamiliaresValues = { tieneFamiliares: '', familiares: [] };

export default function FamiliaresSection() {
  const seccion = useSeccionForm('familiares', familiaresSchema, INICIALES);
  const { form } = seccion;
  const [aviso, setAviso] = useState('');
  const tieneFamiliares = form.watch('tieneFamiliares');

  return (
    <SeccionLayout
      seccion="familiares"
      {...seccion.layout}
      onContinuar={() => void seccion.guardarYContinuar()}
      instrucciones={
        <p>
          Indique si tiene familiares hasta el segundo grado de consanguinidad o cuarto de afinidad que trabajen en el
          IESS o en el BIESS.
        </p>
      }
    >
      <PreguntaSiNo
        form={form}
        name="tieneFamiliares"
        legend="¿Tiene familiares hasta el segundo grado de consanguinidad o cuarto de afinidad trabajando en el IESS o el BIESS?"
        onValueChange={(value) => {
          const cantidad = form.getValues('familiares').length;
          if (value === 'NO' && cantidad > 0) {
            // RF-EF12-08: answering "No" removes the family members already entered.
            form.setValue('familiares', [], { shouldDirty: true });
            setAviso(`Se quitaron ${cantidad} familiares de la lista porque respondió “No”. Se eliminarán al guardar.`);
          } else {
            setAviso('');
          }
        }}
      />
      <p aria-live="polite" className="text-base text-neutral-700">
        {aviso}
      </p>

      {tieneFamiliares === 'SI' ? (
        <section aria-labelledby="familiares-lista" className="space-y-4">
          <h2 id="familiares-lista" className="text-lg font-semibold text-neutral-900">
            Familiares registrados
          </h2>
          <ListaRepetible<FamiliarValues>
            form={form}
            name="familiares"
            schema={familiarSchema}
            singular="familiar"
            preguntaEliminar="¿Eliminar este familiar?"
            minimo="Debe registrar al menos un familiar."
            vacio="Aún no ha registrado familiares. Use “Agregar familiar” para incluir el primero."
            nuevo={() => ({ id: nuevoId(), nombresApellidos: '', parentesco: '', institucion: '', area: '', cargo: '' })}
            resumen={(item) => ({
              titulo: item.nombresApellidos,
              detalles: [
                `Parentesco: ${labelOf(PARENTESCO, item.parentesco)}`,
                `${labelOf(INSTITUCION_FAMILIAR, item.institucion)}, ${item.area}, ${item.cargo}`,
              ],
            })}
            renderCampos={(itemForm) => (
              <>
                <Texto form={itemForm} name="nombresApellidos" label="Nombres y apellidos" required autoComplete="off" />
                <Seleccion form={itemForm} name="parentesco" label="Parentesco" required options={PARENTESCO} />
                <Seleccion form={itemForm} name="institucion" label="Institución" required options={INSTITUCION_FAMILIAR} />
                <Texto form={itemForm} name="area" label="Área" required autoComplete="off" />
                <Texto form={itemForm} name="cargo" label="Cargo" required autoComplete="off" />
              </>
            )}
          />
        </section>
      ) : null}
    </SeccionLayout>
  );
}
