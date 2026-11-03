package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/security"
)

// LoginResult representa la respuesta del inicio de sesión (SDD 07).
type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"` // Encriptado en Cookie HttpOnly
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // 900 segundos (15 min)
	PostulanteID string `json:"postulante_id"`
	Cedula       string `json:"cedula"`
	Email        string `json:"email"`
}

// RefreshResult representa la respuesta de la rotación de token (ADR-006).
type RefreshResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type memorySessionStore struct {
	mu     sync.RWMutex
	tokens map[string]string
}

func newMemorySessionStore() *memorySessionStore {
	return &memorySessionStore{tokens: make(map[string]string)}
}

func (m *memorySessionStore) SaveRefreshToken(ctx context.Context, token, postulanteID string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[token] = postulanteID
	return nil
}

func (m *memorySessionStore) GetRefreshToken(ctx context.Context, token string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.tokens[token]
	if !ok {
		return "", domain.ErrSessionExpired
	}
	return id, nil
}

func (m *memorySessionStore) DeleteRefreshToken(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, token)
	return nil
}

type AuthService struct {
	repo         domain.PostulanteRepository
	validator    domain.DocumentValidationStrategy
	sessionStore domain.SessionStore
	attemptStore domain.AuthAttemptStore
	identity     domain.IdentityProvider // Registro Civil (SDD 7.4.1)

	// Fallback para usuarios semilla en pruebas si repo no está inicializado
	seedUsersMu sync.RWMutex
	seedUsers   map[string]seedUser
}

type seedUser struct {
	postulanteID string
	cedula       string
	email        string
	passwordHash string
}

// NewAuthService construye el servicio inyectando repositorios, estrategia y sesión (ADR-003, ADR-006).
func NewAuthService(repo domain.PostulanteRepository, validator domain.DocumentValidationStrategy, sessionStore ...domain.SessionStore) *AuthService {
	if validator == nil {
		validator = domain.NewCedulaEcuatorianaStrategy()
	}

	var store domain.SessionStore
	if len(sessionStore) > 0 && sessionStore[0] != nil {
		store = sessionStore[0]
	} else {
		store = newMemorySessionStore()
	}

	svc := &AuthService{
		repo:         repo,
		validator:    validator,
		sessionStore: store,
		attemptStore: noopAuthAttemptStore{},
		seedUsers:    make(map[string]seedUser),
	}

	// Semilla didáctica de prueba para desarrollo y demostración (SDD 9.2)
	defaultHash, _ := security.HashPassword("BiessPostulante2026*")
	svc.seedUsers["0926687856"] = seedUser{
		postulanteID: "usr-0926687856",
		cedula:       "0926687856",
		email:        "postulante.test@biess.gob.ec",
		passwordHash: defaultHash,
	}
	return svc
}

func (s *AuthService) WithAuthAttemptStore(store domain.AuthAttemptStore) *AuthService {
	if store != nil {
		s.attemptStore = store
	}
	return s
}

type noopAuthAttemptStore struct{}

func (noopAuthAttemptStore) CheckBlocked(context.Context, domain.AuthFlow, string) (time.Duration, error) {
	return 0, nil
}

func (noopAuthAttemptStore) RecordFailure(context.Context, domain.AuthFlow, string) (bool, time.Duration, error) {
	return false, 0, nil
}

func (noopAuthAttemptStore) Reset(context.Context, domain.AuthFlow, string) error {
	return nil
}

func (s *AuthService) checkAuthBlocked(ctx context.Context, flow domain.AuthFlow, cedula string) error {
	retryAfter, err := s.attemptStore.CheckBlocked(ctx, flow, cedula)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrAuthProtectionDown, err)
	}
	if retryAfter > 0 {
		return &domain.AuthBlockedError{Flow: flow, RetryAfter: retryAfter}
	}
	return nil
}

func (s *AuthService) recordAuthFailure(ctx context.Context, flow domain.AuthFlow, cedula string) error {
	blocked, retryAfter, err := s.attemptStore.RecordFailure(ctx, flow, cedula)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrAuthProtectionDown, err)
	}
	if blocked {
		return &domain.AuthBlockedError{Flow: flow, RetryAfter: retryAfter}
	}
	return nil
}

