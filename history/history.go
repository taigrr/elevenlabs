package history

import (
	"context"
	"io"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

// Service groups the /v1/history endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// Source identifies the product that generated a history item.
type Source string

// Source values.
const (
	SourceTTS             Source = "TTS"
	SourceSTS             Source = "STS"
	SourceProjects        Source = "Projects"
	SourcePD              Source = "PD"
	SourceAN              Source = "AN"
	SourceDubbing         Source = "Dubbing"
	SourcePlayAPI         Source = "PlayAPI"
	SourceConvAI          Source = "ConvAI"
	SourceVoiceGeneration Source = "VoiceGeneration"
	SourceInVPC           Source = "InVPC"
	SourceFlows           Source = "Flows"
	SourceVoicePreview    Source = "VoicePreview"
)

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	// PageSize is the number of items per page (max 1000, default 100).
	PageSize *int `query:"page_size"`
	// StartAfterHistoryItemID is the pagination cursor.
	StartAfterHistoryItemID string `query:"start_after_history_item_id"`
	// VoiceID filters by voice.
	VoiceID string `query:"voice_id"`
	// ModelID filters by model.
	ModelID string `query:"model_id"`
	// DateBeforeUnix returns items created before this unix time.
	DateBeforeUnix *int64 `query:"date_before_unix"`
	// DateAfterUnix returns items created at or after this unix time.
	DateAfterUnix *int64 `query:"date_after_unix"`
	// SortDirection orders the results.
	SortDirection types.SortDirection `query:"sort_direction"`
	// Search filters by text.
	Search string `query:"search"`
	// Source filters by product (TTS, STS or Flows).
	Source Source `query:"source"`
}

// List returns one page of generated audio history. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*ListResponse, error) {
	return transport.DoValue[ListResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/history", Query: req, Opts: opts})
}

// ListAll iterates over every history item matching req, fetching pages on demand.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Item, error] {
	var params ListRequest
	if req != nil {
		params = *req
	}
	return stream.Paginate(ctx, params.StartAfterHistoryItemID, func(ctx context.Context, cursor string) (stream.Page[Item], error) {
		params.StartAfterHistoryItemID = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Item]{}, err
		}
		next := ""
		if resp.LastHistoryItemID != nil {
			next = *resp.LastHistoryItemID
		}
		return stream.Page[Item]{Items: resp.History, Next: next, HasMore: resp.HasMore}, nil
	})
}

// Get returns a single history item.
func (s *Service) Get(ctx context.Context, historyItemID string, opts ...option.RequestOption) (*Item, error) {
	return transport.DoValue[Item](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/history/%s", historyItemID), Opts: opts})
}

// Delete deletes a history item.
func (s *Service) Delete(ctx context.Context, historyItemID string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/history/%s", historyItemID), Opts: opts})
}

// GetAudio streams the audio of a history item. The caller must close the returned reader.
func (s *Service) GetAudio(ctx context.Context, historyItemID string, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/history/%s/audio", historyItemID), Accept: transport.AcceptAny, Opts: opts})
}

// DownloadRequest is the body of Service.Download.
type DownloadRequest struct {
	// HistoryItemIDs lists the items to download. Required.
	HistoryItemIDs []string `json:"history_item_ids"`
	// OutputFormat is "wav" or the default "mp3"; only applied when converting.
	OutputFormat string `json:"output_format,omitempty"`
}

// Download streams one history item's audio, or a zip archive when multiple
// IDs are given. The caller must close the returned reader.
func (s *Service) Download(ctx context.Context, req *DownloadRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/history/download", Body: transport.OrEmpty(req), Accept: transport.AcceptAny, Opts: opts})
}

// Alignments holds the raw and normalized alignment of a history item.
type Alignments struct {
	Alignment           types.CharacterAlignment `json:"alignment"`
	NormalizedAlignment types.CharacterAlignment `json:"normalized_alignment"`
}

// DialogueInput is one line of a text-to-dialogue history item.
type DialogueInput struct {
	Text      string `json:"text"`
	VoiceID   string `json:"voice_id"`
	VoiceName string `json:"voice_name"`
}

// Item is a single generated audio history item.
type Item struct {
	HistoryItemID            string           `json:"history_item_id"`
	RequestID                *string          `json:"request_id,omitempty"`
	VoiceID                  *string          `json:"voice_id,omitempty"`
	ModelID                  *string          `json:"model_id,omitempty"`
	VoiceName                *string          `json:"voice_name,omitempty"`
	VoiceCategory            *voices.Category `json:"voice_category,omitempty"`
	Text                     *string          `json:"text,omitempty"`
	DateUnix                 int64            `json:"date_unix"`
	CharacterCountChangeFrom int64            `json:"character_count_change_from"`
	CharacterCountChangeTo   int64            `json:"character_count_change_to"`
	ContentType              string           `json:"content_type"`
	State                    any              `json:"state,omitempty"`
	Settings                 map[string]any   `json:"settings,omitempty"`
	ShareLinkID              *string          `json:"share_link_id,omitempty"`
	Source                   *Source          `json:"source,omitempty"`
	Alignments               *Alignments      `json:"alignments,omitempty"`
	Dialogue                 []DialogueInput  `json:"dialogue,omitempty"`
	OutputFormat             *string          `json:"output_format,omitempty"`
}

// ListResponse is one page returned by Service.List.
type ListResponse struct {
	History           []Item  `json:"history"`
	LastHistoryItemID *string `json:"last_history_item_id,omitempty"`
	HasMore           bool    `json:"has_more"`
	ScannedUntil      *int64  `json:"scanned_until,omitempty"`
}
