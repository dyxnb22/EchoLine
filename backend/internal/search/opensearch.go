package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// OpenSearchResult is a single search hit from OpenSearch.
type OpenSearchResult struct {
	MessageID      uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Body           string
	Seq            int64
	CreatedAt      time.Time
}

// OpenSearchClient wraps an OpenSearch HTTP endpoint.
type OpenSearchClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOpenSearchClient creates an OpenSearch client targeting baseURL.
// If baseURL is empty, the client is inoperative (callers should fall back to PG).
func NewOpenSearchClient(baseURL string) *OpenSearchClient {
	return &OpenSearchClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Enabled returns true when an OpenSearch URL is configured.
func (c *OpenSearchClient) Enabled() bool {
	return c.baseURL != ""
}

// Search performs a simple match query against the messages index.
func (c *OpenSearchClient) Search(ctx context.Context, userID uuid.UUID, q string, limit int) ([]OpenSearchResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("opensearch not configured")
	}

	body := map[string]any{
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{
				"must": map[string]any{
					"match": map[string]any{
						"body": q,
					},
				},
				"filter": map[string]any{
					"term": map[string]string{
						"member_ids": userID.String(),
					},
				},
			},
		},
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("opensearch marshal: %w", err)
	}

	url := fmt.Sprintf("%s/messages/_search", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("opensearch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("opensearch do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("opensearch %d: %s", resp.StatusCode, data)
	}

	var osResp struct {
		Hits struct {
			Hits []struct {
				Source struct {
					MessageID      string    `json:"message_id"`
					ConversationID string    `json:"conversation_id"`
					SenderID       string    `json:"sender_id"`
					Body           string    `json:"body"`
					Seq            int64     `json:"seq"`
					CreatedAt      time.Time `json:"created_at"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&osResp); err != nil {
		return nil, fmt.Errorf("opensearch decode: %w", err)
	}

	out := make([]OpenSearchResult, 0, len(osResp.Hits.Hits))
	for _, h := range osResp.Hits.Hits {
		msgID, _ := uuid.Parse(h.Source.MessageID)
		convID, _ := uuid.Parse(h.Source.ConversationID)
		sndID, _ := uuid.Parse(h.Source.SenderID)
		out = append(out, OpenSearchResult{
			MessageID:      msgID,
			ConversationID: convID,
			SenderID:       sndID,
			Body:           h.Source.Body,
			Seq:            h.Source.Seq,
			CreatedAt:      h.Source.CreatedAt,
		})
	}
	return out, nil
}
