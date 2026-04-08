package http

import (
	"net/http"

	"payment_ms/application"
)

type initiateRequest struct {
	CustomerID string  `json:"customer_id" validate:"required"`
	OrderID    string  `json:"order_id"    validate:"required"`
	WorkflowID string  `json:"workflow_id" validate:"required"`
	Amount     float64 `json:"amount"      validate:"required,gt=0"`
}

type InitiatePaymentHandler struct {
	uc *application.InitiatePaymentUseCase
}

func NewInitiatePaymentHandler(uc *application.InitiatePaymentUseCase) *InitiatePaymentHandler {
	return &InitiatePaymentHandler{uc: uc}
}

func (h *InitiatePaymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req initiateRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	paymentID, err := h.uc.Execute(r.Context(), application.InitiatePaymentInput{
		CustomerID: req.CustomerID,
		OrderID:    req.OrderID,
		WorkflowID: req.WorkflowID,
		Amount:     req.Amount,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to initiate payment")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"payment_id": paymentID})
}
