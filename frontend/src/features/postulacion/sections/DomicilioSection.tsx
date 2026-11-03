import { useMemo, useRef } from 'react';
import TextField from '@/components/ui/TextField';
import { Texto } from '../components/campos';
import SeccionLayout from '../components/SeccionLayout';
import UbicacionCascada from '../components/UbicacionCascada';
import { useSeccionForm } from '../hooks/useSeccionForm';
import { usePostulacion } from '../PostulacionContext';
import { domicilioSchema } from '../schemas';
import type { DomicilioValues } from '../schemas';

const INICIALES: DomicilioValues = {
  paisId: '',
  provinciaId: '',
  cantonId: '',
  provinciaManual: '',
  ciudadManual: '',
  telefonoCelular: '',
  correoAlterno: '',
  callePrincipal: '',
  numero: '',
  calleSecundaria: '',
  sector: '',
  referencia: '',
  telefonoDomicilio: '',
};

export default function DomicilioSection() {
  const { verificados } = usePostulacion();
  // Updated by the cascade once the cantons of the selected province are known.
  const hayCantones = useRef(true);
  const ecuadorId = useRef<string | null>(null);
  const schema = useMemo(
    () =>
      domicilioSchema(
        verificados.correo,
        () => hayCantones.current,
        (paisId) => paisId !== '' && paisId === ecuadorId.current,
      ),
    [verificados.correo],
  );
  const seccion = useSeccionForm('domicilio', schema, INICIALES);
  const { form } = seccion;

  return (
    <SeccionLayout
      seccion="domicilio"
      {...seccion.layout}
      onContinuar={() => void seccion.guardarYContinuar()}
      instrucciones={<p>Ingrese el lugar donde reside actualmente y los datos para contactarle.</p>}
    >
      <section aria-labelledby="ubicacion-titulo" className="space-y-6">
        <h2 id="ubicacion-titulo" className="text-lg font-semibold text-neutral-900">
          Ubicación
        </h2>
        <UbicacionCascada
          form={form}
          onCantonesCargados={(cantidad) => (hayCantones.current = cantidad > 0)}
          onEcuadorIdentificado={(id) => (ecuadorId.current = id)}
        />
      </section>

      <section aria-labelledby="direccion-titulo" className="space-y-6">
        <h2 id="direccion-titulo" className="text-lg font-semibold text-neutral-900">
          Dirección
        </h2>
        <Texto form={form} name="callePrincipal" label="Calle principal" required autoComplete="address-line1" />
        <div className="grid gap-6 sm:grid-cols-2">
          <Texto form={form} name="numero" label="Número" required autoComplete="off" helperText="Por ejemplo N34-120." />
          <Texto form={form} name="calleSecundaria" label="Calle secundaria" autoComplete="address-line2" />
        </div>
        <Texto form={form} name="sector" label="Sector o barrio" autoComplete="off" />
        <Texto form={form} name="referencia" label="Referencia" autoComplete="off" helperText="Un lugar conocido cerca de su domicilio." />
      </section>

      <section aria-labelledby="contacto-titulo" className="space-y-6">
        <h2 id="contacto-titulo" className="text-lg font-semibold text-neutral-900">
          Contacto
        </h2>
        <TextField
          label="Correo principal"
          value={verificados.correo}
          readOnly
          helperText="Es el correo registrado en su cuenta. No se puede modificar en esta sección."
        />
        <Texto
          form={form}
          name="correoAlterno"
          label="Correo secundario"
          type="email"
          inputMode="email"
          autoComplete="email"
          helperText="Debe ser distinto del correo principal."
        />
        <div className="grid gap-6 sm:grid-cols-2">
          <Texto form={form} name="telefonoCelular" label="Teléfono celular" required type="tel" autoComplete="tel" className="tabular-nums" />
          <Texto form={form} name="telefonoDomicilio" label="Teléfono de domicilio" type="tel" autoComplete="tel" />
        </div>
      </section>
    </SeccionLayout>
  );
}
