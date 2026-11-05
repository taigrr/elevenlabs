package audioisolation

import (
	"context"
	"io"
	"iter"
	"net/http"
	"strconv"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/audio-isolation endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// ConvertRequest is the request of Service.Convert.
type ConvertRequest struct {
	// Audio is the file to remove background noise from. Required.
	Audio types.File `form:"audio"`
	// FileFormat describes the input encoding (default types.AudioFileFormatOther).
	FileFormat types.AudioFileFormat `form:"file_format"`
	// PreviewB64 is an optional base64-encoded preview image.
	PreviewB64 string `form:"preview_b64"`
}

// StreamRequest is the request of Service.Stream.
type StreamRequest struct {
	// Audio is the file to remove background noise from. Required.
	Audio types.File `form:"audio"`
	// FileFormat describes the input encoding (default types.AudioFileFormatOther).
	FileFormat types.AudioFileFormat `form:"file_format"`
}

// Convert removes background noise from audio and returns the isolated
// audio. The caller must close the returned reader.
func (s *Service) Convert(ctx context.Context, req *ConvertRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/audio-isolation", Form: transport.OrEmpty(req), Accept: transport.AcceptAny, Opts: opts})
}

// Stream removes background noise from audio and streams the isolated audio
// as it is produced. The caller must close the returned reader.
func (s *Service) Stream(ctx context.Context, req *StreamRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/audio-isolation/stream", Form: transport.OrEmpty(req), Accept: transport.AcceptAny, Opts: opts})
}

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	// PageSize is the number of items per page.
	PageSize *int `query:"page_size"`
	// Page is the page number to return.
	Page *int `query:"page"`
	// Search filters items by title.
	Search string `query:"search"`
}

// HistoryItem is one previously processed audio isolation.
type HistoryItem struct {
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

// History is one page of audio isolation history.
type History struct {
	Items   []HistoryItem `json:"items"`
	HasMore bool          `json:"has_more"`
}

// List returns one page of audio isolation history. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*History, error) {
	return transport.DoValue[History](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/audio-isolation/history", Query: req, Opts: opts})
}

// ListAll iterates over audio isolation history items, fetching pages on
// demand. The API only paginates searches (1-based Page, starting at
// req.Page or 1); without req.Search only the first page is returned.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[HistoryItem, error] {
	var params ListRequest
	if req != nil {
		params = *req
	}
	first := 1
	if params.Page != nil {
		first = *params.Page
	}
	return stream.Paginate(ctx, strconv.Itoa(first), func(ctx context.Context, cursor string) (stream.Page[HistoryItem], error) {
		n, err := strconv.Atoi(cursor)
		if err != nil {
			return stream.Page[HistoryItem]{}, err
		}
		if params.Search != "" {
			params.Page = &n
		}
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[HistoryItem]{}, err
		}
		return stream.Page[HistoryItem]{Items: resp.Items, Next: strconv.Itoa(n + 1), HasMore: resp.HasMore && params.Search != ""}, nil
	})
}

// Delete deletes an audio isolation history item.
func (s *Service) Delete(ctx context.Context, historyItemID string, opts ...option.RequestOption) error {
	return transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/audio-isolation/history/%s", historyItemID), Opts: opts}, nil)
}
