package elevenlabs

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/taigrr/elevenlabs/v2/internal/transport"
)

// SongMetadata describes a generated song.
type SongMetadata struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
	Languages   []string `json:"languages"`
	IsExplicit  bool     `json:"is_explicit"`
}

// MusicDetailedMetadata is the JSON part of a detailed composition response.
type MusicDetailedMetadata struct {
	CompositionPlan *MusicCompositionPlan `json:"composition_plan,omitempty"`
	SongMetadata    *SongMetadata         `json:"song_metadata,omitempty"`
	WordsTimestamps []WordTimestamp       `json:"words_timestamps,omitempty"`
	WaveformVisual  []float64             `json:"waveform_visual,omitempty"`
	// Raw is the undecoded JSON part, including fields not modeled above.
	Raw json.RawMessage `json:"-"`
}

// MusicDetailedResponse is a generated song with its metadata, mirroring the
// MultipartResponse of the official SDK. The audio is streamed rather than
// buffered: read it from the response (which implements io.ReadCloser) and
// close it when done.
type MusicDetailedResponse struct {
	// Metadata is the decoded JSON part.
	Metadata MusicDetailedMetadata
	// Filename is the audio part's filename (default "generated_music.mp3").
	Filename string
	// ContentType is the audio part's content type, if reported.
	ContentType string
	// SongID is the song-id response header, if present.
	SongID string

	audio  io.Reader
	closer io.Closer
}

// Read reads the song audio.
func (r *MusicDetailedResponse) Read(p []byte) (int, error) { return r.audio.Read(p) }

// Close releases the underlying connection.
func (r *MusicDetailedResponse) Close() error { return r.closer.Close() }

const defaultMusicFilename = "generated_music.mp3"

// ComposeDetailed generates a song and returns its metadata together with a
// streaming reader over the audio. The caller must close the response.
func (s *MusicService) ComposeDetailed(ctx context.Context, req *MusicComposeDetailedRequest, opts ...Option) (*MusicDetailedResponse, error) {
	req = orEmpty(req)
	r := &request{method: http.MethodPost, path: "/v1/music/detailed", query: req, body: req, accept: acceptAny, opts: opts}
	resp, cancel, err := send(ctx, s.cfg, r)
	if err != nil {
		return nil, err
	}
	body := &transport.CancelOnClose{ReadCloser: resp.Body, Cancel: cancel}
	out, err := parseMusicMultipart(body, resp.Header)
	if err != nil {
		body.Close()
		return nil, fmt.Errorf("elevenlabs: parse detailed composition response: %w", err)
	}
	return out, nil
}

// parseMusicMultipart decodes the leading JSON part and positions the
// returned response at the start of the audio part.
func parseMusicMultipart(body io.ReadCloser, header http.Header) (*MusicDetailedResponse, error) {
	var src io.Reader = body
	boundary := ""
	if _, params, err := mime.ParseMediaType(header.Get(headerContentType)); err == nil {
		boundary = params["boundary"]
	}
	if boundary == "" {
		// Like the official SDK, fall back to the first line of the body.
		br := bufio.NewReader(body)
		line, err := br.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") || len(trimmed) <= 2 {
			return nil, errors.New("missing multipart boundary")
		}
		boundary = trimmed[2:]
		src = io.MultiReader(strings.NewReader(line), br)
	}
	mr := multipart.NewReader(src, boundary)
	out := &MusicDetailedResponse{Filename: defaultMusicFilename, SongID: header.Get("song-id"), closer: body}
	haveJSON := false
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			return nil, errors.New("could not find audio part")
		}
		if err != nil {
			return nil, err
		}
		mediaType, _, _ := mime.ParseMediaType(part.Header.Get(headerContentType))
		if mediaType == contentTypeJSON && !haveJSON {
			raw, err := io.ReadAll(part)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(raw, &out.Metadata); err != nil {
				return nil, fmt.Errorf("decode metadata: %w", err)
			}
			out.Metadata.Raw = json.RawMessage(raw)
			haveJSON = true
			continue
		}
		if !haveJSON {
			return nil, errors.New("audio part precedes JSON metadata")
		}
		if name := part.FileName(); name != "" {
			out.Filename = name
		}
		out.ContentType = part.Header.Get(headerContentType)
		out.audio = part
		return out, nil
	}
}
