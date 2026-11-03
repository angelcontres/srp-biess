package domain_test

import (
	"testing"
	"time"

	"srp-biess/internal/domain"
)

func TestValidLuhn(t *testing.T) {
	if !domain.ValidLuhn("4532015112830366") {
		t.Fatal("expected valid Visa test number")
	}
	if domain.ValidLuhn("4532015112830367") {
		t.Fatal("expected invalid Luhn number")
	}
}

func TestValidateCard(t *testing.T) {
	req := domain.CardPaymentRequest{
		CardNumber:  "4532015112830366",
		ExpiryMonth: 12,
		ExpiryYear:  2028,
		CVV:         "123",
		Amount:      25,
	}
	if err := domain.ValidateCard(req, time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("expected valid card: %v", err)
	}
	req.CVV = "12"
	if err := domain.ValidateCard(req, time.Now()); err != domain.ErrInvalidCard {
		t.Fatalf("expected invalid card, got %v", err)
	}
}
