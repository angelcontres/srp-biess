package application_test

import (
	"context"
	"errors"
	"testing"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/payments"
)

func validPaymentRequest() domain.CardPaymentRequest {
	return domain.CardPaymentRequest{
		PostulacionID: "postulacion-1",
		PostulanteID:  "postulante-1",
		CardNumber:    "4532015112830366",
		HolderName:    "JUAN PEREZ",
		ExpiryMonth:   12,
		ExpiryYear:    2028,
		CVV:           "123",
		Amount:        25,
		Currency:      "USD",
	}
}

func TestPaymentService_ApprovedPaymentEmitsInvoice(t *testing.T) {
	repo := payments.NewMemoryPaymentRepository()
	service := application.NewPaymentService(
		nil,
		payments.NewMockPaymentGateway(),
		payments.NewMockInvoicingService(),
		repo,
	)

	invoice, payment, err := service.ProcessPayment(context.Background(), validPaymentRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Status != "APROBADO" || len(invoice.AccessKey) != 49 || !invoice.EmailQueued {
		t.Fatalf("unexpected result: payment=%+v invoice=%+v", payment, invoice)
	}
}

func TestPaymentService_DeclinedPaymentDoesNotEmitInvoice(t *testing.T) {
	req := validPaymentRequest()
	req.HolderName = "TARJETA DECLINADA"
	repo := payments.NewMemoryPaymentRepository()
	service := application.NewPaymentService(
		nil,
		payments.NewMockPaymentGateway(),
		payments.NewMockInvoicingService(),
		repo,
	)

	invoice, payment, err := service.ProcessPayment(context.Background(), req)
	if !errors.Is(err, domain.ErrPaymentDeclined) || invoice != nil || payment == nil {
		t.Fatalf("expected declined payment without invoice: invoice=%v payment=%+v err=%v", invoice, payment, err)
	}
	if _, err := repo.GetInvoice(context.Background(), req.PostulacionID); !errors.Is(err, domain.ErrPaymentNotFound) {
		t.Fatalf("declined payment persisted invoice: %v", err)
	}
}
