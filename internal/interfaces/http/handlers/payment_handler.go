package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"srp-biess/internal/domain"
	"srp-biess/internal/interfaces/http/dto"
)

type PaymentHandler struct {
	service domain.PaymentService
}

func NewPaymentHandler(service domain.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

type paymentRequest struct {
	CardNumber  string  `json:"numeroTarjeta"`
	HolderName  string  `json:"nombreTitular"`
	ExpiryMonth int     `json:"mesExpiracion"`
	ExpiryYear  int     `json:"anioExpiracion"`
	CVV         string  `json:"cvv"`
	Amount      float64 `json:"monto"`
	Currency    string  `json:"moneda"`
	Description string  `json:"descripcion"`
}

func (h *PaymentHandler) Pay(c *fiber.Ctx) error {
	var body paymentRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"codigo": "VALIDATION_ERROR", "mensaje": "Cuerpo de petición inválido"})
	}
	postulanteID, _ := c.Locals("postulante_id").(string)
	if _, err := uuid.Parse(postulanteID); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.Failure("Sesión inválida o expirada"))
	}
	req := domain.CardPaymentRequest{
		PostulacionID: c.Params("id"),
		PostulanteID:  postulanteID,
		CardNumber:    body.CardNumber,
		HolderName:    body.HolderName,
		ExpiryMonth:   body.ExpiryMonth,
		ExpiryYear:    body.ExpiryYear,
		CVV:           body.CVV,
		Amount:        body.Amount,
		Currency:      body.Currency,
		Description:   body.Description,
	}
	invoice, payment, err := h.service.ProcessPayment(c.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCard):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"codigo": "TARJETA_INVALIDA", "mensaje": err.Error()})
		case errors.Is(err, domain.ErrPaymentDeclined):
			return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"codigo": "PAGO_DECLINADO", "mensaje": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"codigo": "PAGO_ERROR", "mensaje": "No fue posible procesar el pago"})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"transaccionId": payment.TransactionID,
		"estado":        payment.Status,
		"monto":         req.Amount,
		"moneda":        req.Currency,
		"franquicia":    payment.Franchise,
		"ultimosCuatro": payment.LastFour,
		"autorizacion":  payment.Authorization,
		"fecha":         time.Now().Format(time.RFC3339),
		"factura": fiber.Map{
			"numeroFactura": invoice.InvoiceNumber,
			"claveAcceso":   invoice.AccessKey,
			"estado":        invoice.Status,
			"enviadaCorreo": invoice.EmailQueued,
		},
	})
}
