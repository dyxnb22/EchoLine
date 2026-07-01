package apierror

import (
	"context"
	"encoding/json"
	"net/http"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// WithRequestID stores a request ID in context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext returns the request ID if present.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// ErrorBody is the standard REST error response.
type ErrorBody struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

// Write writes a JSON API error.
func Write(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	var body ErrorBody
	body.Error.Code = code
	body.Error.Message = message
	body.Error.RequestID = RequestIDFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// WriteJSON writes a success JSON payload.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
