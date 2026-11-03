package payments

import (
	"context"
	"sync"

	"srp-biess/internal/domain"
)

type MemoryPaymentRepository struct {
	mu       sync.RWMutex
	invoices map[string]*domain.InvoiceResult
}

func NewMemoryPaymentRepository() *MemoryPaymentRepository {
	return &MemoryPaymentRepository{invoices: make(map[string]*domain.InvoiceResult)}
}

func (r *MemoryPaymentRepository) CreatePayment(_ context.Context, _ domain.CardPaymentRequest, payment *domain.PaymentResult) (string, error) {
	return payment.TransactionID, nil
}

func (r *MemoryPaymentRepository) CreateInvoice(_ context.Context, req domain.CardPaymentRequest, _ string, invoice *domain.InvoiceResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invoices[req.PostulacionID] = invoice
	return nil
}

func (r *MemoryPaymentRepository) MarkPostulationPaid(context.Context, string) error { return nil }

func (r *MemoryPaymentRepository) EnqueuePaymentNotification(context.Context, domain.CardPaymentRequest, *domain.PaymentResult, *domain.InvoiceResult) error {
	return nil
}

func (r *MemoryPaymentRepository) GetInvoice(_ context.Context, postulacionID string) (*domain.InvoiceResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	invoice, ok := r.invoices[postulacionID]
	if !ok {
		return nil, domain.ErrPaymentNotFound
	}
	return invoice, nil
}
