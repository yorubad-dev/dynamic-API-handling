package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	// setup a logger using slog
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info("Hello Terminal 👋", "user", os.Getenv("USER"))

	http.HandleFunc("/health", HealthCheck(logger))
	http.HandleFunc("/dynamic-hook", HandleDynamicAPI(logger))

	go log.Fatal(http.ListenAndServe(":3000", nil))
}

// this will be used to identify the event type
//
// we care about just the event head
type eventIdentfier struct {
	Event string `json:"event"`
}

type paymentPending struct {
	Event string `json:"event"`
	Data  struct {
		ID               int       `json:"id"`
		Domain           string    `json:"domain"`
		Amount           int       `json:"amount"`
		Currency         string    `json:"currency"`
		DueDate          any       `json:"due_date"`
		HasInvoice       bool      `json:"has_invoice"`
		InvoiceNumber    any       `json:"invoice_number"`
		Description      string    `json:"description"`
		PdfURL           any       `json:"pdf_url"`
		LineItems        []any     `json:"line_items"`
		Tax              []any     `json:"tax"`
		RequestCode      string    `json:"request_code"`
		Status           string    `json:"status"`
		Paid             bool      `json:"paid"`
		PaidAt           any       `json:"paid_at"`
		Metadata         any       `json:"metadata"`
		Notifications    []any     `json:"notifications"`
		OfflineReference string    `json:"offline_reference"`
		Customer         int       `json:"customer"`
		CreatedAt        time.Time `json:"created_at"`
	} `json:"data"`
}

type paymentSuccessful struct {
	Event string `json:"event"`
	Data  struct {
		ID            int       `json:"id"`
		Domain        string    `json:"domain"`
		Amount        int       `json:"amount"`
		Currency      string    `json:"currency"`
		DueDate       any       `json:"due_date"`
		HasInvoice    bool      `json:"has_invoice"`
		InvoiceNumber any       `json:"invoice_number"`
		Description   string    `json:"description"`
		PdfURL        any       `json:"pdf_url"`
		LineItems     []any     `json:"line_items"`
		Tax           []any     `json:"tax"`
		RequestCode   string    `json:"request_code"`
		Status        string    `json:"status"`
		Paid          bool      `json:"paid"`
		PaidAt        time.Time `json:"paid_at"`
		Metadata      any       `json:"metadata"`
		Notifications []struct {
			SentAt  time.Time `json:"sent_at"`
			Channel string    `json:"channel"`
		} `json:"notifications"`
		OfflineReference string    `json:"offline_reference"`
		Customer         int       `json:"customer"`
		CreatedAt        time.Time `json:"created_at"`
	} `json:"data"`
}

func HealthCheck(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]string{"data": "Hello from localhost:3000"}); err != nil {
			l.Error("error encoding data to send as response", "error context", err)
			return
		}
	}
}

func HandleDynamicAPI(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Info("This API is connected", "user", os.Getenv("USER"))
		var (
			eventIdentfier    eventIdentfier
			jsonData          json.RawMessage
			paymentPending    paymentPending
			paymentSuccessful paymentSuccessful
		)

		if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
			l.Error("error decoding json response", "error context", err)
			return
		}

		if err := json.Unmarshal(jsonData, &eventIdentfier); err != nil {
			l.Error("error unmarshalling json data message", "error context", err)
			return
		}

		switch eventIdentfier.Event {
		case "paymentrequest.pending":
			l.Info("payment pending hook event", "response event title", eventIdentfier.Event)

			if err := json.Unmarshal(jsonData, &paymentPending); err != nil {
				l.Error("error marshalling pending payment data", "error context", err)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{"event type": paymentPending.Event, "amount": paymentPending.Data.Amount}); err != nil {
				l.Error("error encoding data to send as response", "error context", err)
				return
			}

			l.Info("pending response data unmarshalled successfully", "pending ID", paymentPending.Data.ID, "pending amount", paymentPending.Data.Amount)
		case "paymentrequest.success":
			l.Info("payment successful hook event", "response event title", eventIdentfier.Event)

			if err := json.Unmarshal(jsonData, &paymentSuccessful); err != nil {
				l.Error("error marshalling successful payment data", "error context", err)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{"event type": paymentSuccessful.Event, "description": paymentSuccessful.Data.Description}); err != nil {
				return
			}

			l.Info("success response data unmarshalled successfully", "success ID", paymentSuccessful.Data.ID, "success amount", paymentSuccessful.Data.Amount)
		default:
			l.Info("no event type found", "response event title", eventIdentfier.Event)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
