import { useMemo, useState } from 'react';
import { Info } from 'lucide-react';
import TextField from '@/components/ui/TextField';
import { formatIsoDate } from '@/utils/format';
import { Seleccion, Texto } from '../components/campos';
import SeccionLayout from '../components/SeccionLayout';
import { ESTADO_CIVIL, ESTADOS_CON_PAREJA, GENERO, labelOf } from '../constants';
import { useSeccionForm } from '../hooks/useSeccionForm';
import { usePostulacion } from '../PostulacionContext';
import { personalSchema } from '../schemas';
import type { PersonalValues } from '../schemas';

const INICIALES: PersonalValues = { estadoCivil: '', parejaNombres: '', parejaCedula: '' };

export default function PersonalSection() {
  const { verificados } = usePostulacion();
  const schema = useMemo(() => personalSchema(verificados.cedula), [verificados.cedula]);
  const seccion = useSeccionForm('personal', schema, INICIALES);
  const { form } = seccion;
  const [anuncio, setAnuncio] = useState('');
  const estadoCivil = form.watch('estadoCivil');
  const conPareja = ESTADOS_CON_PAREJA.includes(estadoCivil);
  // RF-EF12-04: a non-Ecuadorian nationality blocks saving and moving forward.
  const ecuatoriana = /^ecuatorian/i.test(verificados.nacionalidad.trim());

  return (
    <SeccionLayout
      seccion="personal"
      {...seccion.layout}
      soloLectura={!ecuatoriana}
      onContinuar={() => void seccion.guardarYContinuar()}
      instrucciones={
        <p>
          Sus datos de identidad se obtuvieron del Registro Civil y no se pueden modificar. Complete su estado civil.
          Puede guardar su avance y continuar después.
        </p>
      }
    >
      {!ecuatoriana ? (
        <div className="flex gap-3 rounded-md border border-primary border-l-4 bg-info-50 p-4">
          <Info size={20} strokeWidth={2} aria-hidden="true" className="shrink-0 text-primary" />
          <div className="space-y-1">
            <p className="font-semibold">Resultado de la validación de nacionalidad</p>
            <p>Su postulación no puede continuar porque la nacionalidad registrada no es ecuatoriana.</p>
            <p className="text-sm text-muted">Este concurso requiere nacionalidad ecuatoriana, según las bases del concurso.</p>
          </div>
        </div>
      ) : null}

      <section aria-labelledby="datos-verificados" className="space-y-6">
        <h2 id="datos-verificados" className="text-lg font-semibold text-neutral-900">
          Datos verificados
        </h2>
        <div className="grid gap-6 sm:grid-cols-2">
          <TextField label="Cédula de identidad" value={verificados.cedula} verified readOnly className="tabular-nums" />
          <TextField label="Nacionalidad" value={verificados.nacionalidad} verified readOnly />
          <TextField label="Nombres" value={verificados.nombres} verified readOnly />
          <TextField label="Apellidos" value={verificados.apellidos} verified readOnly />
          <TextField label="Género" value={labelOf(GENERO, verificados.genero) || verificados.genero} verified readOnly />
          <TextField
            label="Fecha de nacimiento"
            value={verificados.fechaNacimiento ? formatIsoDate(verificados.fechaNacimiento) : ''}
            verified
            readOnly
            className="tabular-nums"
          />
        </div>
        <p className="text-sm text-muted">
          Si algún dato verificado es incorrecto: [PENDIENTE: canal de soporte para corregir datos del Registro Civil].
        </p>
      </section>

      {ecuatoriana ? (
        <section aria-labelledby="estado-civil-titulo" className="space-y-6">
          <h2 id="estado-civil-titulo" className="text-lg font-semibold text-neutral-900">
            Estado civil
          </h2>
          <Seleccion
            form={form}
            name="estadoCivil"
            label="Estado civil"
            required
            options={ESTADO_CIVIL}
            onValueChange={(value) => {
              const quitaPareja = !ESTADOS_CON_PAREJA.includes(value);
              const { parejaNombres, parejaCedula } = form.getValues();
              if (quitaPareja && (parejaNombres || parejaCedula)) {
                // RF-EF12-05: hidden partner data is cleared and not sent.
                form.setValue('parejaNombres', '', { shouldDirty: true });
                form.setValue('parejaCedula', '', { shouldDirty: true });
                setAnuncio('Se ocultaron y eliminaron los datos de pareja, porque el estado civil seleccionado no los requiere.');
              } else {
                setAnuncio(quitaPareja ? '' : 'Se mostraron los campos de datos de la pareja.');
              }
            }}
          />
          <p aria-live="polite" className="text-sm text-muted">
            {anuncio}
          </p>
          {conPareja ? (
            <div className="space-y-6 border-l-4 border-line pl-4">
              <Texto form={form} name="parejaNombres" label="Nombres y apellidos de la pareja" required autoComplete="off" />
              <Texto
                form={form}
                name="parejaCedula"
                label="Cédula de la pareja"
                required
                numeric
                maxLength={10}
                autoComplete="off"
                helperText="Ingrese los 10 números de la cédula."
              />
            </div>
          ) : null}
        </section>
      ) : null}
    </SeccionLayout>
  );
}
