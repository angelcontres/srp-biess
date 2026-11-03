package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/security"
)

// TokenValidacionTTL es la vigencia del token que habilita el paso 2 del registro (SDD 7.4.2).
const TokenValidacionTTL = 10 * time.Minute

const propositoRegistro = "registro"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func validarFormatoCorreo(email string) bool {
	return emailRegex.MatchString(email)
}

// RegistroClaims son los claims del token de validación de identidad.
type RegistroClaims struct {
	Cedula          string     `json:"cedula"`
	Nombres         string     `json:"nombres,omitempty"`
	Apellidos       string     `json:"apellidos,omitempty"`
	Genero          string     `json:"genero,omitempty"`
	FechaNacimiento *time.Time `json:"fecha_nacimiento,omitempty"`
	Nacionalidad    string     `json:"nacionalidad,omitempty"`
	EstadoCivil     string     `json:"estado_civil,omitempty"`
	Proposito       string     `json:"proposito"`
	jwt.RegisteredClaims
}

// IdentidadValidada es el resultado exitoso de POST /auth/registro/validar-identidad.
type IdentidadValidada struct {
	TokenValidacion string
	Datos           *domain.IdentityResult
}

// WithIdentityProvider inyecta el adaptador del Registro Civil (ADR-003, ADR-010).
func (s *AuthService) WithIdentityProvider(identity domain.IdentityProvider) *AuthService {
	s.identity = identity
	return s
}

// ValidarIdentidadRegistro contrasta cédula + dato con el Registro Civil (SDD 7.4.1).
// Aplica primero el Módulo 10 (SRP-008): si falla, no se llama al proveedor externo.
func (s *AuthService) ValidarIdentidadRegistro(ctx context.Context, cedula string, tipo domain.TipoValidacion, valor string) (*IdentidadValidada, error) {
	if !tipo.EsValido() {
		return nil, domain.ErrInvalidTipoValidacion
	}
	if err := s.ValidarIdentidad(cedula); err != nil {
		return nil, domain.ErrInvalidCedula
	}
	if err := s.checkAuthBlocked(ctx, domain.AuthFlowRegistration, cedula); err != nil {
		return nil, err
	}
	if s.identity == nil {
		return nil, domain.ErrExternalUnavailable
	}

	res, err := s.identity.Validate(ctx, cedula, tipo, valor)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTipoValidacion) {
			return nil, err
		}
		return nil, domain.ErrExternalUnavailable
	}
	if res == nil || !res.Match {
		if recordErr := s.recordAuthFailure(ctx, domain.AuthFlowRegistration, cedula); recordErr != nil {
			return nil, recordErr
		}
		return nil, domain.ErrIdentityMismatch
	}
	if err := s.resetAuthFailures(ctx, domain.AuthFlowRegistration, cedula); err != nil {
		return nil, err
	}

	token, err := s.generarTokenValidacion(cedula, res)
	if err != nil {
		return nil, err
	}
	return &IdentidadValidada{TokenValidacion: token, Datos: res}, nil
}

func (s *AuthService) generarTokenValidacion(cedula string, datos *domain.IdentityResult) (string, error) {
	now := time.Now()
	claims := RegistroClaims{
		Cedula:    cedula,
		Proposito: propositoRegistro,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   cedula,
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenValidacionTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "srp-biess-api",
		},
	}
	if datos != nil {
		claims.Nombres = datos.Nombres
		claims.Apellidos = datos.Apellidos
		claims.Genero = datos.Genero
		if !datos.FechaNacimiento.IsZero() {
			fn := datos.FechaNacimiento
			claims.FechaNacimiento = &fn
		}
		claims.Nacionalidad = datos.Nacionalidad
		claims.EstadoCivil = datos.EstadoCivil
	}
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(getJWTSecret())
	if err != nil {
		return "", fmt.Errorf("error firmando token de validación: %w", err)
	}
	return tokenStr, nil
}

// ValidarTokenValidacion verifica firma, vigencia y propósito del token (lo usa POST /auth/registro, SDD 7.4.2).
func (s *AuthService) ValidarTokenValidacion(tokenStr string) (*RegistroClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RegistroClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", t.Header["alg"])
		}
		return getJWTSecret(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("token de validación inválido: %w", err)
	}
	claims, ok := token.Claims.(*RegistroClaims)
	if !ok || !token.Valid || claims.Proposito != propositoRegistro {
		return nil, errors.New("token de validación inválido")
	}
	return claims, nil
}

