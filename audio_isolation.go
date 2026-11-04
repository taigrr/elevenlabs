package elevenlabs

import (
	"context"
	"io"
	"iter"
	"net/http"
	"strconv"
)

// AudioFileFormat describes the encoding of an uploaded audio file.
type AudioFileFormat string

// AudioFileFormat values.
const (
	// AudioFileFormatPCMS16LE16 is 16-bit PCM, 16kHz, mono, little-endian.
	// It lowers latency compared to encoded formats.
	AudioFileFormatPCMS16LE16 AudioFileFormat = "pcm_s16le_16"
	// AudioFileFormatOther is any other encoded audio format.
	AudioFileFormatOther AudioFileFormat = "other"
)

// AudioIsolationService groups the /v1/audio-isolation endpoints.
type AudioIsolationService struct{ cfg *config }

// AudioIsolationRequest is the request of AudioIsolationService.Convert.
type AudioIsolationRequest struct {
	// Audio is the file to remove background noise from. Required.
	Audio File `form:"audio"`
	// FileFormat describes the input encoding (default AudioFileFormatOther).
	FileFormat AudioFileFormat `form:"file_format"`
	// PreviewB64 is an optional base64-encoded preview image.
	PreviewB64 string `form:"preview_b64"`
}

// AudioIsolationStreamRequest is the request of AudioIsolationService.Stream.
type AudioIsolationStreamRequest struct {
	// Audio is the file to remove background noise from. Required.
	Audio File `form:"audio"`
	// FileFormat describes the input encoding (default AudioFileFormatOther).
	FileFormat AudioFileFormat `form:"file_format"`
}

// Convert removes background noise from audio and returns the isolated
// audio. The caller must close the returned reader.
func (s *AudioIsolationService) Convert(ctx context.Context, req *AudioIsolationRequest, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/audio-isolation", form: orEmpty(req), accept: acceptAny, opts: opts})
}

// Stream removes background noise from audio and streams the isolated audio
// as it is produced. The caller must close the returned reader.
func (s *AudioIsolationService) Stream(ctx context.Context, req *AudioIsolationStreamRequest, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/audio-isolation/stream", form: orEmpty(req), accept: acceptAny, opts: opts})
}

// AudioIsolationListRequest holds the optional parameters of AudioIsolationService.List.
type AudioIsolationListRequest struct {
	// PageSize is the number of items per page.
	PageSize *int `query:"page_size"`
	// Page is the page number to return.
	Page *int `query:"page"`
	// Search filters items by title.
	Search string `query:"search"`
}

// AudioIsolationHistoryItem is one previously processed audio isolation.
type AudioIsolationHistoryItem struct {
	ID                    string   `json:"id"`
	Title                 *string  `json:"title,omitempty"`
	CreatedAtUnix         int64    `json:"created_at_unix"`
	Format                string   `json:"format"`
	DurationSeconds       *float64 `json:"duration_seconds,omitempty"`
	DownloadURL           *string  `json:"download_url,omitempty"`
	IconURL               *string  `json:"icon_url,omitempty"`
	SourceVideoURL        *string  `json:"source_video_url,omitempty"`
	SupportsVideo         bool     `json:"supports_video"`
	Processing            bool     `json:"processing"`
	VideoProcessingFailed bool     `json:"video_processing_failed"`
	PreviewB64            *string  `json:"preview_b64,omitempty"`
}

// AudioIsolationHistory is one page of audio isolation history.
type AudioIsolationHistory struct {
	Items   []AudioIsolationHistoryItem `json:"items"`
	HasMore bool                        `json:"has_more"`
}

// List returns one page of audio isolation history. req may be nil.
func (s *AudioIsolationService) List(ctx context.Context, req *AudioIsolationListRequest, opts ...RequestOption) (*AudioIsolationHistory, error) {
	return doValue[AudioIsolationHistory](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/audio-isolation/history", query: req, opts: opts})
}

// ListAll iterates over audio isolation history items, fetching pages on
// demand. The API only paginates searches (1-based Page, starting at
// req.Page or 1); without req.Search only the first page is returned.
func (s *AudioIsolationService) ListAll(ctx context.Context, req *AudioIsolationListRequest, opts ...RequestOption) iter.Seq2[AudioIsolationHistoryItem, error] {
	var params AudioIsolationListRequest
	if req != nil {
		params = *req
	}
	first := 1
	if params.Page != nil {
		first = *params.Page
	}
	return paginate(ctx, strconv.Itoa(first), func(ctx context.Context, cursor string) (page[AudioIsolationHistoryItem], error) {
		n, err := strconv.Atoi(cursor)
		if err != nil {
			return page[AudioIsolationHistoryItem]{}, err
		}
		if params.Search != "" {
			params.Page = &n
		}
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[AudioIsolationHistoryItem]{}, err
		}
		return page[AudioIsolationHistoryItem]{items: resp.Items, next: strconv.Itoa(n + 1), hasMore: resp.HasMore && params.Search != ""}, nil
	})
}

// Delete deletes an audio isolation history item.
func (s *AudioIsolationService) Delete(ctx context.Context, historyItemID string, opts ...RequestOption) error {
	return doJSON(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/audio-isolation/history/%s", historyItemID), opts: opts}, nil)
}
