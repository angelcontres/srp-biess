package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"srp-biess/internal/domain"
)

type PostgresPaymentRepository struct {
	q Querier
}

func NewPostgresPaymentRepository(q Querier) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{q: q}
}

func (r *PostgresPaymentRepository) CreatePayment(ctx context.Context, req domain.CardPaymentRequest, payment *domain.PaymentResult) (string, error) {
	var transactionID string
	err := r.getQuerier(ctx).QueryRow(ctx, `
		INSERT INTO transaccion_pago
			(postulacion_id, postulante_id, codigo_transaccion, monto, moneda, franquicia,
			 ultimos_cuatro, estado, gateway_provider, autorizacion_codigo)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8::estado_pago, 'MOCK', $9)
		RETURNING id`,
		req.PostulacionID, req.PostulanteID, payment.TransactionID, req.Amount, req.Currency,
		payment.Franchise, payment.LastFour, payment.Status, payment.Authorization,
	).Scan(&transactionID)
	if err != nil {
		return "", fmt.Errorf("insertando transacción de pago: %w", err)
	}
	return transactionID, nil
}

func (r *PostgresPaymentRepository) CreateInvoice(ctx context.Context, req domain.CardPaymentRequest, transactionID string, invoice *domain.InvoiceResult) error {
	_, err := r.getQuerier(ctx).Exec(ctx, `
		INSERT INTO factura
			(transaccion_pago_id, postulacion_id, numero_factura, clave_acceso_sri, subtotal, iva, total, enviada_correo)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)`,
		transactionID, req.PostulacionID, invoice.InvoiceNumber, invoice.AccessKey,
		invoice.Subtotal, invoice.Tax, invoice.Total, invoice.EmailQueued)
	if err != nil {
		return fmt.Errorf("insertando factura: %w", err)
	}
	return nil
}

func (r *PostgresPaymentRepository) MarkPostulationPaid(ctx context.Context, postulacionID string) error {
	tag, err := r.getQuerier(ctx).Exec(ctx, `
		UPDATE postulacion SET estado_arancel = 'PAGADO', updated_at = now()
		WHERE id = $1::uuid`,
		postulacionID)
	if err != nil {
		return fmt.Errorf("actualizando estado de arancel: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("postulación no encontrada para actualizar estado de arancel")
	}
	return nil
}

func (r *PostgresPaymentRepository) EnqueuePaymentNotification(ctx context.Context, req domain.CardPaymentRequest, payment *domain.PaymentResult, invoice *domain.InvoiceResult) error {
	payload, err := json.Marshal(map[string]interface{}{
		"transaccionId": payment.TransactionID,
		"numeroFactura": invoice.InvoiceNumber,
		"claveAcceso":   invoice.AccessKey,
		"monto":         req.Amount,
	})
	if err != nil {
		return fmt.Errorf("serializando notificación de factura: %w", err)
	}

	tag, err := r.getQuerier(ctx).Exec(ctx, `
		INSERT INTO notificacion
			(postulacion_id, postulante_id, tipo, destinatario, asunto, plantilla, payload)
		SELECT $1::uuid, p.id, 'FACTURA_PAGO', p.correo,
			'Comprobante de pago y factura electrónica', 'FACTURA_PAGO', $3::jsonb
		FROM postulante p WHERE p.id = $2::uuid`,
		req.PostulacionID, req.PostulanteID, string(payload))
	if err != nil {
		return fmt.Errorf("encolando notificación de factura: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("no se pudo encolar la notificación de factura")
	}
	return nil
}

func (r *PostgresPaymentRepository) getQuerier(ctx context.Context) Querier {
	return ExtractQuerier(ctx, r.q)
}

func (r *PostgresPaymentRepository) GetInvoice(ctx context.Context, postulacionID string) (*domain.InvoiceResult, error) {
	var invoice domain.InvoiceResult
	var createdAt time.Time
	err := r.getQuerier(ctx).QueryRow(ctx, `
		SELECT numero_factura, clave_acceso_sri, estado, subtotal, iva, total, enviada_correo, created_at
		FROM factura WHERE postulacion_id = $1::uuid ORDER BY created_at DESC LIMIT 1`,
		postulacionID,
	).Scan(&invoice.InvoiceNumber, &invoice.AccessKey, &invoice.Status, &invoice.Subtotal,
		&invoice.Tax, &invoice.Total, &invoice.EmailQueued, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return &invoice, nil
}
