package domain

import (
	"context"
	"time"
)

type EstadoNotificacion string

const (
	NotificacionPendiente EstadoNotificacion = "PENDIENTE"
	NotificacionEnviada   EstadoNotificacion = "ENVIADA"
	NotificacionFallida   EstadoNotificacion = "FALLIDA"
)

// Notificacion representa un mensaje registrado en el patrón Outbox para envío asíncrono (ADR-004, SDD 10.9).
type Notificacion struct {
	ID            string             `json:"id"`
	PostulacionID string             `json:"postulacion_id"`
	PostulanteID  string             `json:"postulante_id"`
	Tipo          string             `json:"tipo"`
	Destinatario  string             `json:"destinatario"`
	Asunto        string             `json:"asunto"`
	Plantilla     string             `json:"plantilla"`
	Payload       map[string]any     `json:"payload"`
	Estado        EstadoNotificacion `json:"estado"`
	CreatedAt     time.Time          `json:"created_at"`
}

// NotificacionRepository define el puerto para encolar notificaciones transaccionales.
type NotificacionRepository interface {
	Enqueue(ctx context.Context, n *Notificacion) error
}
