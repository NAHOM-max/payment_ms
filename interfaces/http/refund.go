package http

import (
	"errors"
	"net/http"

	"payment_ms/application"
)

type refundRequest struct {
	PaymentID string `json:"payment_id" validate:"required"`
}

type RefundHandler struct {
	uc *application.RequestRefundUseCase
}

func NewRefundHandler(uc *application.RequestRefundUseCase) *RefundHandler {
	return &RefundHandler{uc: uc}
}

func (h *RefundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req refundRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := h.uc.Execute(r.Context(), req.PaymentID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrPaymentNotFound):
			writeError(w, http.StatusNotFound, "payment not found")
		case errors.Is(err, application.ErrInvalidStatusTransition):
			writeError(w, http.StatusBadRequest, "refund only allowed for successful payments")
		default:
			writeError(w, http.StatusInternalServerError, "failed to request refund")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"payment_id": out.PaymentID,
		"status":     string(out.Status),
	})
}
