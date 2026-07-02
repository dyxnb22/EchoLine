package ads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
)

// Campaign is an ad campaign for a channel.
type Campaign struct {
	ID        uuid.UUID
	ChannelID uuid.UUID
	Title     string
	Status    string
	CreatedAt time.Time
}

// Repository persists ad campaigns.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an ads repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new campaign.
func (r *Repository) Create(ctx context.Context, channelID uuid.UUID, title string) (*Campaign, error) {
	const q = `
		INSERT INTO ad_campaigns (id, channel_id, title, status, created_at)
		VALUES (gen_random_uuid(), $1, $2, 'draft', $3)
		RETURNING id, channel_id, title, status, created_at
	`
	row := r.pool.QueryRow(ctx, q, channelID, title, time.Now().UTC())
	var c Campaign
	if err := row.Scan(&c.ID, &c.ChannelID, &c.Title, &c.Status, &c.CreatedAt); err != nil {
		return nil, fmt.Errorf("create campaign: %w", err)
	}
	return &c, nil
}

// List returns all campaigns for a channel.
func (r *Repository) List(ctx context.Context, channelID uuid.UUID) ([]Campaign, error) {
	const q = `
		SELECT id, channel_id, title, status, created_at
		FROM ad_campaigns
		WHERE channel_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, channelID)
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}
	defer rows.Close()

	var out []Campaign
	for rows.Next() {
		var c Campaign
		if err := rows.Scan(&c.ID, &c.ChannelID, &c.Title, &c.Status, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan campaign: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// UpdateStatus changes campaign status.
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Campaign, error) {
	const q = `
		UPDATE ad_campaigns SET status = $2 WHERE id = $1
		RETURNING id, channel_id, title, status, created_at
	`
	row := r.pool.QueryRow(ctx, q, id, status)
	var c Campaign
	if err := row.Scan(&c.ID, &c.ChannelID, &c.Title, &c.Status, &c.CreatedAt); err != nil {
		return nil, fmt.Errorf("update campaign: %w", err)
	}
	return &c, nil
}

// EnsureCampaignChannel verifies a campaign belongs to the given channel.
func (r *Repository) EnsureCampaignChannel(ctx context.Context, campaignID, channelID uuid.UUID) error {
	const q = `SELECT channel_id FROM ad_campaigns WHERE id = $1`
	var got uuid.UUID
	if err := r.pool.QueryRow(ctx, q, campaignID).Scan(&got); err != nil {
		return fmt.Errorf("campaign not found: %w", err)
	}
	if got != channelID {
		return fmt.Errorf("campaign channel mismatch")
	}
	return nil
}

// RecordImpression records an ad impression with frequency cap enforcement.
func (r *Repository) RecordImpression(ctx context.Context, campaignID, userID uuid.UUID) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const capQ = `SELECT frequency_cap FROM ad_campaigns WHERE id = $1 FOR UPDATE`
	var cap int
	if err := tx.QueryRow(ctx, capQ, campaignID).Scan(&cap); err != nil {
		return false, fmt.Errorf("get frequency cap: %w", err)
	}
	if cap <= 0 {
		cap = 3
	}

	const countQ = `
		SELECT COUNT(*) FROM ad_impressions
		WHERE campaign_id = $1 AND user_id = $2 AND impression_day = CURRENT_DATE
	`
	var count int
	if err := tx.QueryRow(ctx, countQ, campaignID, userID).Scan(&count); err != nil {
		return false, fmt.Errorf("count impressions: %w", err)
	}
	if count >= cap {
		return false, nil
	}

	const insQ = `
		INSERT INTO ad_impressions (id, campaign_id, user_id, created_at, impression_day)
		VALUES (gen_random_uuid(), $1, $2, NOW(), CURRENT_DATE)
	`
	if _, err := tx.Exec(ctx, insQ, campaignID, userID); err != nil {
		return false, fmt.Errorf("record impression: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit impression: %w", err)
	}
	return true, nil
}

// Handler exposes ad campaign REST endpoints.
type Handler struct {
	repo          *Repository
	owners        *conversation.OwnerChecker
	conversations *conversation.Repository
}

// NewHandler creates an ads handler.
func NewHandler(repo *Repository, owners *conversation.OwnerChecker, conversations *conversation.Repository) *Handler {
	return &Handler{repo: repo, owners: owners, conversations: conversations}
}

// HandleCreate creates a campaign.
// POST /api/channels/{channel_id}/campaigns
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	channelID, err := parseChannelID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid channel_id")
		return
	}

	owner, err := h.owners.IsChannelOwner(r.Context(), channelID, claims.UserID)
	if err != nil || !owner {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "channel owner required")
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "title is required")
		return
	}

	campaign, err := h.repo.Create(r.Context(), channelID, req.Title)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to create campaign")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, campaignPayload(campaign))
}

func (h *Handler) requireChannelMember(w http.ResponseWriter, r *http.Request, channelID, userID uuid.UUID) bool {
	member, err := h.conversations.IsMember(r.Context(), channelID, userID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "channel membership required")
		return false
	}
	return true
}

// HandleList lists campaigns for a channel.
// GET /api/channels/{channel_id}/campaigns
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	channelID, err := parseChannelID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid channel_id")
		return
	}
	if !h.requireChannelMember(w, r, channelID, claims.UserID) {
		return
	}

	campaigns, err := h.repo.List(r.Context(), channelID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list campaigns")
		return
	}

	items := make([]map[string]any, 0, len(campaigns))
	for i := range campaigns {
		items = append(items, campaignPayload(&campaigns[i]))
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"campaigns": items})
}

// HandleRecordImpression records an ad impression with frequency cap.
// POST /api/channels/{channel_id}/campaigns/{campaign_id}/impressions
func (h *Handler) HandleRecordImpression(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	channelID, err := parseChannelID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid channel_id")
		return
	}
	if !h.requireChannelMember(w, r, channelID, claims.UserID) {
		return
	}

	campaignID, err := parseCampaignID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid campaign_id")
		return
	}
	if err := h.repo.EnsureCampaignChannel(r.Context(), campaignID, channelID); err != nil {
		apierror.Write(w, r, http.StatusNotFound, "not_found", "campaign not found for channel")
		return
	}

	recorded, err := h.repo.RecordImpression(r.Context(), campaignID, claims.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to record impression")
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"recorded": recorded,
		"status":   map[bool]string{true: "recorded", false: "frequency_capped"}[recorded],
	})
}

func parseCampaignID(path string) (uuid.UUID, error) {
	const marker = "/campaigns/"
	idx := strings.Index(path, marker)
	if idx < 0 {
		return uuid.Nil, fmt.Errorf("invalid path")
	}
	rest := strings.TrimPrefix(path[idx:], marker)
	parts := strings.SplitN(rest, "/", 2)
	return uuid.Parse(parts[0])
}

func parseChannelID(path string) (uuid.UUID, error) {
	const prefix = "/api/channels/"
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, fmt.Errorf("invalid path")
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.SplitN(rest, "/", 2)
	return uuid.Parse(parts[0])
}

func campaignPayload(c *Campaign) map[string]any {
	return map[string]any{
		"id":         c.ID,
		"channel_id": c.ChannelID,
		"title":      c.Title,
		"status":     c.Status,
		"created_at": c.CreatedAt.UTC().Format(time.RFC3339),
	}
}
