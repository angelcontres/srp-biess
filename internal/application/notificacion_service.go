package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"srp-biess/internal/domain"
)

// Worker del outbox de notificaciones (ADR-011, SDD 10.9, SRS RF-EF12-23): toma una notificación
// pendiente, compone el correo con el formulario PDF adjunto, lo envía y la marca ENVIADA. Si falla,
// reintenta con backoff exponencial hasta 5 veces y luego la marca FALLIDA con una alerta en el log.

const (
	// MaxIntentosNotificacion: RF-EF12-23 ("hasta 5 veces").
	MaxIntentosNotificacion = 5
	// BackoffBaseNotificacion: 1, 2, 4 y 8 minutos entre intentos.
	BackoffBaseNotificacion = time.Minute
)

type NotificacionService struct {
	tx         domain.TxManager
	outbox     domain.OutboxRepository
	sender     domain.EmailSender
	plantillas domain.RenderizadorCorreo
	documentos domain.DocumentoRepository
	almacen    domain.AlmacenDocumentos
	catalogo   domain.CatalogoRepository
	ahora      func() time.Time
}

func NewNotificacionService(
	tx domain.TxManager, outbox domain.OutboxRepository, sender domain.EmailSender, plantillas domain.RenderizadorCorreo,
	documentos domain.DocumentoRepository, almacen domain.AlmacenDocumentos, catalogo domain.CatalogoRepository,
) *NotificacionService {
	return &NotificacionService{
		tx: tx, outbox: outbox, sender: sender, plantillas: plantillas,
		documentos: documentos, almacen: almacen, catalogo: catalogo, ahora: time.Now,
	}
}

// WithReloj fija la hora en pruebas.
func (s *NotificacionService) WithReloj(ahora func() time.Time) *NotificacionService {
	s.ahora = ahora
	return s
}

// ProcesarPendientes envía hasta `limite` notificaciones y devuelve cuántas procesó (enviadas o
// reprogramadas). Se detiene cuando no quedan pendientes listas.
func (s *NotificacionService) ProcesarPendientes(ctx context.Context, limite int) (int, error) {
	procesadas := 0
	for procesadas < limite {
		hubo, err := s.procesarSiguiente(ctx)
		if err != nil {
			return procesadas, err
		}
		if !hubo {
			break
		}
		procesadas++
	}
	return procesadas, nil
}

// procesarSiguiente trabaja una notificación dentro de una transacción: el bloqueo de la fila dura
// hasta marcarla, así otro worker no envía el mismo correo (los reintentos no lo duplican).
func (s *NotificacionService) procesarSiguiente(ctx context.Context) (bool, error) {
	hubo := false
	err := s.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		n, err := s.outbox.TomarPendiente(txCtx)
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		hubo = true

		correo, err := s.componer(txCtx, n)
		if err == nil {
			err = s.sender.Enviar(txCtx, correo)
		}
		if err == nil {
			return s.outbox.MarcarEnviada(txCtx, n.ID)
		}

		intentos := n.Intentos + 1
		definitivo := intentos >= MaxIntentosNotificacion || errors.Is(err, domain.ErrPlantillaDesconocida)
		proximo := s.ahora().Add(BackoffBaseNotificacion << (intentos - 1))
		if definitivo {
			log.Printf("[ALERTA] notificación %s (%s) FALLIDA tras %d intentos: %v", n.ID, n.Tipo, intentos, err)
		} else {
			log.Printf("[WARN] notificación %s: intento %d fallido, reintento a las %s: %v", n.ID, intentos, proximo.Format(time.RFC3339), err)
		}
		return s.outbox.ReprogramarFallo(txCtx, n.ID, intentos, proximo, err.Error(), definitivo)
	})
	return hubo, err
}

// componer arma el correo de la plantilla; la confirmación de postulación adjunta el PDF guardado.
func (s *NotificacionService) componer(ctx context.Context, n *domain.NotificacionEnCola) (domain.Correo, error) {
	datos := map[string]any{}
	for k, v := range n.Payload {
		datos[k] = v
	}
	if fecha, ok := n.Payload["fecha_envio"].(string); ok {
		if t, err := time.Parse(time.RFC3339, fecha); err == nil {
			datos["fecha_envio_texto"] = FechaHoraEcuador(t)
		}
	}
	if id := fmt.Sprint(n.Payload["cargo_id"]); s.catalogo != nil && id != "" && id != "<nil>" {
		if cargo, err := s.catalogo.GetCargoByID(ctx, id); err == nil && cargo != nil {
			datos["cargo"] = cargo.Nombre
		}
	}

	html, texto, err := s.plantillas.Renderizar(n.Plantilla, datos)
	if err != nil {
		return domain.Correo{}, err
	}
	correo := domain.Correo{Para: n.Destinatario, Asunto: n.Asunto, HTML: html, Texto: texto}

	if docID, ok := n.Payload["documento_id"].(string); ok && docID != "" {
		adjunto, err := s.adjunto(ctx, docID)
		if err != nil {
			return domain.Correo{}, fmt.Errorf("no se pudo adjuntar el formulario: %w", err)
		}
		correo.Adjuntos = append(correo.Adjuntos, adjunto)
	}
	return correo, nil
}

func (s *NotificacionService) adjunto(ctx context.Context, docID string) (domain.Adjunto, error) {
	if s.documentos == nil || s.almacen == nil {
		return domain.Adjunto{}, ErrPersistenciaNoDisponible
	}
	doc, err := s.documentos.GetByID(ctx, docID)
	if err != nil {
		return domain.Adjunto{}, err
	}
	archivo, err := s.almacen.Abrir(doc.NombreAlmacenado)
	if err != nil {
		return domain.Adjunto{}, err
	}
	defer archivo.Close()
	contenido, err := io.ReadAll(archivo)
	if err != nil {
		return domain.Adjunto{}, err
	}
	return domain.Adjunto{Nombre: doc.NombreOriginal, Tipo: "application/pdf", Contenido: contenido}, nil
}

// FechaHoraEcuador: "3 de noviembre de 2026, 16:41 (hora de Ecuador)" (SDD 8.5.9).
func FechaHoraEcuador(t time.Time) string {
	zona, err := time.LoadLocation("America/Guayaquil")
	if err != nil {
		zona = time.FixedZone("ECT", -5*3600)
	}
	meses := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto",
		"septiembre", "octubre", "noviembre", "diciembre"}
	l := t.In(zona)
	return fmt.Sprintf("%d de %s de %d, %02d:%02d (hora de Ecuador)", l.Day(), meses[l.Month()-1], l.Year(), l.Hour(), l.Minute())
}
