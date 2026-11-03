package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"srp-biess/internal/domain"
	"srp-biess/internal/interfaces/http/handlers"
)

type paymentServiceStub struct {
	err error
}

func (s paymentServiceStub) ProcessPayment(context.Context, domain.CardPaymentRequest) (*domain.InvoiceResult, *domain.PaymentResult, error) {
	if s.err != nil {
		return nil, &domain.PaymentResult{Status: "RECHAZADO"}, s.err
	}
	return &domain.InvoiceResult{
			InvoiceNumber: "001-001-000004522",
			AccessKey:     "1234567890123456789012345678901234567890123456789",
			Status:        "EMITIDA",
			EmailQueued:   true,
		},
		&domain.PaymentResult{
			TransactionID: "TXN-BIESS-2026-1",
			Status:        "APROBADO",
			Franchise:     "VISA",
			LastFour:      "0366",
			Authorization: "AUTH-MOCK-1",
		}, nil
}

func TestPaymentHandler_ReturnsCreated(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("postulante_id", "550e8400-e29b-41d4-a716-446655440000")
		return c.Next()
	})
	app.Post("/postulaciones/:id/pagos", handlers.NewPaymentHandler(paymentServiceStub{}).Pay)

	body := bytes.NewBufferString(`{"numeroTarjeta":"4532015112830366","mesExpiracion":12,"anioExpiracion":2028,"cvv":"123","monto":25,"moneda":"USD"}`)
	req := httptest.NewRequest(http.MethodPost, "/postulaciones/550e8400-e29b-41d4-a716-446655440001/pagos", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d; want 201", resp.StatusCode)
	}
}

func TestPaymentHandler_Returns402WhenDeclined(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("postulante_id", "550e8400-e29b-41d4-a716-446655440000")
		return c.Next()
	})
	app.Post("/postulaciones/:id/pagos", handlers.NewPaymentHandler(paymentServiceStub{err: domain.ErrPaymentDeclined}).Pay)

	req := httptest.NewRequest(http.MethodPost, "/postulaciones/550e8400-e29b-41d4-a716-446655440001/pagos", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode != http.StatusPaymentRequired || result["codigo"] != "PAGO_DECLINADO" {
		t.Fatalf("unexpected response: status=%d body=%v", resp.StatusCode, result)
	}
}

func TestPaymentHandler_Returns401WithoutValidSession(t *testing.T) {
	app := fiber.New()
	app.Post("/postulaciones/:id/pagos", handlers.NewPaymentHandler(paymentServiceStub{}).Pay)

	req := httptest.NewRequest(http.MethodPost, "/postulaciones/550e8400-e29b-41d4-a716-446655440001/pagos", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", resp.StatusCode)
	}
}
