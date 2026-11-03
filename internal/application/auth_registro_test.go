package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
)

const cedulaRegistro = "1710034065"

// registroCivilFake simula el Registro Civil y cuenta las llamadas recibidas.
type registroCivilFake struct {
	res     *domain.IdentityResult
	err     error
	llamado int
}

func (f *registroCivilFake) Validate(context.Context, string, domain.TipoValidacion, string) (*domain.IdentityResult, error) {
	f.llamado++
	return f.res, f.err
}

// intentosFake permite fijar un bloqueo vigente o un Redis caído.
type intentosFake struct {
	bloqueo   time.Duration
	err       error
	fallos    int
	reinicios int
}

func (f *intentosFake) CheckBlocked(context.Context, domain.AuthFlow, string) (time.Duration, error) {
	return f.bloqueo, f.err
}

func (f *intentosFake) RecordFailure(context.Context, domain.AuthFlow, string) (bool, time.Duration, error) {
	if f.err != nil {
		return false, 0, f.err
	}
	f.fallos++
	if f.fallos >= 3 {
		return true, 10 * time.Minute, nil
	}
	return false, 0, nil
}

func (f *intentosFake) Reset(context.Context, domain.AuthFlow, string) error {
	if f.err != nil {
		return f.err
	}
	f.reinicios++
	f.fallos = 0
	return nil
}

func servicioRegistro(rc domain.IdentityProvider, intentos domain.AuthAttemptStore) *application.AuthService {
	return application.NewAuthService(nil, nil).WithIdentityProvider(rc).WithAuthAttemptStore(intentos)
}

func TestValidarIdentidadRegistro_EntradasInvalidasNoLlamanAlRegistroCivil(t *testing.T) {
	rc := &registroCivilFake{res: &domain.IdentityResult{Match: true}}
	svc := servicioRegistro(rc, &intentosFake{})
	ctx := context.Background()

	if _, err := svc.ValidarIdentidadRegistro(ctx, cedulaRegistro, "OTRO", "x"); !errors.Is(err, domain.ErrInvalidTipoValidacion) {
		t.Errorf("tipo inválido: esperado ErrInvalidTipoValidacion, obtenido %v", err)
	}
	// SRP-008: si falla el Módulo 10 no se consulta al proveedor externo.
	if _, err := svc.ValidarIdentidadRegistro(ctx, "1710034060", domain.ValidacionCodigoDactilar, "x"); !errors.Is(err, domain.ErrInvalidCedula) {
		t.Errorf("cédula inválida: esperado ErrInvalidCedula, obtenido %v", err)
	}
	if rc.llamado != 0 {
		t.Errorf("el Registro Civil fue llamado %d veces con entradas inválidas", rc.llamado)
	}
}

func TestValidarIdentidadRegistro_ProveedorNoDisponible(t *testing.T) {
	ctx := context.Background()

	sinProveedor := servicioRegistro(nil, &intentosFake{})
	if _, err := sinProveedor.ValidarIdentidadRegistro(ctx, cedulaRegistro, domain.ValidacionCodigoDactilar, "x"); !errors.Is(err, domain.ErrExternalUnavailable) {
		t.Errorf("sin proveedor: esperado ErrExternalUnavailable, obtenido %v", err)
	}

	// SDD 10.3: un error técnico del proveedor no consume intento.
	intentos := &intentosFake{}
	conFalla := servicioRegistro(&registroCivilFake{err: errors.New("timeout")}, intentos)
	if _, err := conFalla.ValidarIdentidadRegistro(ctx, cedulaRegistro, domain.ValidacionCodigoDactilar, "x"); !errors.Is(err, domain.ErrExternalUnavailable) {
		t.Errorf("proveedor caído: esperado ErrExternalUnavailable, obtenido %v", err)
	}
	if intentos.fallos != 0 {
		t.Errorf("un error técnico no debe contar como intento fallido, fallos=%d", intentos.fallos)
	}
}

