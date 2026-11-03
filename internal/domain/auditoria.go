package domain

import "context"

// Auditoría append-only en `auditoria_evento` (SDD 4.5.5, SDD 9.9).

// Tipos de actor (SDD 4.5.5).
const (
	ActorPostulante = "POSTULANTE"
	ActorSistema    = "SISTEMA"
)

// Acciones auditadas.
const (
	AccionCargoConfirmado       = "CARGO_CONFIRMADO"
	AccionPostulacionNoElegible = "POSTULACION_NO_ELEGIBLE"
	AccionAccesoDenegado        = "ACCESO_DENEGADO"
	AccionDocumentoCargado      = "DOCUMENTO_CARGADO"
	AccionDocumentoEliminado    = "DOCUMENTO_ELIMINADO"
)

// ContextoSolicitud es el origen de la acción: IP, navegador y trace_id (SDD 9.9).
type ContextoSolicitud struct {
	IP        string
	UserAgent string
	TraceID   string
}

// EventoAuditoria es un registro de auditoría. Antes y Despues deben ir redactados:
// nunca contraseñas, tokens ni datos personales que no aporten a la evidencia.
type EventoAuditoria struct {
	ActorID   string
	ActorTipo string
	Accion    string
	Entidad   string
	EntidadID string
	Antes     map[string]any
	Despues   map[string]any
	Solicitud ContextoSolicitud
}

// AuditoriaRepository persiste eventos; dentro de una transacción del TxManager se une a ella.
type AuditoriaRepository interface {
	Registrar(ctx context.Context, e EventoAuditoria) error
}
