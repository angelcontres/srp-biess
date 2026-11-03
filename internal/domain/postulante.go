package domain

import (
	"context"
	"time"
)

// Postulante representa a un aspirante registrado en el sistema.
type Postulante struct {
	ID        string `json:"id"`
	Cedula    string `json:"cedula"`
	Nombres   string `json:"nombres"`
	Apellidos string `json:"apellidos"`
	// Datos del Registro Civil (RF-EF12-03): se guardan al registrarse y son de solo lectura.
	Genero             string     `json:"genero,omitempty"`           // MASCULINO, FEMENINO, OTRO, PREFIERO_NO_DECIR
	FechaNacimiento    *time.Time `json:"fecha_nacimiento,omitempty"` // fecha calendario
	Nacionalidad       string     `json:"nacionalidad,omitempty"`     // vacío: ECUATORIANA
	EstadoCivil        string     `json:"estado_civil,omitempty"`
	Email              string     `json:"email"`
	Telefono           string     `json:"telefono"`
	Direccion          string     `json:"direccion"`
	PaisID             string     `json:"pais_id"`
	ProvinciaID        string     `json:"provincia_id"`
	CantonID           string     `json:"canton_id"`
	PasswordHash       string     `json:"-"`
	TieneFamiliarBiess bool       `json:"tiene_familiar_biess"`
	DetalleFamiliar    string     `json:"detalle_familiar,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// PostulanteRepository define el puerto de persistencia para el postulante.
type PostulanteRepository interface {
	GetByID(ctx context.Context, id string) (*Postulante, error)
	GetByCedula(ctx context.Context, cedula string) (*Postulante, error)
	GetByEmail(ctx context.Context, email string) (*Postulante, error)
	Save(ctx context.Context, p *Postulante) error
	Update(ctx context.Context, p *Postulante) error
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
}
