package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
)

type txDirecto struct{}

func (txDirecto) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type outboxFake struct {
	cola     []*domain.NotificacionEnCola
	enviadas []string
	fallos   []struct {
		intentos   int
		proximo    time.Time
		definitivo bool
	}
}

func (o *outboxFake) TomarPendiente(context.Context) (*domain.NotificacionEnCola, error) {
	if len(o.cola) == 0 {
		return nil, domain.ErrNotFound
	}
	n := o.cola[0]
	o.cola = o.cola[1:]
	return n, nil
}
func (o *outboxFake) MarcarEnviada(_ context.Context, id string) error {
	o.enviadas = append(o.enviadas, id)
	return nil
}
func (o *outboxFake) ReprogramarFallo(_ context.Context, _ string, intentos int, proximo time.Time, _ string, definitivo bool) error {
	o.fallos = append(o.fallos, struct {
		intentos   int
		proximo    time.Time
		definitivo bool
	}{intentos, proximo, definitivo})
	return nil
}

type senderFake struct {
	err      error
	enviados []domain.Correo
}

func (s *senderFake) Enviar(_ context.Context, c domain.Correo) error {
	if s.err != nil {
		return s.err
	}
	s.enviados = append(s.enviados, c)
	return nil
}

type plantillasFake struct{}

func (plantillasFake) Renderizar(p string, datos map[string]any) (string, string, error) {
	if p != "confirmacion_postulacion.html" {
		return "", "", domain.ErrPlantillaDesconocida
	}
	return "<p>" + datos["fecha_envio_texto"].(string) + "</p>", "texto", nil
}

func notificacion(intentos int, plantilla string) *domain.NotificacionEnCola {
	return &domain.NotificacionEnCola{Intentos: intentos, Notificacion: domain.Notificacion{
		ID: "n1", Destinatario: "maria@correo.ec", Asunto: "Confirmación", Plantilla: plantilla,
		Payload: map[string]any{"codigo_unico": "BIESS-1", "fecha_envio": "2026-11-03T21:41:00Z"},
	}}
}

func TestNotificacionService(t *testing.T) {
	ahora := time.Date(2026, 11, 3, 22, 0, 0, 0, time.UTC)
	nuevo := func(o *outboxFake, s *senderFake) *application.NotificacionService {
		return application.NewNotificacionService(txDirecto{}, o, s, plantillasFake{}, nil, nil, nil).
			WithReloj(func() time.Time { return ahora })
	}

	t.Run("envía y marca ENVIADA, con la fecha en hora de Ecuador", func(t *testing.T) {
		o, s := &outboxFake{cola: []*domain.NotificacionEnCola{notificacion(0, "confirmacion_postulacion.html")}}, &senderFake{}
		n, err := nuevo(o, s).ProcesarPendientes(context.Background(), 10)
		if err != nil || n != 1 || len(o.enviadas) != 1 || len(s.enviados) != 1 {
			t.Fatalf("n=%d err=%v enviadas=%v", n, err, o.enviadas)
		}
		if s.enviados[0].HTML != "<p>3 de noviembre de 2026, 16:41 (hora de Ecuador)</p>" || s.enviados[0].Para != "maria@correo.ec" {
			t.Fatalf("correo inesperado: %+v", s.enviados[0])
		}
	})

	t.Run("un fallo reprograma con backoff exponencial", func(t *testing.T) {
		o := &outboxFake{cola: []*domain.NotificacionEnCola{notificacion(2, "confirmacion_postulacion.html")}}
		_, err := nuevo(o, &senderFake{err: errors.New("SMTP caído")}).ProcesarPendientes(context.Background(), 10)
		if err != nil || len(o.fallos) != 1 || len(o.enviadas) != 0 {
			t.Fatalf("err=%v fallos=%+v", err, o.fallos)
		}
		// Tercer intento: 1 min << 2 = 4 min.
		if f := o.fallos[0]; f.intentos != 3 || f.definitivo || !f.proximo.Equal(ahora.Add(4*time.Minute)) {
			t.Fatalf("reprogramación inesperada: %+v", f)
		}
	})

	t.Run("el quinto intento fallido la marca FALLIDA", func(t *testing.T) {
		o := &outboxFake{cola: []*domain.NotificacionEnCola{notificacion(4, "confirmacion_postulacion.html")}}
		_, _ = nuevo(o, &senderFake{err: errors.New("SMTP caído")}).ProcesarPendientes(context.Background(), 10)
		if len(o.fallos) != 1 || o.fallos[0].intentos != application.MaxIntentosNotificacion || !o.fallos[0].definitivo {
			t.Fatalf("esperado FALLIDA al quinto intento: %+v", o.fallos)
		}
	})

	t.Run("una plantilla desconocida falla de inmediato, sin reintentos", func(t *testing.T) {
		o, s := &outboxFake{cola: []*domain.NotificacionEnCola{notificacion(0, "otra.html")}}, &senderFake{}
		_, _ = nuevo(o, s).ProcesarPendientes(context.Background(), 10)
		if len(o.fallos) != 1 || !o.fallos[0].definitivo || len(s.enviados) != 0 {
			t.Fatalf("esperado fallo definitivo: %+v", o.fallos)
		}
	})
}
