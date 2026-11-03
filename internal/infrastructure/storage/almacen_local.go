package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"srp-biess/internal/domain"
)

// AlmacenLocal implementa domain.AlmacenDocumentos en disco (STORAGE_MODE=local, STORAGE_PATH).
// Los temporales viven en <raíz>/tmp, en el mismo sistema de archivos que los definitivos,
// para que Confirmar sea un rename atómico.
type AlmacenLocal struct {
	raiz string
	tmp  string
}

var _ domain.AlmacenDocumentos = (*AlmacenLocal)(nil)

// Solo nombres generados por el sistema: <uuid>.pdf o tmp-*.pdf. Evita path traversal.
var (
	nombreDefinitivo = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.pdf$`)
	nombreTemporal   = regexp.MustCompile(`^tmp-[0-9]+\.pdf$`)
)

var ErrNombreInvalido = errors.New("nombre de archivo no válido")

// NewAlmacenLocal crea los directorios con permisos privados (0700).
func NewAlmacenLocal(raiz string) (*AlmacenLocal, error) {
	abs, err := filepath.Abs(raiz)
	if err != nil {
		return nil, fmt.Errorf("ruta de almacenamiento inválida: %w", err)
	}
	a := &AlmacenLocal{raiz: filepath.Join(abs, "documentos"), tmp: filepath.Join(abs, "documentos", "tmp")}
	if err := os.MkdirAll(a.tmp, 0o700); err != nil {
		return nil, fmt.Errorf("no se pudo crear el almacenamiento de documentos: %w", err)
	}
	return a, nil
}

func (a *AlmacenLocal) EscribirTemporal(r io.Reader) (string, int64, error) {
	f, err := os.CreateTemp(a.tmp, "tmp-*.pdf")
	if err != nil {
		return "", 0, fmt.Errorf("no se pudo crear el temporal: %w", err)
	}
	n, copiaErr := io.Copy(f, r)
	cierreErr := f.Close()
	ref := filepath.Base(f.Name())
	if copiaErr != nil || cierreErr != nil {
		_ = os.Remove(f.Name())
		return "", 0, fmt.Errorf("no se pudo escribir el temporal: %w", errors.Join(copiaErr, cierreErr))
	}
	return ref, n, nil
}

func (a *AlmacenLocal) Confirmar(ref, nombre string) (string, error) {
	if !nombreTemporal.MatchString(ref) || !nombreDefinitivo.MatchString(nombre) {
		return "", ErrNombreInvalido
	}
	destino := filepath.Join(a.raiz, nombre)
	if err := os.Rename(filepath.Join(a.tmp, ref), destino); err != nil {
		return "", fmt.Errorf("no se pudo confirmar el documento: %w", err)
	}
	return destino, nil
}

func (a *AlmacenLocal) Descartar(ref string) error {
	if !nombreTemporal.MatchString(ref) {
		return ErrNombreInvalido
	}
	return ignorarInexistente(os.Remove(filepath.Join(a.tmp, ref)))
}

func (a *AlmacenLocal) Eliminar(nombre string) error {
	if !nombreDefinitivo.MatchString(nombre) {
		return ErrNombreInvalido
	}
	return ignorarInexistente(os.Remove(filepath.Join(a.raiz, nombre)))
}

func (a *AlmacenLocal) Abrir(nombre string) (io.ReadCloser, error) {
	if !nombreDefinitivo.MatchString(nombre) {
		return nil, ErrNombreInvalido
	}
	f, err := os.Open(filepath.Join(a.raiz, nombre))
	if errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrDocumentoNoEncontrado
	}
	return f, err
}

func ignorarInexistente(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
