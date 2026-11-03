package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"srp-biess/internal/domain"
)

type PaymentServiceImpl struct {
	txManager domain.TxManager
	gateway   domain.PaymentGateway
	invoice   domain.InvoicingService
	repo      domain.PaymentRepository
	now       func() time.Time
}

func NewPaymentService(txManager domain.TxManager, gateway domain.PaymentGateway, invoice domain.InvoicingService, repo domain.PaymentRepository) *PaymentServiceImpl {
	if txManager == nil {
		txManager = domain.NoopTxManager{}
	}
	return &PaymentServiceImpl{txManager: txManager, gateway: gateway, invoice: invoice, repo: repo, now: time.Now}
}

func (s *PaymentServiceImpl) ProcessPayment(ctx context.Context, req domain.CardPaymentRequest) (*domain.InvoiceResult, *domain.PaymentResult, error) {
	if err := domain.ValidateCard(req, s.now()); err != nil {
		return nil, nil, err
	}
	payment, err := s.gateway.AuthorizeAndCapture(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentDeclined) {
			return nil, payment, err
		}
		return nil, nil, fmt.Errorf("error en la pasarela de pago: %w", err)
	}
	invoice, err := s.invoice.GenerateAndSendInvoice(ctx, req, payment)
	if err != nil {
		return nil, payment, fmt.Errorf("error generando factura: %w", err)
	}
	var persistenceErr error
	if err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		transactionID, err := s.repo.CreatePayment(txCtx, req, payment)
		if err != nil {
			return err
		}
		if err := s.repo.CreateInvoice(txCtx, req, transactionID, invoice); err != nil {
			return err
		}
		if err := s.repo.MarkPostulationPaid(txCtx, req.PostulacionID); err != nil {
			return err
		}
		return s.repo.EnqueuePaymentNotification(txCtx, req, payment, invoice)
	}); err != nil {
		persistenceErr = err
	}
	if persistenceErr != nil {
		return nil, payment, fmt.Errorf("%w: %v", domain.ErrPaymentPersistence, persistenceErr)
	}
	return invoice, payment, nil
}
