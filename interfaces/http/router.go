package http

import (
	"net/http"

	"payment_ms/application"
)

func NewRouter(
	initiateUC *application.InitiatePaymentUseCase,
	webhookUC *application.HandleWebhookUseCase,
) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /payments/initiate", NewInitiatePaymentHandler(initiateUC))
	mux.Handle("POST /payments/webhook", NewWebhookHandler(webhookUC))

	return mux
}
