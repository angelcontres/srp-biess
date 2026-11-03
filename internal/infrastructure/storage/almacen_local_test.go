package storage_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/storage"
)

func TestAlmacenLocal(t *testing.T) {
	raiz := t.TempDir()
	a, err := storage.NewAlmacenLocal(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(filepath.Join(raiz, "documentos")); err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("el directorio debía crearse con permisos 0700: %v %v", info, err)
	}

	ref, n, err := a.EscribirTemporal(strings.NewReader("%PDF-contenido"))
	if err != nil || n != 14 {
		t.Fatalf("EscribirTemporal: %d %v", n, err)
	}
	nombre := "7c9e6679-7425-40de-944b-e07fc1f90ae7.pdf"
	ruta, err := a.Confirmar(ref, nombre)
	if err != nil || ruta != filepath.Join(raiz, "documentos", nombre) {
		t.Fatalf("Confirmar: %s %v", ruta, err)
	}
	f, err := a.Abrir(nombre)
	if err != nil {
		t.Fatal(err)
	}
	contenido, _ := io.ReadAll(f)
	_ = f.Close()
	if string(contenido) != "%PDF-contenido" {
		t.Fatalf("contenido inesperado: %q", contenido)
	}
	if err := a.Eliminar(nombre); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Abrir(nombre); !errors.Is(err, domain.ErrDocumentoNoEncontrado) {
		t.Fatalf("tras eliminar debía ser ErrDocumentoNoEncontrado: %v", err)
	}
	if err := a.Eliminar(nombre); err != nil {
		t.Fatalf("eliminar dos veces no debía fallar: %v", err)
	}

	// Path traversal: solo nombres generados por el sistema.
	for _, malo := range []string{"../../etc/passwd", "a.pdf", "../" + nombre, ""} {
		if _, err := a.Abrir(malo); !errors.Is(err, storage.ErrNombreInvalido) {
			t.Errorf("Abrir(%q) debía rechazarse: %v", malo, err)
		}
		if err := a.Eliminar(malo); !errors.Is(err, storage.ErrNombreInvalido) {
			t.Errorf("Eliminar(%q) debía rechazarse: %v", malo, err)
		}
	}
	ref, _, _ = a.EscribirTemporal(strings.NewReader("x"))
	if _, err := a.Confirmar(ref, "../fuera.pdf"); !errors.Is(err, storage.ErrNombreInvalido) {
		t.Fatalf("Confirmar con nombre externo debía rechazarse: %v", err)
	}
	if err := a.Descartar(ref); err != nil {
		t.Fatal(err)
	}
	if entradas, _ := os.ReadDir(filepath.Join(raiz, "documentos", "tmp")); len(entradas) != 0 {
		t.Fatalf("no debían quedar temporales: %v", entradas)
	}
}
