import { useState } from 'react';
import { AreaTexto, Documento, Fecha, PreguntaSiNo, Seleccion, Texto } from '../components/campos';
import ListaRepetible from '../components/ListaRepetible';
import SeccionLayout from '../components/SeccionLayout';
import { labelOf, MAX_MB_DOCUMENTO, TIPO_PUBLICACION } from '../constants';
import { useSeccionForm } from '../hooks/useSeccionForm';
import { publicacionesSchema, publicacionItemSchema } from '../schemas';
import type { PublicacionesValues, PublicacionItem } from '../schemas';
import { nuevoId } from '../schemas/comunes';
import { fechaLegible } from '../utils/fechas';

const INICIALES: PublicacionesValues = { tienePublicaciones: '', publicaciones: [] };

export default function PublicacionesSection() {
  const seccion = useSeccionForm('publicaciones', publicacionesSchema, INICIALES);
  const { form } = seccion;
  const [aviso, setAviso] = useState('');
  const tiene = form.watch('tienePublicaciones');

  return (
    <SeccionLayout
      seccion="publicaciones"
      {...seccion.layout}
      onContinuar={() => void seccion.guardarYContinuar()}
      instrucciones={<p>Indique si cuenta con publicaciones o investigaciones académicas. Cada registro debe incluir su PDF.</p>}
    >
      <PreguntaSiNo
        form={form}
        name="tienePublicaciones"
        legend="¿Cuenta con publicaciones o investigaciones académicas?"
        onValueChange={(value) => {
          const cantidad = form.getValues('publicaciones').length;
          if (value === 'NO' && cantidad > 0) {
            // RF-EF12-14: answering "No" clears the records.
            form.setValue('publicaciones', [], { shouldDirty: true });
            setAviso(`Se quitaron ${cantidad} publicaciones de la lista porque respondió “No”. Se eliminarán al guardar.`);
          } else {
            setAviso('');
          }
        }}
      />
      <p aria-live="polite" className="text-base text-neutral-700">
        {aviso}
      </p>

      {tiene === 'SI' ? (
        <section aria-labelledby="publicaciones-lista" className="space-y-4">
          <h2 id="publicaciones-lista" className="text-lg font-semibold text-neutral-900">
            Publicaciones registradas
          </h2>
          <ListaRepetible<PublicacionItem>
            form={form}
            name="publicaciones"
            schema={publicacionItemSchema}
            singular="publicación"
            preguntaEliminar="¿Eliminar esta publicación?"
            minimo="Debe registrar al menos una publicación."
            vacio="Aún no ha registrado publicaciones. Use “Agregar publicación” para incluir la primera."
            nuevo={() => ({ id: nuevoId(), titulo: '', tipo: '', medio: '', fecha: '', relacion: '', documento: null })}
            resumen={(item) => ({
              titulo: item.titulo,
              detalles: [`${labelOf(TIPO_PUBLICACION, item.tipo)} en ${item.medio}`, `Publicada el ${fechaLegible(item.fecha)}`],
              documento: item.documento,
            })}
            renderCampos={(itemForm) => (
              <>
                <Texto form={itemForm} name="titulo" label="Título" required autoComplete="off" />
                <Seleccion form={itemForm} name="tipo" label="Tipo" required options={TIPO_PUBLICACION} />
                <Texto form={itemForm} name="medio" label="Medio de publicación" required autoComplete="off" />
                <Fecha form={itemForm} name="fecha" label="Fecha de publicación" required />
                <AreaTexto
                  form={itemForm}
                  name="relacion"
                  label="Relación con su formación o ejercicio profesional"
                  required
                  maxLength={500}
                />
                <Documento form={itemForm} name="documento" label="Publicación en PDF" tipo="PUBLICACION" maxSizeMB={MAX_MB_DOCUMENTO} />
              </>
            )}
          />
        </section>
      ) : null}
    </SeccionLayout>
  );
}
