package handlers

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/interfaces/http/dto"
)

type AuthHandler struct {
	authService *application.AuthService
}

func NewAuthHandler(authService *application.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type ValidarCedulaRequest struct {
	Cedula string `json:"cedula"`
}

type LoginRequest struct {
	Cedula   string `json:"cedula"`
	Password string `json:"password"`
}

// responderControlIntentos traduce el bloqueo a 423 + Retry-After y la caída de Redis a 503 (SRP-006, SDD 9.5).
// Devuelve false si el error no corresponde al control de intentos.
func responderControlIntentos(c *fiber.Ctx, err error) (bool, error) {
	var bloqueo *domain.LockoutError
	if errors.As(err, &bloqueo) {
		segundos := int(math.Ceil(bloqueo.RetryAfter.Seconds()))
		code := "AUTH_REGISTRO_BLOQUEADO"
		if bloqueo.Scope == domain.ScopeLogin {
			code = "AUTH_LOGIN_BLOQUEADO"
		}
		c.Set(fiber.HeaderRetryAfter, strconv.Itoa(segundos))
		return true, c.Status(fiber.StatusLocked).JSON(fiber.Map{
			"success":           false,
			"code":              code,
			"error":             bloqueo.Error(),
			"retryAfterSeconds": segundos,
		})
	}
	if errors.Is(err, domain.ErrAttemptLimiterUnavailable) {
		return true, c.Status(fiber.StatusServiceUnavailable).JSON(dto.Failure(err.Error()))
	}
	return false, nil
}

// Login maneja la autenticación por cédula y contraseña emitiendo JWT (15 min) y cookie HttpOnly (8 h) (ADR-006, SDD 07).
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("Cuerpo de petición inválido"))
	}

	if req.Cedula == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("La cédula y la contraseña son obligatorias"))
	}

	result, err := h.authService.Login(c.Context(), req.Cedula, req.Password)
	if err != nil {
		if ok, resErr := responderControlIntentos(c, err); ok {
			return resErr
		}
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Failure("Sus credenciales de ingreso son incorrectas"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Failure("Error interno durante el inicio de sesión"))
	}

	// Establecer Refresh Token opaco en cookie segura HttpOnly (ADR-006 / SDD 7.1)
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(8 * time.Hour),
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Strict",
	})

	return c.JSON(dto.Success(fiber.Map{
		"access_token":  result.AccessToken,
		"token_type":    result.TokenType,
		"expires_in":    result.ExpiresIn,
		"postulante_id": result.PostulanteID,
		"cedula":        result.Cedula,
		"email":         result.Email,
	}, "Inicio de sesión exitoso"))
}

// Refresh renueva el Access Token y rota el Refresh Token opaco en la cookie HttpOnly (ADR-006).
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	cookie := c.Cookies("refresh_token")
	if cookie == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.Failure("Cookie de renovación (refresh_token) no encontrada"))
	}

	result, err := h.authService.RefreshSession(c.Context(), cookie)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.Failure("Sesión expirada o token de renovación no válido"))
	}

	// Rotación de cookie HttpOnly
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(8 * time.Hour),
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Strict",
	})

	return c.JSON(dto.Success(fiber.Map{
		"access_token": result.AccessToken,
		"token_type":   result.TokenType,
		"expires_in":   result.ExpiresIn,
	}, "Token renovado exitosamente"))
}

// Logout revoca la sesión activa en el almacén y elimina la cookie HttpOnly (ADR-006).
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	cookie := c.Cookies("refresh_token")
	if cookie != "" {
		_ = h.authService.Logout(c.Context(), cookie)
	}

	// Limpiar cookie en el cliente
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Strict",
		MaxAge:   -1,
	})

	return c.JSON(dto.Success(nil, "Sesión cerrada correctamente"))
}

// ValidarCedula valida el formato de documento con la estrategia Módulo 10.
// TODO(backend: SRP-008): alinear ruta y respuesta con SDD 7.6.1 (ver TODO en router.go).
func (h *AuthHandler) ValidarCedula(c *fiber.Ctx) error {
	var req ValidarCedulaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("Cuerpo de petición inválido"))
	}

	if err := h.authService.ValidarIdentidad(req.Cedula); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.Failure(err.Error()))
	}

	return c.JSON(dto.Success(fiber.Map{
		"cedula": req.Cedula,
		"valida": true,
	}, "Cédula verificada correctamente con algoritmo Módulo 10"))
}

// Me retorna la identidad del postulante autenticado mediante el contexto inyectado (SRP-003).
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	postulanteID, _ := c.Locals("postulante_id").(string)
	cedula, _ := c.Locals("cedula").(string)
	email, _ := c.Locals("email").(string)

	return c.JSON(dto.Success(fiber.Map{
		"postulante_id": postulanteID,
		"cedula":        cedula,
		"email":         email,
	}, "Sesión activa y válida"))
}

type ValidarIdentidadRequest struct {
	Cedula         string `json:"cedula"`
	TipoValidacion string `json:"tipoValidacion"`
	Valor          string `json:"valor"` // fecha en ISO (YYYY-MM-DD) o código dactilar
}

// ValidarIdentidadRegistro implementa POST /auth/registro/validar-identidad (SDD 7.4.1).
func (h *AuthHandler) ValidarIdentidadRegistro(c *fiber.Ctx) error {
	var req ValidarIdentidadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("Cuerpo de petición inválido"))
	}
	req.Cedula = strings.TrimSpace(req.Cedula)
	req.Valor = strings.TrimSpace(req.Valor)
	if req.Cedula == "" || req.Valor == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("La cédula y el dato de validación son obligatorios"))
	}

	res, err := h.authService.ValidarIdentidadRegistro(c.Context(), req.Cedula, domain.TipoValidacion(req.TipoValidacion), req.Valor)
	if err != nil {
		if ok, resErr := responderControlIntentos(c, err); ok {
			return resErr
		}
		switch {
		case errors.Is(err, domain.ErrInvalidTipoValidacion):
			return c.Status(fiber.StatusBadRequest).JSON(dto.Failure("Tipo de validación no soportado"))
		case errors.Is(err, domain.ErrInvalidCedula):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.Failure("La cédula ingresada no es válida"))
		case errors.Is(err, domain.ErrIdentityMismatch):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.Failure(domain.ErrIdentityMismatch.Error()))
		case errors.Is(err, domain.ErrExternalUnavailable):
			return c.Status(fiber.StatusServiceUnavailable).JSON(dto.Failure(domain.ErrExternalUnavailable.Error()))
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(dto.Failure("Error interno durante la validación de identidad"))
		}
	}

	d := res.Datos
	return c.JSON(dto.Success(fiber.Map{
		"identidadValidada": true,
		"tokenValidacion":   res.TokenValidacion,
		"datos": fiber.Map{
			"nombres":         d.Nombres,
			"apellidos":       d.Apellidos,
			"genero":          d.Genero,
			"fechaNacimiento": d.FechaNacimiento.Format("2006-01-02"),
			"nacionalidad":    d.Nacionalidad,
		},
	}, "Identidad validada con el Registro Civil"))
}
