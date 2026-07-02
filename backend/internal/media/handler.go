package media

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
)

// Handler exposes media upload endpoints.
type Handler struct {
	client        *Client
	attachments   *Repository
	conversations *conversation.Repository
}

// NewHandler creates a media handler.
func NewHandler(client *Client, attachments *Repository, conversations *conversation.Repository) *Handler {
	return &Handler{client: client, attachments: attachments, conversations: conversations}
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

type downloadURLRequest struct {
	ObjectKey string `json:"object_key"`
}

// HandlePresignDownload returns a presigned GET URL for an owned attachment.
func (h *Handler) HandlePresignDownload(w http.ResponseWriter, r *http.Request) {
	if h.client == nil || h.attachments == nil {
		apierror.Write(w, r, http.StatusServiceUnavailable, "unavailable", "media storage not configured")
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req downloadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	att, err := h.attachments.GetAccessible(r.Context(), claims.UserID, req.ObjectKey, func(convID, userID uuid.UUID) (bool, error) {
		if h.conversations == nil {
			return false, nil
		}
		return h.conversations.IsMember(r.Context(), convID, userID)
	})
	if err != nil {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "attachment not accessible")
		return
	}

	downloadURL, err := h.client.PresignGetURL(r.Context(), att.ObjectKey)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to presign download")
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"download_url": downloadURL,
		"object_key":   att.ObjectKey,
		"mime_type":    att.MimeType,
		"expires_in":   900,
	})
}
