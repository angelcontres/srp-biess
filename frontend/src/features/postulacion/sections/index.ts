import type { ComponentType } from 'react';
import type { SeccionId } from '../constants';
import type { EnvioResultado } from '../types';
import CapacitacionSection from './CapacitacionSection';
import CargoSection from './CargoSection';
import DomicilioSection from './DomicilioSection';
import ExperienciaSection from './ExperienciaSection';
import FamiliaresSection from './FamiliaresSection';
import FormacionSection from './FormacionSection';
import InhabilidadesSection from './InhabilidadesSection';
import PersonalSection from './PersonalSection';
import PublicacionesSection from './PublicacionesSection';
import ResumenSection from './ResumenSection';

export interface SeccionProps {
  onEnviada: (resultado: EnvioResultado) => void;
}

export const SECCION_COMPONENTES: Record<SeccionId, ComponentType<SeccionProps>> = {
  personal: PersonalSection,
  domicilio: DomicilioSection,
  familiares: FamiliaresSection,
  cargo: CargoSection,
  formacion: FormacionSection,
  experiencia: ExperienciaSection,
  publicaciones: PublicacionesSection,
  capacitacion: CapacitacionSection,
  inhabilidades: InhabilidadesSection,
  resumen: ResumenSection,
};