func TestValidarIdentidadRegistro_TercerRechazoBloquea(t *testing.T) {
	ctx := context.Background()
	intentos := &intentosFake{}
	svc := servicioRegistro(&registroCivilFake{res: &domain.IdentityResult{Match: false}}, intentos)

	for i := 1; i <= 2; i++ {
		if _, err := svc.ValidarIdentidadRegistro(ctx, cedulaRegistro, domain.ValidacionCodigoDactilar, "x"); !errors.Is(err, domain.ErrIdentityMismatch) {
			t.Fatalf("rechazo %d: esperado ErrIdentityMismatch, obtenido %v", i, err)
		}
	}
	_, err := svc.ValidarIdentidadRegistro(ctx, cedulaRegistro, domain.ValidacionCodigoDactilar, "x")
	var bloqueo *domain.AuthBlockedError
	if !errors.As(err, &bloqueo) || bloqueo.Flow != domain.AuthFlowRegistration || bloqueo.RetryAfter != 10*time.Minute {
		t.Fatalf("tercer rechazo: esperado AuthBlockedError de registro por 10 min, obtenido %v", err)
	}
}

func TestValidarIdentidadRegistro_BloqueoVigenteNoConsultaAlRegistroCivil(t *testing.T) {
	rc := &registroCivilFake{res: &domain.IdentityResult{Match: true}}
	svc := servicioRegistro(rc, &intentosFake{bloqueo: 7 * time.Minute})

	_, err := svc.ValidarIdentidadRegistro(context.Background(), cedulaRegistro, domain.ValidacionCodigoDactilar, "x")
	var bloqueo *domain.AuthBlockedError
	if !errors.As(err, &bloqueo) || bloqueo.RetryAfter != 7*time.Minute {
		t.Fatalf("esperado AuthBlockedError con 7 min restantes, obtenido %v", err)
	}
	if rc.llamado != 0 {
		t.Error("con bloqueo vigente no se debe consultar al Registro Civil")
	}
}

func TestValidarIdentidadRegistro_RedisCaidoFallaCerrado(t *testing.T) {
	rc := &registroCivilFake{res: &domain.IdentityResult{Match: true}}
	svc := servicioRegistro(rc, &intentosFake{err: errors.New("redis caído")})

	_, err := svc.ValidarIdentidadRegistro(context.Background(), cedulaRegistro, domain.ValidacionCodigoDactilar, "x")
	if !errors.Is(err, domain.ErrAuthProtectionDown) {
		t.Fatalf("esperado ErrAuthProtectionDown, obtenido %v", err)
	}
	if rc.llamado != 0 {
		t.Error("sin protección de intentos no se debe consultar al Registro Civil")
	}
}

func TestValidarIdentidadRegistro_ExitoEmiteTokenYReiniciaIntentos(t *testing.T) {
	intentos := &intentosFake{fallos: 2}
	datos := &domain.IdentityResult{Match: true, Cedula: cedulaRegistro, Nombres: "ANA"}
	svc := servicioRegistro(&registroCivilFake{res: datos}, intentos)

	res, err := svc.ValidarIdentidadRegistro(context.Background(), cedulaRegistro, domain.ValidacionFechaExpedicion, "2018-06-15")
	if err != nil {
		t.Fatalf("ValidarIdentidadRegistro: %v", err)
	}
	if res.Datos != datos || intentos.reinicios != 1 || intentos.fallos != 0 {
		t.Errorf("resultado o reinicio inesperado: datos=%v reinicios=%d fallos=%d", res.Datos, intentos.reinicios, intentos.fallos)
	}

	claims, err := svc.ValidarTokenValidacion(res.TokenValidacion)
	if err != nil {
		t.Fatalf("el token de validación emitido no es válido: %v", err)
	}
	if claims.Cedula != cedulaRegistro || claims.Subject != cedulaRegistro || claims.Proposito != "registro" {
		t.Errorf("claims inesperados: %+v", claims)
	}
	if vence := time.Until(claims.ExpiresAt.Time); vence <= 0 || vence > application.TokenValidacionTTL {
		t.Errorf("vigencia fuera de los 10 minutos: %v", vence)
	}
}

func TestValidarTokenValidacion_RechazaOtrosTokens(t *testing.T) {
	svc := application.NewAuthService(nil, nil)

	// Un access token tiene firma válida pero no el propósito "registro".
	access, err := svc.GenerarAccessToken("pos-1", cedulaRegistro, "ana@correo.ec")
	if err != nil {
		t.Fatalf("GenerarAccessToken: %v", err)
	}
	for nombre, token := range map[string]string{"access token": access, "texto": "no-es-un-jwt", "vacío": ""} {
		if _, err := svc.ValidarTokenValidacion(token); err == nil {
			t.Errorf("%s: debía rechazarse como token de validación", nombre)
		}
	}
}
