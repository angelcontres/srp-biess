package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidCard        = errors.New("los datos de la tarjeta no son válidos")
	ErrPaymentDeclined    = errors.New("el pago fue declinado")
	ErrPaymentNotFound    = errors.New("pago no encontrado")
	ErrPaymentPersistence = errors.New("no fue posible registrar el pago")
)

type CardPaymentRequest struct {
	PostulacionID string
	PostulanteID  string
	CardNumber    string
	HolderName    string
	ExpiryMonth   int
	ExpiryYear    int
	CVV           string
	Amount        float64
	Currency      string
	Description   string
}

type PaymentResult struct {
	TransactionID string
	Status        string
	Franchise     string
	LastFour      string
	Authorization string
	Reason        string
}

type InvoiceResult struct {
	InvoiceNumber string
	AccessKey     string
	Status        string
	Subtotal      float64
	Tax           float64
	Total         float64
	EmailQueued   bool
}

type PaymentGateway interface {
	AuthorizeAndCapture(ctx context.Context, req CardPaymentRequest) (*PaymentResult, error)
}

type InvoicingService interface {
	GenerateAndSendInvoice(ctx context.Context, req CardPaymentRequest, payment *PaymentResult) (*InvoiceResult, error)
}

type PaymentRepository interface {
	CreatePayment(ctx context.Context, req CardPaymentRequest, payment *PaymentResult) (string, error)
	CreateInvoice(ctx context.Context, req CardPaymentRequest, transactionID string, invoice *InvoiceResult) error
	MarkPostulationPaid(ctx context.Context, postulacionID string) error
	EnqueuePaymentNotification(ctx context.Context, req CardPaymentRequest, payment *PaymentResult, invoice *InvoiceResult) error
	GetInvoice(ctx context.Context, postulacionID string) (*InvoiceResult, error)
}

type PaymentService interface {
	ProcessPayment(ctx context.Context, req CardPaymentRequest) (*InvoiceResult, *PaymentResult, error)
}

func ValidateCard(req CardPaymentRequest, now time.Time) error {
	if !ValidLuhn(req.CardNumber) || req.ExpiryMonth < 1 || req.ExpiryMonth > 12 ||
		req.ExpiryYear < now.Year() || (req.ExpiryYear == now.Year() && req.ExpiryMonth < int(now.Month())) ||
		len(req.CVV) < 3 || len(req.CVV) > 4 || req.Amount <= 0 {
		return ErrInvalidCard
	}
	for _, digit := range req.CVV {
		if digit < '0' || digit > '9' {
			return ErrInvalidCard
		}
	}
	return nil
}

func ValidLuhn(number string) bool {
	if len(number) < 13 || len(number) > 19 {
		return false
	}
	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		if number[i] < '0' || number[i] > '9' {
			return false
		}
		digit := int(number[i] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}

func CardFranchise(number string) string {
	switch {
	case len(number) > 0 && number[0] == '4':
		return "VISA"
	case len(number) > 0 && number[0] == '5':
		return "MASTERCARD"
	case len(number) > 0 && number[0] == '3':
		return "AMEX_DINERS"
	default:
		return "OTRA"
	}
}
