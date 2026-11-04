package elevenlabs

import (
	"context"
	"io"
	"iter"
	"net/http"
)

// HistoryService groups the /v1/history endpoints.
type HistoryService struct{ cfg *config }

// SortDirection is a sort order.
type SortDirection string

// SortDirection values.
const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

// HistorySource identifies the product that generated a history item.
type HistorySource string

// HistorySource values.
const (
	HistorySourceTTS             HistorySource = "TTS"
	HistorySourceSTS             HistorySource = "STS"
	HistorySourceProjects        HistorySource = "Projects"
	HistorySourcePD              HistorySource = "PD"
	HistorySourceAN              HistorySource = "AN"
	HistorySourceDubbing         HistorySource = "Dubbing"
	HistorySourcePlayAPI         HistorySource = "PlayAPI"
	HistorySourceConvAI          HistorySource = "ConvAI"
	HistorySourceVoiceGeneration HistorySource = "VoiceGeneration"
	HistorySourceInVPC           HistorySource = "InVPC"
	HistorySourceFlows           HistorySource = "Flows"
	HistorySourceVoicePreview    HistorySource = "VoicePreview"
)

// HistoryListRequest holds the optional parameters of HistoryService.List.
type HistoryListRequest struct {
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
	SortDirection SortDirection `query:"sort_direction"`
	// Search filters by text.
	Search string `query:"search"`
	// Source filters by product (TTS, STS or Flows).
	Source HistorySource `query:"source"`
}

// List returns one page of generated audio history. req may be nil.
func (s *HistoryService) List(ctx context.Context, req *HistoryListRequest, opts ...RequestOption) (*GetSpeechHistoryResponse, error) {
	return doValue[GetSpeechHistoryResponse](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/history", query: req, opts: opts})
}

// ListAll iterates over every history item matching req, fetching pages on demand.
func (s *HistoryService) ListAll(ctx context.Context, req *HistoryListRequest, opts ...RequestOption) iter.Seq2[SpeechHistoryItem, error] {
	var params HistoryListRequest
	if req != nil {
		params = *req
	}
	return paginate(ctx, params.StartAfterHistoryItemID, func(ctx context.Context, cursor string) (page[SpeechHistoryItem], error) {
		params.StartAfterHistoryItemID = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[SpeechHistoryItem]{}, err
		}
		next := ""
		if resp.LastHistoryItemID != nil {
			next = *resp.LastHistoryItemID
		}
		return page[SpeechHistoryItem]{items: resp.History, next: next, hasMore: resp.HasMore}, nil
	})
}

// Get returns a single history item.
func (s *HistoryService) Get(ctx context.Context, historyItemID string, opts ...RequestOption) (*SpeechHistoryItem, error) {
	return doValue[SpeechHistoryItem](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/history/%s", historyItemID), opts: opts})
}

// Delete deletes a history item.
func (s *HistoryService) Delete(ctx context.Context, historyItemID string, opts ...RequestOption) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/history/%s", historyItemID), opts: opts})
}

// GetAudio streams the audio of a history item. The caller must close the returned reader.
func (s *HistoryService) GetAudio(ctx context.Context, historyItemID string, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/history/%s/audio", historyItemID), accept: acceptAny, opts: opts})
}

// DownloadHistoryRequest is the body of HistoryService.Download.
type DownloadHistoryRequest struct {
	// HistoryItemIDs lists the items to download. Required.
	HistoryItemIDs []string `json:"history_item_ids"`
	// OutputFormat is "wav" or the default "mp3"; only applied when converting.
	OutputFormat string `json:"output_format,omitempty"`
}

// Download streams one history item's audio, or a zip archive when multiple
// IDs are given. The caller must close the returned reader.
func (s *HistoryService) Download(ctx context.Context, req *DownloadHistoryRequest, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/history/download", body: orEmpty(req), accept: acceptAny, opts: opts})
}

// CharacterAlignment holds per-character timing information.
type CharacterAlignment struct {
	Characters                 []string  `json:"characters"`
	CharacterStartTimesSeconds []float64 `json:"character_start_times_seconds"`
	CharacterEndTimesSeconds   []float64 `json:"character_end_times_seconds"`
}

// HistoryAlignments holds the raw and normalized alignment of a history item.
type HistoryAlignments struct {
	Alignment           CharacterAlignment `json:"alignment"`
	NormalizedAlignment CharacterAlignment `json:"normalized_alignment"`
}

// DialogueInput is one line of a text-to-dialogue history item.
type DialogueInput struct {
	Text      string `json:"text"`
	VoiceID   string `json:"voice_id"`
	VoiceName string `json:"voice_name"`
}

// SpeechHistoryItem is a single generated audio history item.
type SpeechHistoryItem struct {
	HistoryItemID            string             `json:"history_item_id"`
	RequestID                *string            `json:"request_id,omitempty"`
	VoiceID                  *string            `json:"voice_id,omitempty"`
	ModelID                  *string            `json:"model_id,omitempty"`
	VoiceName                *string            `json:"voice_name,omitempty"`
	VoiceCategory            *VoiceCategory     `json:"voice_category,omitempty"`
	Text                     *string            `json:"text,omitempty"`
	DateUnix                 int64              `json:"date_unix"`
	CharacterCountChangeFrom int64              `json:"character_count_change_from"`
	CharacterCountChangeTo   int64              `json:"character_count_change_to"`
	ContentType              string             `json:"content_type"`
	State                    any                `json:"state,omitempty"`
	Settings                 map[string]any     `json:"settings,omitempty"`
	ShareLinkID              *string            `json:"share_link_id,omitempty"`
	Source                   *HistorySource     `json:"source,omitempty"`
	Alignments               *HistoryAlignments `json:"alignments,omitempty"`
	Dialogue                 []DialogueInput    `json:"dialogue,omitempty"`
	OutputFormat             *string            `json:"output_format,omitempty"`
}

// GetSpeechHistoryResponse is one page returned by HistoryService.List.
type GetSpeechHistoryResponse struct {
	History           []SpeechHistoryItem `json:"history"`
	LastHistoryItemID *string             `json:"last_history_item_id,omitempty"`
	HasMore           bool                `json:"has_more"`
	ScannedUntil      *int64              `json:"scanned_until,omitempty"`
}
