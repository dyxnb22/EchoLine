package media

import (
	"encoding/json"
	"net/http"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// Handler exposes media upload endpoints.
type Handler struct {
	client     *Client
	attachments *Repository
}

// NewHandler creates a media handler.
func NewHandler(client *Client, attachments *Repository) *Handler {
	return &Handler{client: client, attachments: attachments}
}

type uploadURLRequest struct {
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	Checksum  string `json:"checksum"`
}

// HandlePresignUpload returns a presigned PUT URL for direct upload.
func (h *Handler) HandlePresignUpload(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		apierror.Write(w, r, http.StatusServiceUnavailable, "unavailable", "media storage not configured")
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req uploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	uploadURL, objectKey, err := h.client.PresignPutURL(r.Context(), claims.UserID, req.MimeType)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to presign upload")
		return
	}

	if h.attachments != nil {
		if _, err := h.attachments.RegisterPending(r.Context(), claims.UserID, objectKey, req.MimeType, req.SizeBytes, req.Checksum); err != nil {
			apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to register attachment")
			return
		}
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"upload_url":  uploadURL,
		"object_key":  objectKey,
		"bucket":      h.client.Bucket(),
		"expires_in":  900,
	})
}
