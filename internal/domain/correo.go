package domain

import (
	"context"
	"errors"
	"time"
)

// Envío de correos del outbox de notificaciones (ADR-011, SDD 10.9, SRS RF-EF12-23).

// Adjunto es un archivo del correo (el formulario PDF de la postulación).
type Adjunto struct {
	Nombre    string
	Tipo      string // Content-Type, p. ej. application/pdf
	Contenido []byte
}

// Correo es un mensaje listo para enviar: versión HTML y versión de texto plano.
type Correo struct {
	Para     string   // Destinatario principal
	CC       []string // Destinatarios en copia (correos secundarios)
	Asunto   string
	HTML     string
	Texto    string
	Adjuntos []Adjunto
}

// EmailSender es el puerto hacia el servidor SMTP (SDD 5: EmailSender).
type EmailSender interface {
	Enviar(ctx context.Context, c Correo) error
}

// RenderizadorCorreo compone el cuerpo de un correo a partir de su plantilla y sus datos.
type RenderizadorCorreo interface {
	Renderizar(plantilla string, datos map[string]any) (html, texto string, err error)
}

// OutboxRepository toma y cierra notificaciones pendientes. Tomar debe ejecutarse dentro de una
// transacción: bloquea la fila (FOR UPDATE SKIP LOCKED) para que dos workers no envíen el mismo correo.
type OutboxRepository interface {
	TomarPendiente(ctx context.Context) (*NotificacionEnCola, error) // ErrNotFound si no hay
	MarcarEnviada(ctx context.Context, id string) error
	// ReprogramarFallo suma el intento y programa el siguiente; si `definitivo`, la marca FALLIDA.
	ReprogramarFallo(ctx context.Context, id string, intentos int, proximo time.Time, motivo string, definitivo bool) error
}

// NotificacionEnCola es una notificación tomada del outbox, con sus intentos previos.
type NotificacionEnCola struct {
	Notificacion
	Intentos int
}

// ErrPlantillaDesconocida: la notificación pide una plantilla que no existe (fallo definitivo).
var ErrPlantillaDesconocida = errors.New("plantilla de correo desconocida")
