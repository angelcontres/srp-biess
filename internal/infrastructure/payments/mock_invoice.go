package payments

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"srp-biess/internal/domain"
)

var invoiceSequence uint64 = 4521

type MockInvoicingService struct {
	now func() time.Time
}

func NewMockInvoicingService() *MockInvoicingService {
	return &MockInvoicingService{now: time.Now}
}

func (s *MockInvoicingService) GenerateAndSendInvoice(_ context.Context, req domain.CardPaymentRequest, _ *domain.PaymentResult) (*domain.InvoiceResult, error) {
	sequence := atomic.AddUint64(&invoiceSequence, 1)
	date := s.now()
	subtotal := req.Amount / 1.12
	tax := req.Amount - subtotal
	accessKey := fmt.Sprintf("%s%09d", date.Format("20060102150405"), sequence)
	accessKey = accessKey + strings.Repeat("0", 49-len(accessKey))
	return &domain.InvoiceResult{
		InvoiceNumber: fmt.Sprintf("001-001-%09d", sequence),
		AccessKey:     accessKey[:49],
		Status:        "EMITIDA",
		Subtotal:      subtotal,
		Tax:           tax,
		Total:         req.Amount,
		EmailQueued:   true,
	}, nil
}
