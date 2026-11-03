package domain

import "errors"

// Errores comunes de dominio (ADR-003)
var (
	ErrNotFound              = errors.New("recurso no encontrado")
	ErrInvalidCedula         = errors.New("la cédula ecuatoriana ingresada no es válida según Módulo 10")
	ErrCedulaBlocked         = errors.New("cédula bloqueada temporalmente por exceso de intentos fallidos")
	ErrUnauthorized          = errors.New("credenciales inválidas o no autorizado")
	ErrPostulacionExists     = errors.New("ya existe una postulación activa para este proceso")
	ErrIneligibleCandidate   = errors.New("el candidato no cumple los requisitos mínimos de elegibilidad")
	ErrExternalUnavailable   = errors.New("servicio externo no disponible, inténtelo más tarde")
	ErrInvalidTipoValidacion = errors.New("tipo de validación no soportado")
	ErrInvalidSector         = errors.New("sector de cargo no válido")
)
