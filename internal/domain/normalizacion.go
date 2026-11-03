package domain

import (
	"context"
	"time"
)

// NormalizadorPostulacion vuelca las secciones guardadas como JSONB en `postulacion_seccion`
// a las tablas de SDD 4.6 (SRP-017e). Se invoca dentro de la transacción de envío (SRP-017c)
// para que el volcado y el cambio de estado sean atómicos; no expone endpoints.
type NormalizadorPostulacion interface {
	Normalizar(ctx context.Context, postulacionID string) error
}

// TiempoExperiencia es la duración de una experiencia en años, meses y días (SRS 7.5).
type TiempoExperiencia struct {
	Anios int
	Meses int
	Dias  int
}

// CalcularTiempoExperiencia devuelve la diferencia de calendario entre inicio y fin anclada
// en el día de inicio, con el mismo criterio que la SPA (utils/experiencia.ts):
// 15/01/2020 -> 10/03/2023 = 3 años, 1 mes, 23 días; 31/01/2023 -> 28/02/2023 = 1 mes.
// ok es false si fin es anterior a inicio.
func CalcularTiempoExperiencia(inicio, fin time.Time) (t TiempoExperiencia, ok bool) {
	inicio = fechaCalendario(inicio)
	fin = fechaCalendario(fin)
	if fin.Before(inicio) {
		return TiempoExperiencia{}, false
	}
	meses := (fin.Year()-inicio.Year())*12 + int(fin.Month()-inicio.Month())
	for meses > 0 && sumarMesesAjustado(inicio, meses).After(fin) {
		meses--
	}
	ancla := sumarMesesAjustado(inicio, meses)
	dias := int(fin.Sub(ancla).Hours() / 24)
	return TiempoExperiencia{Anios: meses / 12, Meses: meses % 12, Dias: dias}, true
}

func fechaCalendario(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// sumarMesesAjustado suma meses conservando el día, limitado al último día del mes destino.
func sumarMesesAjustado(inicio time.Time, meses int) time.Time {
	primero := time.Date(inicio.Year(), inicio.Month()+time.Month(meses), 1, 0, 0, 0, 0, time.UTC)
	ultimoDia := primero.AddDate(0, 1, -1).Day()
	dia := inicio.Day()
	if dia > ultimoDia {
		dia = ultimoDia
	}
	return time.Date(primero.Year(), primero.Month(), dia, 0, 0, 0, 0, time.UTC)
}

// Efectos de una respuesta de inhabilidad (enum efecto_inhabilidad, SDD 4.3).
const (
	EfectoBloqueante  = "BLOQUEANTE"
	EfectoAlerta      = "ALERTA"
	EfectoInformativo = "INFORMATIVO"
)

// EfectoInhabilidad calcula el efecto de una respuesta (RN-26): una pregunta BLOQUEANTE bloquea
// solo con su respuesta bloqueante (1.9.1 con "No"); una pregunta ALERTA genera alerta con "Sí".
// Cualquier otra combinación queda como INFORMATIVO.
func EfectoInhabilidad(efectoPregunta string, respuestaBloqueante *bool, respuesta bool) string {
	switch efectoPregunta {
	case EfectoBloqueante:
		if respuestaBloqueante != nil && respuesta == *respuestaBloqueante {
			return EfectoBloqueante
		}
	case EfectoAlerta:
		if respuesta {
			return EfectoAlerta
		}
	}
	return EfectoInformativo
}
