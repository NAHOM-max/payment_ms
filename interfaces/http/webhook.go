package http

import (
	"errors"
	"net/http"

	"payment_ms/application"
	"payment_ms/domain"
)

var validStatuses = map[domain.PaymentStatus]bool{
	domain.StatusSuccessful: true,
	domain.StatusFailed:     true,
	domain.StatusCanceled:   true,
}

type webhookRequest struct {
	PaymentID string `json:"payment_id" validate:"required"`
	Status    string `json:"status"     validate:"required"`
}

type WebhookHandler struct {
	uc *application.HandleWebhookUseCase
}

func NewWebhookHandler(uc *application.HandleWebhookUseCase) *WebhookHandler {
	return &WebhookHandler{uc: uc}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req webhookRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := domain.PaymentStatus(req.Status)
	if !validStatuses[status] {
		writeError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	_, err := h.uc.Execute(r.Context(), application.HandleWebhookInput{
		PaymentID: req.PaymentID,
		Status:    status,
	})
	if err != nil {
		switch {
		case errors.Is(err, application.ErrPaymentNotFound):
			writeError(w, http.StatusNotFound, "payment not found")
		case errors.Is(err, application.ErrInvalidStatusTransition):
			writeError(w, http.StatusUnprocessableEntity, "invalid status transition")
		default:
			writeError(w, http.StatusInternalServerError, "failed to process webhook")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
