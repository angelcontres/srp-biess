package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
)

// sessionStoreFake es un almacén de refresh tokens en memoria observable desde la prueba.
type sessionStoreFake struct {
	mu     sync.Mutex
	tokens map[string]string
}

func newSessionStoreFake() *sessionStoreFake {
	return &sessionStoreFake{tokens: map[string]string{}}
}

func (s *sessionStoreFake) SaveRefreshToken(_ context.Context, token, postulanteID string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = postulanteID
	return nil
}

func (s *sessionStoreFake) GetRefreshToken(_ context.Context, token string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.tokens[token]
	if !ok {
		return "", domain.ErrSessionExpired
	}
	return id, nil
}

func (s *sessionStoreFake) DeleteRefreshToken(_ context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
	return nil
}

// postulanteRepoFake solo responde GetByID, que es lo que usa RefreshSession.
type postulanteRepoFake struct {
	domain.PostulanteRepository
	postulantes map[string]*domain.Postulante
}

func (r *postulanteRepoFake) GetByID(_ context.Context, id string) (*domain.Postulante, error) {
	if p, ok := r.postulantes[id]; ok {
		return p, nil
	}
	return nil, domain.ErrNotFound
}

func TestAuthService_RefreshSession_TokenVacioOInexistente(t *testing.T) {
	svc := application.NewAuthService(nil, nil, newSessionStoreFake())
	for _, token := range []string{"", "token-que-no-existe"} {
		if _, err := svc.RefreshSession(context.Background(), token); !errors.Is(err, domain.ErrSessionExpired) {
			t.Errorf("RefreshSession(%q): esperado ErrSessionExpired, obtenido %v", token, err)
		}
	}
}

// ADR-006: cada renovación rota el refresh token; el anterior queda inválido (anti-replay).
func TestAuthService_RefreshSession_RotaElToken(t *testing.T) {
	ctx := context.Background()
	store := newSessionStoreFake()
	svc := application.NewAuthService(nil, nil, store)

	login, err := svc.Login(ctx, "0926687856", "BiessPostulante2026*")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	renovado, err := svc.RefreshSession(ctx, login.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshSession: %v", err)
	}
	if renovado.RefreshToken == "" || renovado.RefreshToken == login.RefreshToken {
		t.Fatal("la renovación debe emitir un refresh token distinto")
	}
	if renovado.TokenType != "Bearer" || renovado.ExpiresIn != 900 {
		t.Errorf("respuesta inesperada: %+v", renovado)
	}

	claims, err := svc.ValidarAccessToken(renovado.AccessToken)
	if err != nil {
		t.Fatalf("el nuevo access token no es válido: %v", err)
	}
	if claims.PostulanteID != login.PostulanteID || claims.Cedula != "0926687856" || claims.Email != login.Email {
		t.Errorf("claims del usuario semilla incorrectos: %+v", claims)
	}

	if _, err := svc.RefreshSession(ctx, login.RefreshToken); !errors.Is(err, domain.ErrSessionExpired) {
		t.Errorf("reutilizar el refresh token anterior debe fallar, obtenido %v", err)
	}
	if _, err := svc.RefreshSession(ctx, renovado.RefreshToken); err != nil {
		t.Errorf("el refresh token nuevo debe servir para la siguiente renovación: %v", err)
	}
}

func TestAuthService_RefreshSession_ClaimsDesdeRepositorio(t *testing.T) {
	ctx := context.Background()
	store := newSessionStoreFake()
	repo := &postulanteRepoFake{postulantes: map[string]*domain.Postulante{
		"pos-1": {ID: "pos-1", Cedula: "1710034065", Email: "ana@correo.ec"},
	}}
	svc := application.NewAuthService(repo, nil, store)
	_ = store.SaveRefreshToken(ctx, "refresh-1", "pos-1", time.Hour)

	renovado, err := svc.RefreshSession(ctx, "refresh-1")
	if err != nil {
		t.Fatalf("RefreshSession: %v", err)
	}
	claims, err := svc.ValidarAccessToken(renovado.AccessToken)
	if err != nil {
		t.Fatalf("ValidarAccessToken: %v", err)
	}
	if claims.PostulanteID != "pos-1" || claims.Cedula != "1710034065" || claims.Email != "ana@correo.ec" {
		t.Errorf("claims deben venir del repositorio: %+v", claims)
	}
	if id, _ := store.GetRefreshToken(ctx, renovado.RefreshToken); id != "pos-1" {
		t.Errorf("el nuevo refresh token debe guardarse para el mismo postulante, guardado para %q", id)
	}
}

func TestAuthService_Logout(t *testing.T) {
	ctx := context.Background()
	store := newSessionStoreFake()
	svc := application.NewAuthService(nil, nil, store)

	if err := svc.Logout(ctx, ""); err != nil {
		t.Errorf("Logout sin token no debe fallar: %v", err)
	}

	login, err := svc.Login(ctx, "0926687856", "BiessPostulante2026*")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if err := svc.Logout(ctx, login.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, err := svc.RefreshSession(ctx, login.RefreshToken); !errors.Is(err, domain.ErrSessionExpired) {
		t.Errorf("tras Logout el refresh token debe quedar revocado, obtenido %v", err)
	}
}