// Registrar implementa el paso 2 del registro de cuenta (SDD 7.4.2).
func (s *AuthService) Registrar(ctx context.Context, tokenValidacion, correo, password string) (*domain.Postulante, error) {
	claims, err := s.ValidarTokenValidacion(tokenValidacion)
	if err != nil {
		return nil, domain.ErrTokenInvalido
	}

	correo = strings.ToLower(strings.TrimSpace(correo))
	if correo == "" || !validarFormatoCorreo(correo) {
		return nil, domain.ErrCorreoInvalido
	}

	if err := domain.ValidarPoliticaPassword(password, claims.Cedula, correo); err != nil {
		return nil, err
	}

	if s.repo != nil {
		existente, _ := s.repo.GetByCedula(ctx, claims.Cedula)
		if existente != nil {
			return nil, domain.ErrCedulaDuplicada
		}
		existenteEmail, _ := s.repo.GetByEmail(ctx, correo)
		if existenteEmail != nil {
			return nil, domain.ErrCorreoDuplicado
		}
	} else {
		s.seedUsersMu.RLock()
		if _, ok := s.seedUsers[claims.Cedula]; ok {
			s.seedUsersMu.RUnlock()
			return nil, domain.ErrCedulaDuplicada
		}
		for _, u := range s.seedUsers {
			if strings.EqualFold(u.email, correo) {
				s.seedUsersMu.RUnlock()
				return nil, domain.ErrCorreoDuplicado
			}
		}
		s.seedUsersMu.RUnlock()
	}

	hash, err := s.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error generando hash de contraseña: %w", err)
	}

	postulante := &domain.Postulante{
		Cedula:          claims.Cedula,
		Nombres:         claims.Nombres,
		Apellidos:       claims.Apellidos,
		Genero:          claims.Genero,
		FechaNacimiento: claims.FechaNacimiento,
		Nacionalidad:    claims.Nacionalidad,
		EstadoCivil:     claims.EstadoCivil,
		Email:           correo,
		PasswordHash:    hash,
	}

	if s.repo != nil {
		if err := s.repo.Save(ctx, postulante); err != nil {
			return nil, fmt.Errorf("error guardando postulante: %w", err)
		}
	} else {
		postulante.ID = uuid.New().String()
		s.seedUsersMu.Lock()
		s.seedUsers[claims.Cedula] = seedUser{
			postulanteID: postulante.ID,
			cedula:       claims.Cedula,
			email:        correo,
			passwordHash: hash,
		}
		s.seedUsersMu.Unlock()
	}

	return postulante, nil
}

// SolicitarRecuperacionPassword inicia el flujo de recuperación (SDD 10.5).
// Responde siempre de forma genérica para evitar la enumeración de cuentas (SDD 9.2.3).
func (s *AuthService) SolicitarRecuperacionPassword(ctx context.Context, identificador string) (string, error) {
	identificador = strings.TrimSpace(identificador)
	if identificador == "" {
		return "", nil
	}

	var postulante *domain.Postulante
	if s.repo != nil {
		p, err := s.repo.GetByCedula(ctx, identificador)
		if err == nil && p != nil {
			postulante = p
		} else {
			p, err = s.repo.GetByEmail(ctx, strings.ToLower(identificador))
			if err == nil && p != nil {
				postulante = p
			}
		}
	} else {
		s.seedUsersMu.RLock()
		if u, ok := s.seedUsers[identificador]; ok {
			postulante = &domain.Postulante{ID: u.postulanteID, Cedula: u.cedula, Email: u.email}
		} else {
			for _, u := range s.seedUsers {
				if strings.EqualFold(u.email, identificador) {
					postulante = &domain.Postulante{ID: u.postulanteID, Cedula: u.cedula, Email: u.email}
					break
				}
			}
		}
		s.seedUsersMu.RUnlock()
	}

	if postulante == nil {
		// Silent return para evitar enumeración (SDD 9.2.3)
		return "", nil
	}

	token, err := security.GenerateOpaqueToken()
	if err != nil {
		return "", fmt.Errorf("error generando token de recuperación: %w", err)
	}

	if s.resetStore != nil {
		if err := s.resetStore.SaveResetToken(ctx, token, postulante.ID, 30*time.Minute); err != nil {
			return "", fmt.Errorf("error persistiendo token de recuperación: %w", err)
		}
	}

	log.Printf("[INFO] Solicitud de recuperación para postulante %s. Token temporal (30 min): %s", postulante.ID, token)
	return token, nil
}

// RestablecerPassword valida el token de recuperación y actualiza la contraseña (SDD 10.5).
func (s *AuthService) RestablecerPassword(ctx context.Context, token, nuevaPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.ErrTokenInvalido
	}

	if s.resetStore == nil {
		return domain.ErrAuthProtectionDown
	}

	postulanteID, err := s.resetStore.GetResetToken(ctx, token)
	if err != nil || postulanteID == "" {
		return domain.ErrTokenInvalido
	}

	var cedula, email string
	if s.repo != nil {
		p, err := s.repo.GetByID(ctx, postulanteID)
		if err != nil || p == nil {
			return domain.ErrNotFound
		}
		cedula = p.Cedula
		email = p.Email
	}

	if err := domain.ValidarPoliticaPassword(nuevaPassword, cedula, email); err != nil {
		return err
	}

	hash, err := s.HashPassword(nuevaPassword)
	if err != nil {
		return fmt.Errorf("error generando hash: %w", err)
	}

	if s.repo != nil {
		if err := s.repo.UpdatePassword(ctx, postulanteID, hash); err != nil {
			return fmt.Errorf("error actualizando contraseña: %w", err)
		}
	} else {
		s.seedUsersMu.Lock()
		for k, u := range s.seedUsers {
			if u.postulanteID == postulanteID {
				u.passwordHash = hash
				s.seedUsers[k] = u
				break
			}
		}
		s.seedUsersMu.Unlock()
	}

	// Invalidar token de recuperación usado
	_ = s.resetStore.DeleteResetToken(ctx, token)
	return nil
}
