package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/client/types"
)

// newTestClient creates a Client pointed at the given test server.
func newTestClient(ts *httptest.Server) Client {
	return New("test-api-key").WithEndpoint(ts.URL)
}

// assertAPIKey checks that the request carries the expected API key header.
func assertAPIKey(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("xi-api-key"); got != "test-api-key" {
		t.Errorf("xi-api-key = %q, want %q", got, "test-api-key")
	}
}

func TestNew(t *testing.T) {
	c := New("my-key")
	if c.apiKey != "my-key" {
		t.Fatalf("apiKey = %q, want %q", c.apiKey, "my-key")
	}
	if c.endpoint != apiEndpoint {
		t.Fatalf("endpoint = %q, want %q", c.endpoint, apiEndpoint)
	}
	if c.httpClient == nil {
		t.Fatal("httpClient is nil")
	}
}

func TestWithEndpoint(t *testing.T) {
	c := New("k").WithEndpoint("http://custom")
	if c.endpoint != "http://custom" {
		t.Fatalf("endpoint = %q", c.endpoint)
	}
}

func TestWithAPIKey(t *testing.T) {
	c := New("old").WithAPIKey("new")
	if c.apiKey != "new" {
		t.Fatalf("apiKey = %q", c.apiKey)
	}
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c := New("k").WithHTTPClient(custom)
	if c.httpClient != custom {
		t.Fatal("httpClient not set")
	}
}

// --- TTS tests ---