func (s *AuthService) resetAuthFailures(ctx context.Context, flow domain.AuthFlow, cedula string) error {
	if err := s.attemptStore.Reset(ctx, flow, cedula); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrAuthProtectionDown, err)
	}
	return nil
}

type TokenClaims struct {
	PostulanteID string `json:"postulante_id"`
	Cedula       string `json:"cedula"`
	Email        string `json:"email"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "srp_biess_dev_secret_key_change_in_prod"
	}
	return []byte(secret)
}

// Login autentica las credenciales del postulante y emite Access Token (15 min) + Refresh Token (8 h) (ADR-006).
func (s *AuthService) Login(ctx context.Context, cedula, password string) (*LoginResult, error) {
	// 1. Validar formato de cédula con la estrategia inyectada (Fail-Fast Módulo 10)
	if err := s.ValidarIdentidad(cedula); err != nil {
		return nil, domain.ErrInvalidCredentials // Mensaje estándar SDD 9.2.3 para evitar enumeración
	}
	if err := s.checkAuthBlocked(ctx, domain.AuthFlowLogin, cedula); err != nil {
		return nil, err
	}

	var postulanteID, email, passHash string

	if s.repo != nil {
		p, err := s.repo.GetByCedula(ctx, cedula)
		if err != nil || p == nil {
			if recordErr := s.recordAuthFailure(ctx, domain.AuthFlowLogin, cedula); recordErr != nil {
				return nil, recordErr
			}
			return nil, domain.ErrInvalidCredentials
		}
		postulanteID = p.ID
		email = p.Email
		passHash = p.PasswordHash
	} else {
		// Modo scaffold con usuarios semilla
		s.seedUsersMu.RLock()
		u, ok := s.seedUsers[cedula]
		s.seedUsersMu.RUnlock()
		if !ok {
			if recordErr := s.recordAuthFailure(ctx, domain.AuthFlowLogin, cedula); recordErr != nil {
				return nil, recordErr
			}
			return nil, domain.ErrInvalidCredentials
		}
		postulanteID = u.postulanteID
		email = u.email
		passHash = u.passwordHash
	}

	// 2. Verificar contraseña en tiempo constante con Argon2id
	match, err := s.VerifyPassword(password, passHash)
	if err != nil || !match {
		if recordErr := s.recordAuthFailure(ctx, domain.AuthFlowLogin, cedula); recordErr != nil {
			return nil, recordErr
		}
		return nil, domain.ErrInvalidCredentials
	}
	if err := s.resetAuthFailures(ctx, domain.AuthFlowLogin, cedula); err != nil {
		return nil, err
	}

	// 3. Generar Access Token (15 min)
	accessToken, err := s.GenerarAccessToken(postulanteID, cedula, email)
	if err != nil {
		return nil, fmt.Errorf("error generando access token: %w", err)
	}

	// 4. Generar Refresh Token opaco de 256 bits
	refreshToken, err := s.GenerarRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("error generando refresh token: %w", err)
	}

	// 5. Persistir Refresh Token en el SessionStore con TTL de 8 horas
	ttl := 8 * time.Hour
	if err := s.sessionStore.SaveRefreshToken(ctx, refreshToken, postulanteID, ttl); err != nil {
		return nil, fmt.Errorf("error guardando sesión: %w", err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutos en segundos
		PostulanteID: postulanteID,
		Cedula:       cedula,
		Email:        email,
	}, nil
}

// RefreshSession renueva el Access Token y ROTA el Refresh Token de forma atómica (ADR-006).
func (s *AuthService) RefreshSession(ctx context.Context, oldRefreshToken string) (*RefreshResult, error) {
	if oldRefreshToken == "" {
		return nil, domain.ErrSessionExpired
	}

	// 1. Validar existencia del token actual
	postulanteID, err := s.sessionStore.GetRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return nil, domain.ErrSessionExpired
	}

	// 2. Rotación: Invalida el token anterior para evitar replay attacks
	_ = s.sessionStore.DeleteRefreshToken(ctx, oldRefreshToken)

	// 3. Obtener metadatos del postulante
	cedula := "0926687856"
	email := "postulante@biess.gob.ec"

	if s.repo != nil {
		p, err := s.repo.GetByID(ctx, postulanteID)
		if err == nil && p != nil {
			cedula = p.Cedula
			email = p.Email
		}
	} else {
		s.seedUsersMu.RLock()
		for _, u := range s.seedUsers {
			if u.postulanteID == postulanteID {
				cedula = u.cedula
				email = u.email
				break
			}
		}
		s.seedUsersMu.RUnlock()
	}

	// 4. Emitir nuevo par de tokens
	newAccessToken, err := s.GenerarAccessToken(postulanteID, cedula, email)
	if err != nil {
		return nil, fmt.Errorf("error renovando access token: %w", err)
	}

	newRefreshToken, err := s.GenerarRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("error generando nuevo refresh token: %w", err)
	}

	// 5. Guardar nuevo Refresh Token (8 horas)
	_ = s.sessionStore.SaveRefreshToken(ctx, newRefreshToken, postulanteID, 8*time.Hour)

	return &RefreshResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}

// Logout revoca el Refresh Token en el almacén de sesiones (ADR-006).
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.sessionStore.DeleteRefreshToken(ctx, refreshToken)
}

// RegisterSeedUser permite registrar usuarios en memoria para tests.
func (s *AuthService) RegisterSeedUser(postulanteID, cedula, email, plainPassword string) error {
	hash, err := s.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	s.seedUsersMu.Lock()
	s.seedUsers[cedula] = seedUser{
		postulanteID: postulanteID,
		cedula:       cedula,
		email:        email,
		passwordHash: hash,
	}
	s.seedUsersMu.Unlock()
	return nil
}

// GenerarAccessToken crea un Access Token JWT de corta duración (15 minutos) conforme a ADR-006.
func (s *AuthService) GenerarAccessToken(postulanteID, cedula, email string) (string, error) {
	claims := TokenClaims{
		PostulanteID: postulanteID,
		Cedula:       cedula,
		Email:        email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // jti único por token (SDD 7.2.2)
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "srp-biess-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(getJWTSecret())
	if err != nil {
		return "", fmt.Errorf("error firmando token: %w", err)
	}

	return tokenStr, nil
}

// GenerarJWT se mantiene por retrocompatibilidad delegando a GenerarAccessToken.
func (s *AuthService) GenerarJWT(postulanteID, cedula, email string) (string, error) {
	return s.GenerarAccessToken(postulanteID, cedula, email)
}

// ValidarAccessToken parsea, verifica la firma HMAC y valida la expiración del Access Token.
func (s *AuthService) ValidarAccessToken(tokenStr string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", t.Header["alg"])
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("claims de token inválidos")
	}

	return claims, nil
}

// GenerarRefreshToken crea un token opaco de alta entropía (32 bytes) para rotación en cookie HttpOnly (ADR-006).
func (s *AuthService) GenerarRefreshToken() (string, error) {
	return security.GenerateOpaqueToken()
}

// HashPassword deriva la contraseña usando Argon2id (ADR-006).
func (s *AuthService) HashPassword(password string) (string, error) {
	return security.HashPassword(password)
}

// VerifyPassword valida la contraseña en texto plano contra el hash Argon2id en tiempo constante.
func (s *AuthService) VerifyPassword(password, hash string) (bool, error) {
	return security.VerifyPassword(password, hash)
}

// ValidarIdentidad delega la validación a la estrategia inyectada (SRP + Strategy Pattern).
func (s *AuthService) ValidarIdentidad(documento string) error {
	valido, err := s.validator.Validate(documento)
	if err != nil || !valido {
		return domain.ErrInvalidCedula
	}
	return nil
}

// SetValidator permite intercambiar la estrategia de validación dinámicamente.
func (s *AuthService) SetValidator(validator domain.DocumentValidationStrategy) {
	if validator != nil {
		s.validator = validator
	}
}
