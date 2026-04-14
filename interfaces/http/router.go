package http

import (
	"net/http"

	"payment_ms/application"
)

func NewRouter(
	initiateUC *application.InitiatePaymentUseCase,
	webhookUC *application.HandleWebhookUseCase,
	refundUC *application.RequestRefundUseCase,
) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /payments/initiate", NewInitiatePaymentHandler(initiateUC))
	mux.Handle("POST /payments/webhook", NewWebhookHandler(webhookUC))
	mux.Handle("POST /payments/refund", NewRefundHandler(refundUC))

	return mux
}
