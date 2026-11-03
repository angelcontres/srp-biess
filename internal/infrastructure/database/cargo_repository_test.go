package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
	"srp-biess/internal/infrastructure/external"
)

type auditoriaQueFalla struct{}

func (auditoriaQueFalla) Registrar(context.Context, domain.EventoAuditoria) error {
	return errors.New("auditoría no disponible")
}

// Prueba de integración de la confirmación del cargo (SRP-017a). Requiere PostgreSQL.
func TestCargoRepository_Integracion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	aplicarMigraciones(ctx, t, func(ctx context.Context, sql string) error {
		_, err := pool.Exec(ctx, sql)
		return err
	})

	repo := database.NewPostgresFormularioRepository(pool)
	nuevoBorrador := func(t *testing.T, etiqueta string) (*domain.Postulante, *domain.EstadoFormulario) {
		t.Helper()
		sufijo := time.Now().UnixNano()
		p := &domain.Postulante{
			Cedula:       fmt.Sprintf("08%08d", sufijo%100000000),
			Nombres:      "PRUEBA",
			Apellidos:    "CARGO",
			Email:        fmt.Sprintf("cargo_%s_%d@biess.gob.ec", etiqueta, sufijo),
			PasswordHash: "hash_seguro",
		}
		if err := database.NewPostgresPostulanteRepository(pool).Save(ctx, p); err != nil {
			t.Fatalf("no se pudo crear el postulante: %v", err)
		}
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM postulacion WHERE postulante_id = $1::uuid`, p.ID)
			_, _ = pool.Exec(context.Background(), `DELETE FROM postulante WHERE id = $1::uuid`, p.ID)
		})
		b, err := repo.CrearBorrador(ctx, p.ID)
		if errors.Is(err, domain.ErrFaseCerrada) {
			t.Skip("No hay convocatoria con la fase habilitada en la fecha actual")
		}
		if err != nil {
			t.Fatalf("CrearBorrador: %v", err)
		}
		return p, b
	}

	t.Run("elegible guarda cargo y elegibilidad", func(t *testing.T) {
		p, b := nuevoBorrador(t, "elegible")
		c, err := repo.ContextoConfirmacion(ctx, b.ID, p.ID, 1)
		if err != nil {
			t.Fatalf("ContextoConfirmacion: %v", err)
		}
		if c.Cedula != p.Cedula || c.Sector != domain.SectorAfiliados || c.ElegibilidadEstado != domain.ElegibilidadPendiente || c.Estado != "BORRADOR" {
			t.Fatalf("contexto inesperado: %+v", c)
		}
		if err := repo.ConfirmarCargo(ctx, b.ID, p.ID, domain.ResultadoElegibilidad{CargoID: 1, Elegible: true, Motivo: "Postulante afiliado al IESS"}, b.Version); err != nil {
			t.Fatalf("ConfirmarCargo: %v", err)
		}
		estado, err := repo.EstadoPostulacion(ctx, b.ID, p.ID)
		if err != nil {
			t.Fatalf("EstadoPostulacion: %v", err)
		}
		if estado.Estado != "BORRADOR" || estado.Version != b.Version+1 || estado.Elegibilidad == nil ||
			!estado.Elegibilidad.Elegible || estado.Elegibilidad.CargoID != 1 || estado.Elegibilidad.Mensaje != "Postulante afiliado al IESS" {
			t.Fatalf("estado inesperado: %+v %+v", estado, estado.Elegibilidad)
		}
		// Selección definitiva: no se puede volver a confirmar.
		if err := repo.ConfirmarCargo(ctx, b.ID, p.ID, domain.ResultadoElegibilidad{CargoID: 2, Elegible: true}, estado.Version); !errors.Is(err, domain.ErrCargoYaConfirmado) {
			t.Fatalf("esperado ErrCargoYaConfirmado, obtenido %v", err)
		}
	})

	t.Run("no elegible cierra la postulación y queda auditado", func(t *testing.T) {
		p, b := nuevoBorrador(t, "noelegible")
		// La cédula de prueba no está en el mock del IESS: no cumple ninguna calidad.
		svc := application.NewCargoService(database.NewPgxTxManager(pool), repo, database.NewPostgresAuditoriaRepository(pool),
			application.NewEligibilityService(external.NewMockEmploymentProvider()), time.Second)
		sol := domain.ContextoSolicitud{IP: "192.0.2.10", UserAgent: "prueba-integracion", TraceID: "55555555-5555-4555-8555-555555555555"}

		estado, err := svc.ConfirmarCargo(ctx, sol, b.ID, p.ID, 2, true, b.Version)
		var noElegible *domain.EligibilityError
		if !errors.As(err, &noElegible) || noElegible.Code != "POS_NO_ELEGIBLE_JUB" {
			t.Fatalf("esperado POS_NO_ELEGIBLE_JUB, obtenido %v", err)
		}
		if estado.Estado != "NO_ELEGIBLE" || estado.Elegibilidad == nil || estado.Elegibilidad.Elegible || estado.Elegibilidad.Mensaje != noElegible.Mensaje {
			t.Fatalf("estado inesperado: %+v %+v", estado, estado.Elegibilidad)
		}

		var accion, actorTipo, ip, ua, trace, despues string
		err = pool.QueryRow(ctx, `
			SELECT accion, actor_tipo, host(ip), user_agent, trace_id::text, datos_despues->>'estado'
			FROM auditoria_evento WHERE entidad = 'postulacion' AND entidad_id = $1 AND actor_id = $2::uuid`,
			b.ID, p.ID).Scan(&accion, &actorTipo, &ip, &ua, &trace, &despues)
		if err != nil {
			t.Fatalf("no se encontró el evento de auditoría: %v", err)
		}
		if accion != domain.AccionPostulacionNoElegible || actorTipo != domain.ActorPostulante || ip != sol.IP ||
			ua != sol.UserAgent || trace != sol.TraceID || despues != "NO_ELEGIBLE" {
			t.Fatalf("evento inesperado: %s %s %s %s %s %s", accion, actorTipo, ip, ua, trace, despues)
		}
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM auditoria_evento WHERE entidad_id = $1`, b.ID)
		})

		if err := repo.ConfirmarCargo(ctx, b.ID, p.ID, domain.ResultadoElegibilidad{CargoID: 1, Elegible: true}, estado.Version); !errors.Is(err, domain.ErrPostulacionCerrada) {
			t.Fatalf("esperado ErrPostulacionCerrada, obtenido %v", err)
		}
	})

	t.Run("si falla la auditoría no se guarda el cargo", func(t *testing.T) {
		p, b := nuevoBorrador(t, "rollback")
		svc := application.NewCargoService(database.NewPgxTxManager(pool), repo, auditoriaQueFalla{},
			application.NewEligibilityService(external.NewMockEmploymentProvider()), time.Second)
		if _, err := svc.ConfirmarCargo(ctx, domain.ContextoSolicitud{}, b.ID, p.ID, 2, true, b.Version); err == nil || errors.As(err, new(*domain.EligibilityError)) {
			t.Fatalf("esperado error de auditoría, obtenido %v", err)
		}
		actual, err := repo.EstadoPostulacion(ctx, b.ID, p.ID)
		if err != nil || actual.Estado != "BORRADOR" || actual.Elegibilidad != nil || actual.Version != b.Version {
			t.Fatalf("el rollback debía dejar la postulación intacta: %+v %v", actual, err)
		}
	})

	t.Run("rechazos", func(t *testing.T) {
		p, b := nuevoBorrador(t, "rechazos")
		if _, err := repo.ContextoConfirmacion(ctx, b.ID, p.ID, 99); !errors.Is(err, domain.ErrCargoInvalido) {
			t.Fatalf("cargo inexistente: esperado ErrCargoInvalido, obtenido %v", err)
		}
		if _, err := repo.ContextoConfirmacion(ctx, b.ID, "33333333-3333-3333-3333-333333333333", 1); !errors.Is(err, domain.ErrAccesoDenegado) {
			t.Fatalf("postulación ajena: esperado ErrAccesoDenegado, obtenido %v", err)
		}
		if _, err := repo.ContextoConfirmacion(ctx, "44444444-4444-4444-4444-444444444444", p.ID, 1); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("postulación inexistente: esperado ErrNotFound, obtenido %v", err)
		}
		if err := repo.ConfirmarCargo(ctx, b.ID, p.ID, domain.ResultadoElegibilidad{CargoID: 1, Elegible: true}, b.Version+5); !errors.Is(err, domain.ErrConflictoVersion) {
			t.Fatalf("versión desactualizada: esperado ErrConflictoVersion, obtenido %v", err)
		}
		actual, err := repo.ObtenerActual(ctx, p.ID)
		if err != nil || actual.Elegibilidad != nil || actual.Version != b.Version {
			t.Fatalf("un rechazo no debía modificar la postulación: %+v %v", actual, err)
		}
	})
}
