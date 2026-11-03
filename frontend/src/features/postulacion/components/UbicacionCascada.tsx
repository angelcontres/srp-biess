import { useCallback, useEffect, useMemo, useState } from 'react';
import { RotateCw } from 'lucide-react';
import type { UseFormReturn } from 'react-hook-form';
import Button from '@/components/ui/Button';
import { ECUADOR_ISO } from '../constants';
import { usePostulacion } from '../PostulacionContext';
import type { DomicilioValues } from '../schemas';
import type { OpcionPais } from '../types';
import { Seleccion, Texto } from './campos';

type Estado = { opciones: OpcionPais[]; cargando: boolean; error: boolean };
const VACIO: Estado = { opciones: [], cargando: false, error: false };

function useCatalogo(cargar: (() => Promise<OpcionPais[]>) | null, alCargar: (opciones: OpcionPais[]) => void) {
  const [estado, setEstado] = useState<Estado>(VACIO);
  const ejecutar = useCallback(async () => {
    if (!cargar) {
      setEstado(VACIO);
      return;
    }
    setEstado({ opciones: [], cargando: true, error: false });
    try {
      const opciones = await cargar();
      setEstado({ opciones, cargando: false, error: false });
      alCargar(opciones);
    } catch {
      setEstado({ opciones: [], cargando: false, error: true });
    }
    // `alCargar` only re-syncs the select; the dependency is the loader.
  }, [cargar]);
  useEffect(() => {
    void ejecutar();
  }, [ejecutar]);
  return { ...estado, reintentar: ejecutar };
}

function CargandoCampo({ label }: { label: string }) {
  return (
    <div aria-busy="true">
      <p className="text-base font-medium text-neutral-900">{label}</p>
      <div className="mt-2 h-11 w-full rounded-md bg-line motion-safe:animate-pulse" aria-hidden="true" />
      <span className="sr-only" role="status">
        Cargando {label.toLowerCase()}
      </span>
    </div>
  );
}

function ErrorCatalogo({ onReintentar }: { onReintentar: () => void }) {
  return (
    <div role="alert" className="flex flex-col gap-2 rounded-md border border-danger border-l-4 bg-danger-50 p-3">
      <p className="text-base text-neutral-900">No se pudieron cargar los catálogos. Reintente.</p>
      <Button variant="secondary" onClick={onReintentar} leftIcon={<RotateCw size={20} strokeWidth={2} aria-hidden="true" />}>
        Reintentar
      </Button>
    </div>
  );
}

/**
 * País → provincia → cantón (RF-EF12-07, SDD 8.6.5). Changing a parent clears and
 * announces its children; outside Ecuador the province and city are typed.
 * Catalog caching and the Redis fallback are owned by SRP-014 / SRP-007.
 */
export interface UbicacionCascadaProps {
  form: UseFormReturn<DomicilioValues>;
  /** Reports how many cantons the selected province has (0 makes the canton optional). */
  onCantonesCargados?: (cantidad: number) => void;
  /** Reports the catalog id of Ecuador (ISO "EC") once the countries are loaded. */
  onEcuadorIdentificado?: (paisId: string | null) => void;
}

export default function UbicacionCascada({ form, onCantonesCargados, onEcuadorIdentificado }: UbicacionCascadaProps) {
  const { gateway } = usePostulacion();
  const [anuncio, setAnuncio] = useState('');
  const paisId = form.watch('paisId');
  const provinciaId = form.watch('provinciaId');

  const resync = (campo: 'paisId' | 'provinciaId' | 'cantonId') => () =>
    window.setTimeout(() => form.setValue(campo, form.getValues(campo)), 0);

  // Children are shown only after the countries load, so the ISO code is always known.
  const [ecuadorId, setEcuadorId] = useState<string | null>(null);
  const esEcuador = ecuadorId !== null && paisId === ecuadorId;

  const cargarPaises = useMemo(() => () => gateway.catalogo.paises(), [gateway]);
  const cargarProvincias = useMemo(
    () => (esEcuador ? () => gateway.catalogo.provincias(paisId) : null),
    [esEcuador, gateway, paisId],
  );
  const cargarCantones = useMemo(
    () => (esEcuador && provinciaId ? () => gateway.catalogo.cantones(provinciaId) : null),
    [esEcuador, gateway, provinciaId],
  );
  const paises = useCatalogo(cargarPaises, (opciones) => {
    const id = opciones.find((pais) => pais.codigo === ECUADOR_ISO)?.value ?? null;
    setEcuadorId(id);
    onEcuadorIdentificado?.(id);
    resync('paisId')();
  });
  const provincias = useCatalogo(cargarProvincias, resync('provinciaId'));
  const cantones = useCatalogo(cargarCantones, (opciones) => {
    onCantonesCargados?.(opciones.length);
    resync('cantonId')();
  });

  const limpiar = (campos: Array<keyof DomicilioValues>) =>
    campos.forEach((campo) => form.setValue(campo, '', { shouldDirty: true }));

  return (
    <div className="space-y-6">
      {paises.error ? (
        <ErrorCatalogo onReintentar={() => void paises.reintentar()} />
      ) : paises.cargando ? (
        <CargandoCampo label="País de residencia" />
      ) : (
        <Seleccion
          form={form}
          name="paisId"
          label="País de residencia"
          required
          autoComplete="country"
          options={paises.opciones}
          onValueChange={() => {
            limpiar(['provinciaId', 'cantonId', 'provinciaManual', 'ciudadManual']);
            setAnuncio('Se reiniciaron la provincia y el cantón o ciudad.');
          }}
        />
      )}

      {paisId && ecuadorId !== null && esEcuador ? (
        <div className="grid gap-6 sm:grid-cols-2">
          {provincias.error ? (
            <ErrorCatalogo onReintentar={() => void provincias.reintentar()} />
          ) : provincias.cargando ? (
            <CargandoCampo label="Provincia" />
          ) : (
            <Seleccion
              form={form}
              name="provinciaId"
              label="Provincia"
              required
              autoComplete="address-level1"
              options={provincias.opciones}
              onValueChange={() => {
                limpiar(['cantonId']);
                setAnuncio('Se reinició el cantón.');
              }}
            />
          )}
          {cantones.error ? (
            <ErrorCatalogo onReintentar={() => void cantones.reintentar()} />
          ) : cantones.cargando ? (
            <CargandoCampo label="Cantón" />
          ) : (
            <Seleccion
              form={form}
              name="cantonId"
              label="Cantón"
              required={cantones.opciones.length > 0 || !provinciaId}
              autoComplete="address-level2"
              options={cantones.opciones}
              disabled={!provinciaId || cantones.opciones.length === 0}
              helperText={
                !provinciaId
                  ? 'Seleccione primero la provincia.'
                  : cantones.opciones.length === 0
                    ? 'No hay cantones registrados para esta provincia. Puede continuar sin seleccionarlo.'
                    : undefined
              }
            />
          )}
        </div>
      ) : null}

      {paisId && !paises.cargando && !paises.error && !esEcuador ? (
        <div className="grid gap-6 sm:grid-cols-2">
          <Texto form={form} name="provinciaManual" label="Provincia o estado" required autoComplete="address-level1" />
          <Texto form={form} name="ciudadManual" label="Ciudad" required autoComplete="address-level2" />
        </div>
      ) : null}

      <p aria-live="polite" className="sr-only">
        {anuncio}
      </p>
    </div>
  );
}
