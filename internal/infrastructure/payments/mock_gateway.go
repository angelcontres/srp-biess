package payments

import (
	"context"
	"fmt"
	"strings"
	"time"

	"srp-biess/internal/domain"
)

type MockPaymentGateway struct {
	now func() time.Time
}

func NewMockPaymentGateway() *MockPaymentGateway {
	return &MockPaymentGateway{now: time.Now}
}

func (g *MockPaymentGateway) AuthorizeAndCapture(_ context.Context, req domain.CardPaymentRequest) (*domain.PaymentResult, error) {
	result := &domain.PaymentResult{
		TransactionID: fmt.Sprintf("TXN-BIESS-%d", g.now().UnixNano()),
		Status:        "APROBADO",
		Franchise:     domain.CardFranchise(req.CardNumber),
		LastFour:      req.CardNumber[len(req.CardNumber)-4:],
		Authorization: fmt.Sprintf("AUTH-MOCK-%d", g.now().UnixNano()%100000),
	}
	if strings.HasSuffix(req.CardNumber, "0005") || strings.Contains(strings.ToUpper(req.HolderName), "DECLINADA") {
		result.Status = "RECHAZADO"
		result.Reason = "Fondos insuficientes"
		return result, domain.ErrPaymentDeclined
	}
	return result, nil
}