func TestTTS(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAPIKey(t, r)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/text-to-speech/") {
			t.Errorf("path = %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		var body types.TTS
		json.NewDecoder(r.Body).Decode(&body)
		if body.Text != "hello" {
			t.Errorf("text = %q", body.Text)
		}
		w.WriteHeader(200)
		w.Write([]byte("audio-bytes"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data, err := c.TTS(context.Background(), "hello", "voice1", "model1", types.SynthesisOptions{Stability: 0.5, SimilarityBoost: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "audio-bytes" {
		t.Errorf("data = %q", string(data))
	}
}

func TestTTSUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.TTS(context.Background(), "hello", "voice1", "", types.SynthesisOptions{})
	if err != ErrUnauthorized {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestTTSStream(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/stream") {
			t.Errorf("path = %s, want stream suffix", r.URL.Path)
		}
		w.WriteHeader(200)
		w.Write([]byte("streamed"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	var buf strings.Builder
	err := c.TTSStream(context.Background(), &buf, "text", "v1", types.SynthesisOptions{Stability: 0.5, SimilarityBoost: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "streamed" {
		t.Errorf("got = %q", buf.String())
	}
}

func TestTTSWithOptionalParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body types.TTS
		json.NewDecoder(r.Body).Decode(&body)
		if body.PreviousText != "prev" {
			t.Errorf("previous_text = %q", body.PreviousText)
		}
		if body.NextText != "next" {
			t.Errorf("next_text = %q", body.NextText)
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.TTS(context.Background(), "hello", "v1", "m1",
		types.SynthesisOptions{Stability: 0.5, SimilarityBoost: 0.5},
		WithPreviousText("prev"), WithNextText("next"))
	if err != nil {
		t.Fatal(err)
	}
}

// --- Voice tests ---

func TestGetVoices(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAPIKey(t, r)
		if r.URL.Path != "/v1/voices" {
			t.Errorf("path = %s", r.URL.Path)
		}
		resp := types.GetVoicesResponseModel{
			Voices: []types.VoiceResponseModel{
				{VoiceID: "id1", Name: "Alice"},
				{VoiceID: "id2", Name: "Bob"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	voices, err := c.GetVoices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(voices) != 2 {
		t.Fatalf("len = %d", len(voices))
	}
	if voices[0].Name != "Alice" {
		t.Errorf("name = %s", voices[0].Name)
	}
}

func TestGetVoiceIDs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.GetVoicesResponseModel{
			Voices: []types.VoiceResponseModel{
				{VoiceID: "id1"},
				{VoiceID: "id2"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	ids, err := c.GetVoiceIDs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != "id1" || ids[1] != "id2" {
		t.Errorf("ids = %v", ids)
	}
}

func TestGetVoice(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAPIKey(t, r)
		if r.URL.Path != "/v1/voices/v123" {
			t.Errorf("path = %s", r.URL.Path)
		}
		resp := types.VoiceResponseModel{VoiceID: "v123", Name: "Test"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	voice, err := c.GetVoice(context.Background(), "v123")
	if err != nil {
		t.Fatal(err)
	}
	if voice.VoiceID != "v123" {
		t.Errorf("voiceID = %s", voice.VoiceID)
	}
}

func TestGetVoiceUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.GetVoice(context.Background(), "v1")
	if err != ErrUnauthorized {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestDeleteVoice(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAPIKey(t, r)
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	err := c.DeleteVoice(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetVoiceSettings(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.SynthesisOptions{Stability: 0.5, SimilarityBoost: 0.8}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	settings, err := c.GetVoiceSettings(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
	if settings.Stability != 0.5 {
		t.Errorf("stability = %f", settings.Stability)
	}
}

func TestEditVoiceSettings(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(types.SynthesisOptions{})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	err := c.EditVoiceSettings(context.Background(), "v1", types.SynthesisOptions{Stability: 0.5})
	if err != nil {
		t.Fatal(err)
	}
}

// --- User tests ---

func TestGetUserInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAPIKey(t, r)
		resp := types.UserResponseModel{
			Subscription: types.Subscription{Tier: "pro"},
			IsNewUser:    false,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	info, err := c.GetUserInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Subscription.Tier != "pro" {
		t.Errorf("tier = %s", info.Subscription.Tier)
	}
}

func TestGetSubscriptionInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.UserResponseModel{
			Subscription: types.Subscription{Tier: "starter", CharacterLimit: 10000},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	sub, err := c.GetSubscriptionInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sub.CharacterLimit != 10000 {
		t.Errorf("limit = %d", sub.CharacterLimit)
	}
}

// --- History tests ---

func TestGetHistoryItemList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAPIKey(t, r)
		resp := types.GetHistoryResponse{
			History: []types.HistoryItemList{
				{HistoryItemID: "h1", VoiceID: "v1", Text: "hello"},
				{HistoryItemID: "h2", VoiceID: "v2", Text: "world"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.GetHistoryItemList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len = %d", len(items))
	}
}

func TestGetHistoryIDs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.GetHistoryResponse{
			History: []types.HistoryItemList{
				{HistoryItemID: "h1", VoiceID: "v1"},
				{HistoryItemID: "h2", VoiceID: "v2"},
				{HistoryItemID: "h3", VoiceID: "v1"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)

	// No filter
	ids, err := c.GetHistoryIDs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 {
		t.Errorf("unfiltered len = %d", len(ids))
	}
}

func TestGetHistoryIDsFiltered(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.GetHistoryResponse{
			History: []types.HistoryItemList{
				{HistoryItemID: "h1", VoiceID: "v1"},
				{HistoryItemID: "h2", VoiceID: "v2"},
				{HistoryItemID: "h3", VoiceID: "v1"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	ids, err := c.GetHistoryIDs(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Errorf("filtered len = %d, want 2", len(ids))
	}
}

func TestHistoryDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	ok, err := c.HistoryDelete(context.Background(), "h1")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected true")
	}
}

func TestHistoryDownloadAudio(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("audio-data"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data, err := c.HistoryDownloadAudio(context.Background(), "h1")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

// --- Samples tests ---

func TestDeleteVoiceSample(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	ok, err := c.DeleteVoiceSample(context.Background(), "v1", "s1")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected true")
	}
}

func TestDownloadVoiceSample(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("sample-audio"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data, err := c.DownloadVoiceSample(context.Background(), "v1", "s1")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

// --- Sound generation tests ---

func TestSoundGeneration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAPIKey(t, r)
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		var body types.SoundGeneration
		json.NewDecoder(r.Body).Decode(&body)
		if body.Text != "thunder" {
			t.Errorf("text = %q", body.Text)
		}
		w.WriteHeader(200)
		w.Write([]byte("sound-bytes"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data, err := c.SoundGeneration(context.Background(), "thunder", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "sound-bytes" {
		t.Errorf("data = %q", string(data))
	}
}

func TestSoundGenerationWriter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("streamed-sound"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	var buf strings.Builder
	err := c.SoundGenerationWriter(context.Background(), &buf, "rain", 5.0, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "streamed-sound" {
		t.Errorf("got = %q", buf.String())
	}
}

// --- STT tests ---

func TestConvertSpeechToTextFromReader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAPIKey(t, r)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("Content-Type = %s", r.Header.Get("Content-Type"))
		}
		resp := types.SpeechToTextResponse{
			Text:         "hello world",
			LanguageCode: "en",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	resp, err := c.ConvertSpeechToTextFromReader(
		context.Background(),
		strings.NewReader("fake-audio"),
		"test.wav",
		types.SpeechToTextRequest{ModelID: types.SpeechToTextModelScribeV1},
	)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != "hello world" {
		t.Errorf("text = %q", resp.Text)
	}
}

func TestConvertSpeechToTextUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.ConvertSpeechToTextFromReader(
		context.Background(),
		strings.NewReader("fake"),
		"test.wav",
		types.SpeechToTextRequest{ModelID: types.SpeechToTextModelScribeV1},
	)
	if err != ErrUnauthorized {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

// --- Types tests ---

func TestSynthesisOptionsClamp(t *testing.T) {
	tests := []struct {
		name string
		in   types.SynthesisOptions
		want types.SynthesisOptions
	}{
		{
			name: "valid values unchanged",
			in:   types.SynthesisOptions{Stability: 0.5, SimilarityBoost: 0.8, Style: 0.3},
			want: types.SynthesisOptions{Stability: 0.5, SimilarityBoost: 0.8, Style: 0.3},
		},
		{
			name: "out of range clamped",
			in:   types.SynthesisOptions{Stability: 2.0, SimilarityBoost: -1.0, Style: 5.0},
			want: types.SynthesisOptions{Stability: 0.75, SimilarityBoost: 0.75, Style: 0.0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.in.Clamp()
			if tt.in.Stability != tt.want.Stability {
				t.Errorf("Stability = %f, want %f", tt.in.Stability, tt.want.Stability)
			}
			if tt.in.SimilarityBoost != tt.want.SimilarityBoost {
				t.Errorf("SimilarityBoost = %f, want %f", tt.in.SimilarityBoost, tt.want.SimilarityBoost)
			}
			if tt.in.Style != tt.want.Style {
				t.Errorf("Style = %f, want %f", tt.in.Style, tt.want.Style)
			}
		})
	}
}

func TestValidationErrorString(t *testing.T) {
	ve := types.ValidationError{Msg: "bad input", Type_: "value_error"}
	s := ve.Error()
	if !strings.Contains(s, "bad input") {
		t.Errorf("error = %q", s)
	}
}

func TestParamErrorString(t *testing.T) {
	pe := types.ParamError{}
	pe.Detail.Status = "error"
	pe.Detail.Message = "invalid param"
	s := pe.Error()
	if !strings.Contains(s, "invalid param") {
		t.Errorf("error = %q", s)
	}
}

// --- TTSWriter test ---

func TestTTSWriter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("written-audio"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	var buf strings.Builder
	err := c.TTSWriter(context.Background(), &buf, "hello", "m1", "v1", types.SynthesisOptions{Stability: 0.5, SimilarityBoost: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "written-audio" {
		t.Errorf("got = %q", buf.String())
	}
}

// --- DownloadVoiceSampleWriter test ---

func TestDownloadVoiceSampleWriter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("sample-stream"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	var buf strings.Builder
	err := c.DownloadVoiceSampleWriter(context.Background(), &buf, "v1", "s1")
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "sample-stream" {
		t.Errorf("got = %q", buf.String())
	}
}

// --- HistoryDownloadAudioWriter test ---

func TestHistoryDownloadAudioWriter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("audio-stream"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	var buf strings.Builder
	err := c.HistoryDownloadAudioWriter(context.Background(), &buf, "h1")
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "audio-stream" {
		t.Errorf("got = %q", buf.String())
	}
}

// --- HistoryDownloadZip tests ---

func TestHistoryDownloadZip(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		var body types.HistoryPost
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.HistoryItemIds) < 2 {
			t.Errorf("expected at least 2 ids, got %d", len(body.HistoryItemIds))
		}
		w.WriteHeader(200)
		w.Write([]byte("zip-data"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data, err := c.HistoryDownloadZip(context.Background(), "h1", "h2")
	if err != nil {
		t.Fatal(err)
	}
	_ = data
}

func TestHistoryDownloadZipWriter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("zip-stream"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	var buf strings.Builder
	err := c.HistoryDownloadZipWriter(context.Background(), &buf, "h1", "h2")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "zip-stream") {
		t.Errorf("got = %q", buf.String())
	}
}

// Verify CreateVoice sends multipart with correct field names
func TestCreateVoiceFieldNames(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAPIKey(t, r)
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		name := r.FormValue("name")
		if name != "TestVoice" {
			t.Errorf("name = %q", name)
		}
		labels := r.FormValue("labels")
		if labels != "english, male" {
			t.Errorf("labels = %q, want %q", labels, "english, male")
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	err := c.CreateVoice(context.Background(), "TestVoice", "A test voice", []string{"english", "male"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

// Verify handling of 422 validation errors
func TestTTS422(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(types.ValidationError{Msg: "invalid", Type_: "value_error"})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.TTS(context.Background(), "hello", "v1", "", types.SynthesisOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error = %q", err.Error())
	}
}

// Test that STT properly sends diarize and other fields
func TestSTTFields(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("model_id") != "scribe_v1" {
			t.Errorf("model_id = %q", r.FormValue("model_id"))
		}
		if r.FormValue("diarize") != "true" {
			t.Errorf("diarize = %q", r.FormValue("diarize"))
		}
		if r.FormValue("language_code") != "en" {
			t.Errorf("language_code = %q", r.FormValue("language_code"))
		}
		resp := types.SpeechToTextResponse{Text: "ok"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.ConvertSpeechToTextFromReader(
		context.Background(),
		strings.NewReader("audio"),
		"test.wav",
		types.SpeechToTextRequest{
			ModelID:      types.SpeechToTextModelScribeV1,
			LanguageCode: "en",
			Diarize:      true,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

// Test that io.Writer variants correctly proxy
func TestDownloadVoiceSampleWriterProxy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = io.WriteString(w, "proxy-test")
	}))
	defer ts.Close()

	c := newTestClient(ts)
	var buf strings.Builder
	err := c.DownloadVoiceSampleWriter(context.Background(), &buf, "v1", "s1")
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "proxy-test" {
		t.Errorf("got = %q", buf.String())
	}
}
