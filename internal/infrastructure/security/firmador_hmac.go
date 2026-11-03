package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"srp-biess/internal/domain"
)

// FirmadorHMAC firma las URL de descarga con HMAC-SHA256(secreto, docId|vence) (SDD 9.7).
// La descarga exige además el JWT del dueño (doble validación, #40).
type FirmadorHMAC struct {
	secreto []byte
}

var _ domain.FirmadorDescarga = (*FirmadorHMAC)(nil)

func NewFirmadorHMAC(secreto []byte) *FirmadorHMAC {
	return &FirmadorHMAC{secreto: secreto}
}

func (f *FirmadorHMAC) firma(docID string, vence int64) []byte {
	mac := hmac.New(sha256.New, f.secreto)
	mac.Write([]byte(docID + "|" + strconv.FormatInt(vence, 10)))
	return mac.Sum(nil)
}

func (f *FirmadorHMAC) Firmar(docID string, vence time.Time) string {
	return hex.EncodeToString(f.firma(docID, vence.Unix()))
}

func (f *FirmadorHMAC) Verificar(docID string, vence int64, firma string, ahora time.Time) error {
	recibida, err := hex.DecodeString(firma)
	// hmac.Equal compara en tiempo constante (evita ataques de tiempo).
	if err != nil || !hmac.Equal(recibida, f.firma(docID, vence)) {
		return domain.ErrFirmaDescargaInvalida
	}
	if ahora.Unix() > vence {
		return domain.ErrFirmaDescargaInvalida
	}
	return nil
}
