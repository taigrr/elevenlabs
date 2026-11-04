package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func assertJSON(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotV, wantV any
	if err := json.Unmarshal(got, &gotV); err != nil {
		t.Fatalf("invalid JSON body %q: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &wantV); err != nil {
		t.Fatalf("invalid want JSON: %v", err)
	}
	if !reflect.DeepEqual(gotV, wantV) {
		t.Errorf("body = %s\nwant   %s", got, want)
	}
}

// endpointCase describes one API call and what the server should observe.
type endpointCase struct {
	name     string
	call     func(ctx context.Context, c *Client) (any, error)
	method   string
	path     string
	query    string
	body     string              // expected JSON body, if any
	form     map[string][]string // expected multipart fields, if any
	accept   string
	respType string
	respBody string
	check    func(t *testing.T, result any)
}

func readAll(rc io.ReadCloser, err error) (any, error) {
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	return string(b), err
}

func coreEndpointCases() []endpointCase {
	return []endpointCase{
		{
			name:     "user.get",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.User.Get(ctx) },
			method:   http.MethodGet,
			path:     "/v1/user",
			respBody: `{"user_id":"u1","subscription":{"tier":"creator","max_credit_limit_extension":"unlimited","character_count":5},"seat_type":"workspace_admin","created_at":17}`,
			check: func(t *testing.T, r any) {
				u := r.(*User)
				if u.UserID != "u1" || u.Subscription.Tier != "creator" || !u.Subscription.MaxCreditLimitExtension.Unlimited ||
					u.Subscription.CharacterCount != 5 || u.SeatType != SeatTypeWorkspaceAdmin || u.CreatedAt != 17 {
					t.Errorf("user = %+v", u)
				}
			},
		},
		{
			name:     "user.subscription.get",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.User.Subscription.Get(ctx) },
			method:   http.MethodGet,
			path:     "/v1/user/subscription",
			respBody: `{"tier":"pro","max_credit_limit_extension":1000,"status":"active","open_invoices":[{"amount_due_cents":500,"discounts":[],"payment_intent_statusses":["succeeded"]}],"pending_change":{"kind":"cancellation","timestamp_seconds":9}}`,
			check: func(t *testing.T, r any) {
				s := r.(*Subscription)
				if s.Tier != "pro" || s.MaxCreditLimitExtension.Value != 1000 || s.Status != SubscriptionStatusActive ||
					len(s.OpenInvoices) != 1 || s.OpenInvoices[0].PaymentIntentStatusses[0] != PaymentIntentStatusSucceeded ||
					s.PendingChange.Kind != "cancellation" {
					t.Errorf("subscription = %+v", s)
				}
			},
		},
		{
			name: "voices.getAll",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.GetAll(ctx, &VoicesGetAllRequest{ShowLegacy: Ptr(true)})
			},
			method:   http.MethodGet,
			path:     "/v1/voices",
			query:    "show_legacy=true",
			respBody: `{"voices":[{"voice_id":"v1","name":"Rachel","category":"premade","settings":{"stability":0.5}}]}`,
			check: func(t *testing.T, r any) {
				v := r.(*GetVoicesResponse)
				if len(v.Voices) != 1 || v.Voices[0].VoiceID != "v1" || *v.Voices[0].Name != "Rachel" ||
					*v.Voices[0].Category != VoiceCategoryPremade || *v.Voices[0].Settings.Stability != 0.5 {
					t.Errorf("voices = %+v", v)
				}
			},
		},
		{
			name: "voices.search",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.Search(ctx, &VoicesSearchRequest{NextPageToken: "tok", Sort: "name", SortDirection: "asc", VoiceType: "personal", Category: "cloned", UseCases: []string{"a", "b"}, HighQuality: Ptr(true), MinNoticePeriodDays: Ptr(3)})
			},
			method:   http.MethodGet,
			path:     "/v2/voices",
			query:    "category=cloned&high_quality=true&min_notice_period_days=3&next_page_token=tok&sort=name&sort_direction=asc&use_cases=a&use_cases=b&voice_type=personal",
			respBody: `{"voices":[{"voice_id":"v2"}],"has_more":true,"total_count":2,"next_page_token":"n"}`,
			check: func(t *testing.T, r any) {
				v := r.(*GetVoicesV2Response)
				if !v.HasMore || v.TotalCount != 2 || *v.NextPageToken != "n" || v.Voices[0].VoiceID != "v2" {
					t.Errorf("search = %+v", v)
				}
			},
		},
		{
			name: "voices.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.Get(ctx, "voice/1", &VoicesGetRequest{WithSettings: Ptr(true)})
			},
			method:   http.MethodGet,
			path:     "/v1/voices/voice%2F1",
			query:    "with_settings=true",
			respBody: `{"voice_id":"voice/1","fine_tuning":{"state":{"eleven_multilingual_v2":"fine_tuned"}},"samples":[{"sample_id":"s1","size_bytes":10}]}`,
			check: func(t *testing.T, r any) {
				v := r.(*Voice)
				if v.FineTuning.State["eleven_multilingual_v2"] != FineTuningStateFineTuned || *v.Samples[0].SampleID != "s1" {
					t.Errorf("voice = %+v", v)
				}
			},
		},
		{
			name:     "voices.delete",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.Voices.Delete(ctx, "v1") },
			method:   http.MethodDelete,
			path:     "/v1/voices/v1",
			respBody: `{"status":"ok"}`,
			check:    checkStatusOK,
		},
		{
			name: "voices.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.Update(ctx, "v1", &VoicesUpdateRequest{
					Name:                  "New",
					Files:                 []File{NewFile(strings.NewReader("abc"), "s.mp3")},
					RemoveBackgroundNoise: Ptr(true),
					Description:           "desc",
					Labels:                map[string]string{"k": "v"},
					ModerateMetadata:      Ptr(false),
				})
			},
			method: http.MethodPost,
			path:   "/v1/voices/v1/edit",
			form: map[string][]string{
				"name": {"New"}, "files": {"file:s.mp3:abc"}, "remove_background_noise": {"true"},
				"description": {"desc"}, "labels": {`{"k":"v"}`}, "moderate_metadata": {"false"},
			},
			respBody: `{"status":"ok"}`,
			check:    checkStatusOK,
		},
		{
			name: "voices.ivc.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.IVC.Create(ctx, &IVCCreateRequest{
					Name:        "Clone",
					Files:       []File{NewFile(strings.NewReader("1"), "a.mp3"), NewFile(strings.NewReader("2"), "b.mp3")},
					Description: "d",
				})
			},
			method:   http.MethodPost,
			path:     "/v1/voices/add",
			form:     map[string][]string{"name": {"Clone"}, "files": {"file:a.mp3:1", "file:b.mp3:2"}, "description": {"d"}},
			respBody: `{"voice_id":"new","requires_verification":true}`,
			check: func(t *testing.T, r any) {
				if v := r.(*AddVoiceIVCResponse); v.VoiceID != "new" || !v.RequiresVerification {
					t.Errorf("ivc = %+v", v)
				}
			},
		},
		{
			name:     "voices.settings.getDefault",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.Voices.Settings.GetDefault(ctx) },
			method:   http.MethodGet,
			path:     "/v1/voices/settings/default",
			respBody: `{"stability":0.5,"similarity_boost":0.75,"use_speaker_boost":true,"speed":1}`,
			check: func(t *testing.T, r any) {
				if s := r.(*VoiceSettings); *s.SimilarityBoost != 0.75 || !*s.UseSpeakerBoost || *s.Speed != 1 {
					t.Errorf("settings = %+v", s)
				}
			},
		},
		{
			name:     "voices.settings.get",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.Voices.Settings.Get(ctx, "v1") },
			method:   http.MethodGet,
			path:     "/v1/voices/v1/settings",
			respBody: `{"stability":0.1}`,
			check: func(t *testing.T, r any) {
				if s := r.(*VoiceSettings); *s.Stability != 0.1 || s.Style != nil {
					t.Errorf("settings = %+v", s)
				}
			},
		},
		{
			name: "voices.settings.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.Settings.Update(ctx, "v1", &VoiceSettings{Stability: Ptr(0.0), Style: Ptr(0.3)})
			},
			method:   http.MethodPost,
			path:     "/v1/voices/v1/settings/edit",
			body:     `{"stability":0,"style":0.3}`,
			respBody: `{"status":"ok"}`,
			check:    checkStatusOK,
		},
		{
			name: "voices.samples.audio.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.Voices.Samples.Audio.Get(ctx, "v1", "s1"))
			},
			method:   http.MethodGet,
			path:     "/v1/voices/v1/samples/s1/audio",
			accept:   acceptAny,
			respType: "audio/mpeg",
			respBody: "ID3audio",
			check:    checkString("ID3audio"),
		},
		{
			name:     "samples.delete",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.Samples.Delete(ctx, "v1", "s1") },
			method:   http.MethodDelete,
			path:     "/v1/voices/v1/samples/s1",
			respBody: `{"status":"ok"}`,
			check:    checkStatusOK,
		},
		{
			name: "history.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.History.List(ctx, &HistoryListRequest{
					PageSize: Ptr(10), StartAfterHistoryItemID: "h0", VoiceID: "v", ModelID: "m",
					DateBeforeUnix: Ptr(int64(200)), DateAfterUnix: Ptr(int64(100)),
					SortDirection: SortDirectionAsc, Search: "hello", Source: HistorySourceTTS,
				})
			},
			method:   http.MethodGet,
			path:     "/v1/history",
			query:    "date_after_unix=100&date_before_unix=200&model_id=m&page_size=10&search=hello&sort_direction=asc&source=TTS&start_after_history_item_id=h0&voice_id=v",
			respBody: `{"history":[{"history_item_id":"h1","date_unix":5,"source":"TTS","settings":{"a":1}}],"last_history_item_id":"h1","has_more":false}`,
			check: func(t *testing.T, r any) {
				h := r.(*GetSpeechHistoryResponse)
				if len(h.History) != 1 || h.History[0].HistoryItemID != "h1" || *h.History[0].Source != HistorySourceTTS || *h.LastHistoryItemID != "h1" {
					t.Errorf("history = %+v", h)
				}
			},
		},
		{
			name:     "history.get",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.History.Get(ctx, "h1") },
			method:   http.MethodGet,
			path:     "/v1/history/h1",
			respBody: `{"history_item_id":"h1","alignments":{"alignment":{"characters":["a"],"character_start_times_seconds":[0],"character_end_times_seconds":[1]},"normalized_alignment":{"characters":[],"character_start_times_seconds":[],"character_end_times_seconds":[]}}}`,
			check: func(t *testing.T, r any) {
				if h := r.(*SpeechHistoryItem); h.Alignments.Alignment.Characters[0] != "a" {
					t.Errorf("item = %+v", h)
				}
			},
		},
		{
			name:     "history.delete",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.History.Delete(ctx, "h1") },
			method:   http.MethodDelete,
			path:     "/v1/history/h1",
			respBody: `{"status":"ok"}`,
			check:    checkStatusOK,
		},
		{
			name:     "history.getAudio",
			call:     func(ctx context.Context, c *Client) (any, error) { return readAll(c.History.GetAudio(ctx, "h1")) },
			method:   http.MethodGet,
			path:     "/v1/history/h1/audio",
			accept:   acceptAny,
			respType: "audio/mpeg",
			respBody: "mp3data",
			check:    checkString("mp3data"),
		},
		{
			name: "history.download",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.History.Download(ctx, &DownloadHistoryRequest{HistoryItemIDs: []string{"a", "b"}, OutputFormat: "wav"}))
			},
			method:   http.MethodPost,
			path:     "/v1/history/download",
			body:     `{"history_item_ids":["a","b"],"output_format":"wav"}`,
			accept:   acceptAny,
			respType: "application/zip",
			respBody: "PK",
			check:    checkString("PK"),
		},
		{
			name: "textToSpeech.convert",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.TextToSpeech.Convert(ctx, "v1", fullTTSRequest()))
			},
			method:   http.MethodPost,
			path:     "/v1/text-to-speech/v1",
			query:    "enable_logging=false&optimize_streaming_latency=2&output_format=pcm_16000",
			body:     fullTTSBody,
			accept:   acceptAny,
			respType: "audio/mpeg",
			respBody: "audio",
			check:    checkString("audio"),
		},
		{
			name: "textToSpeech.stream",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.TextToSpeech.Stream(ctx, "v1", &TextToSpeechRequest{Text: "hi", ModelID: "eleven_flash_v2_5"}))
			},
			method:   http.MethodPost,
			path:     "/v1/text-to-speech/v1/stream",
			body:     `{"text":"hi","model_id":"eleven_flash_v2_5"}`,
			accept:   acceptAny,
			respType: "audio/mpeg",
			respBody: "chunked-audio",
			check:    checkString("chunked-audio"),
		},
		{
			name: "textToSpeech.convertWithTimestamps",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.TextToSpeech.ConvertWithTimestamps(ctx, "v1", &TextToSpeechRequest{Text: "hi", OutputFormat: OutputFormatMP3_44100_128})
			},
			method:   http.MethodPost,
			path:     "/v1/text-to-speech/v1/with-timestamps",
			query:    "output_format=mp3_44100_128",
			body:     `{"text":"hi"}`,
			respBody: `{"audio_base64":"AAA=","alignment":{"characters":["h","i"],"character_start_times_seconds":[0,0.1],"character_end_times_seconds":[0.1,0.2]}}`,
			check: func(t *testing.T, r any) {
				a := r.(*AudioWithTimestamps)
				if a.AudioBase64 != "AAA=" || len(a.Alignment.Characters) != 2 || a.NormalizedAlignment != nil {
					t.Errorf("timestamps = %+v", a)
				}
			},
		},
		{
			name: "textToSpeech.streamWithTimestamps",
			call: func(ctx context.Context, c *Client) (any, error) {
				stream, err := c.TextToSpeech.StreamWithTimestamps(ctx, "v1", &TextToSpeechRequest{Text: "hi"})
				if err != nil {
					return nil, err
				}
				var chunks []string
				for chunk, err := range stream.All() {
					if err != nil {
						return nil, err
					}
					chunks = append(chunks, chunk.AudioBase64)
				}
				return strings.Join(chunks, "|"), nil
			},
			method:   http.MethodPost,
			path:     "/v1/text-to-speech/v1/stream/with-timestamps",
			body:     `{"text":"hi"}`,
			respBody: "{\"audio_base64\":\"AA==\"}\n{\"audio_base64\":\"AQ==\"}\n",
			check:    checkString("AA==|AQ=="),
		},
		{
			name: "textToSoundEffects.convert",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.TextToSoundEffects.Convert(ctx, &CreateSoundEffectRequest{
					Text: "rain", OutputFormat: OutputFormatMP3_22050_32, EnableLogging: Ptr(true),
					Loop: Ptr(true), DurationSeconds: Ptr(5.5), PromptInfluence: Ptr(0.0), ModelID: SoundEffectsModelV2,
				}))
			},
			method:   http.MethodPost,
			path:     "/v1/sound-generation",
			query:    "enable_logging=true&output_format=mp3_22050_32",
			body:     `{"text":"rain","loop":true,"duration_seconds":5.5,"prompt_influence":0,"model_id":"eleven_text_to_sound_v2"}`,
			accept:   acceptAny,
			respType: "audio/mpeg",
			respBody: "sfx",
			check:    checkString("sfx"),
		},
		{
			name: "speechToText.convert",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SpeechToText.Convert(ctx, fullSTTRequest())
			},
			method:   http.MethodPost,
			path:     "/v1/speech-to-text",
			query:    "enable_logging=false&token=tok",
			form:     fullSTTForm,
			respBody: `{"language_code":"en","language_probability":0.98,"text":"hello","words":[{"text":"hello","start":0,"end":0.5,"type":"word","speaker_id":"speaker_0","logprob":-0.1}]}`,
			check: func(t *testing.T, r any) {
				resp := r.(*SpeechToTextConvertResponse)
				if resp.Chunk == nil || resp.Chunk.Text != "hello" || resp.Chunk.Words[0].Type != SpeechToTextWordTypeWord || *resp.Chunk.Words[0].SpeakerID != "speaker_0" {
					t.Errorf("stt = %+v", resp)
				}
			},
		},
	}
}

const fullTTSBody = `{
	"text":"hello","model_id":"eleven_multilingual_v2","language_code":"en",
	"voice_settings":{"stability":0.5,"use_speaker_boost":false},
	"pronunciation_dictionary_locators":[{"pronunciation_dictionary_id":"pd","version_id":"v"}],
	"seed":7,"previous_text":"a","next_text":"b","previous_request_ids":["r1"],"next_request_ids":["r2"],
	"use_pvc_as_ivc":false,"apply_text_normalization":"off","apply_language_text_normalization":true,
	"audio_effects":{"filter_preset_id":"radio","pan":-0.5}
}`

func fullTTSRequest() *TextToSpeechRequest {
	return &TextToSpeechRequest{
		EnableLogging:                   Ptr(false),
		OptimizeStreamingLatency:        Ptr(2),
		OutputFormat:                    OutputFormatPCM16000,
		Text:                            "hello",
		ModelID:                         "eleven_multilingual_v2",
		LanguageCode:                    "en",
		VoiceSettings:                   &VoiceSettings{Stability: Ptr(0.5), UseSpeakerBoost: Ptr(false)},
		PronunciationDictionaryLocators: []PronunciationDictionaryLocator{{PronunciationDictionaryID: "pd", VersionID: "v"}},
		Seed:                            Ptr(int64(7)),
		PreviousText:                    "a",
		NextText:                        "b",
		PreviousRequestIDs:              []string{"r1"},
		NextRequestIDs:                  []string{"r2"},
		UsePVCAsIVC:                     Ptr(false),
		ApplyTextNormalization:          TextNormalizationOff,
		ApplyLanguageTextNormalization:  Ptr(true),
		AudioEffects:                    &AudioEffects{FilterPresetID: "radio", Pan: Ptr(-0.5)},
	}
}

func fullSTTRequest() *SpeechToTextConvertRequest {
	return &SpeechToTextConvertRequest{
		Token:                   "tok",
		EnableLogging:           Ptr(false),
		ModelID:                 SpeechToTextModelScribeV1,
		File:                    &File{Reader: bytes.NewReader([]byte("wav")), Name: "a.wav"},
		LanguageCode:            "en",
		TranscriptEdit:          `{"x":1}`,
		TagAudioEvents:          Ptr(true),
		NumSpeakers:             Ptr(2),
		TimestampsGranularity:   TimestampsGranularityCharacter,
		Diarize:                 Ptr(true),
		DiarizationThreshold:    Ptr(0.25),
		AdditionalFormats:       []ExportOptions{{Format: ExportFormatSRT, MaxCharactersPerLine: Ptr(42)}},
		FileFormat:              SpeechToTextFileFormatOther,
		CloudStorageURL:         "https://example.com/a.wav",
		SourceURL:               "https://example.com/v",
		Webhook:                 Ptr(false),
		WebhookID:               "wh",
		Temperature:             Ptr(0.5),
		Seed:                    Ptr(int64(9)),
		UseMultiChannel:         Ptr(false),
		MultichannelOutputStyle: MultichannelOutputStyleCombined,
		WebhookMetadata:         map[string]any{"job": "1"},
		EntityDetection:         []string{"pii", "phi"},
		NoVerbatim:              Ptr(true),
		UseSpeakerLibrary:       Ptr(false),
		DetectSpeakerRoles:      Ptr(true),
		EntityRedaction:         []string{"all"},
		EntityRedactionMode:     "mask",
		Keyterms:                []string{"ElevenLabs", "Scribe"},
	}
}

var fullSTTForm = map[string][]string{
	"model_id":                  {"scribe_v1"},
	"file":                      {"file:a.wav:wav"},
	"language_code":             {"en"},
	"transcript_edit":           {`{"x":1}`},
	"tag_audio_events":          {"true"},
	"num_speakers":              {"2"},
	"timestamps_granularity":    {"character"},
	"diarize":                   {"true"},
	"diarization_threshold":     {"0.25"},
	"additional_formats":        {`[{"format":"srt","max_characters_per_line":42}]`},
	"file_format":               {"other"},
	"cloud_storage_url":         {"https://example.com/a.wav"},
	"source_url":                {"https://example.com/v"},
	"webhook":                   {"false"},
	"webhook_id":                {"wh"},
	"temperature":               {"0.5"},
	"seed":                      {"9"},
	"use_multi_channel":         {"false"},
	"multichannel_output_style": {"combined"},
	"webhook_metadata":          {`{"job":"1"}`},
	"entity_detection":          {"pii", "phi"},
	"no_verbatim":               {"true"},
	"use_speaker_library":       {"false"},
	"detect_speaker_roles":      {"true"},
	"entity_redaction":          {"all"},
	"entity_redaction_mode":     {"mask"},
	"keyterms":                  {"ElevenLabs", "Scribe"},
}

func checkStatusOK(t *testing.T, r any) {
	t.Helper()
	if s := r.(*StatusResponse); s.Status != "ok" {
		t.Errorf("status = %q", s.Status)
	}
}

func checkString(want string) func(t *testing.T, r any) {
	return func(t *testing.T, r any) {
		t.Helper()
		if r.(string) != want {
			t.Errorf("result = %q, want %q", r, want)
		}
	}
}

func TestEndpoints(t *testing.T) {
	for _, tt := range endpointCases() {
		t.Run(tt.name, func(t *testing.T) {
			respType := tt.respType
			if respType == "" {
				respType = contentTypeJSON
			}
			client, ts := newTestServer(t, respond(http.StatusOK, respType, tt.respBody))
			result, err := tt.call(context.Background(), client)
			if err != nil {
				t.Fatalf("call failed: %v", err)
			}
			got := ts.last(t)
			if got.Method != tt.method || got.Path != tt.path {
				t.Errorf("request = %s %s, want %s %s", got.Method, got.Path, tt.method, tt.path)
			}
			if got.RawQuery != tt.query {
				t.Errorf("query = %q\nwant    %q", got.RawQuery, tt.query)
			}
			wantAccept := tt.accept
			if wantAccept == "" {
				wantAccept = contentTypeJSON
			}
			if a := got.Header.Get(headerAccept); a != wantAccept {
				t.Errorf("accept = %q, want %q", a, wantAccept)
			}
			if k := got.Header.Get(headerAPIKey); k != "test-key" {
				t.Errorf("api key = %q", k)
			}
			switch {
			case tt.body != "":
				if got.ContentType != contentTypeJSON {
					t.Errorf("content type = %q", got.ContentType)
				}
				assertJSON(t, got.Body, tt.body)
			case tt.form != nil:
				if fields := formFields(t, got); !reflect.DeepEqual(fields, tt.form) {
					t.Errorf("form = %v\nwant   %v", fields, tt.form)
				}
			default:
				if len(got.Body) != 0 {
					t.Errorf("unexpected body %q", got.Body)
				}
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestEndpointsMapErrors(t *testing.T) {
	statuses := []struct {
		code     int
		sentinel error
	}{
		{http.StatusUnauthorized, ErrUnauthorized},
		{http.StatusNotFound, ErrNotFound},
		{http.StatusUnprocessableEntity, ErrUnprocessableEntity},
	}
	for _, tt := range endpointCases() {
		for _, st := range statuses {
			t.Run(tt.name+"/"+http.StatusText(st.code), func(t *testing.T) {
				client, _ := newTestServer(t, respond(st.code, contentTypeJSON, `{"detail":{"status":"x","message":"nope"}}`))
				_, err := tt.call(context.Background(), client)
				if !errors.Is(err, st.sentinel) {
					t.Fatalf("error = %v, want %v", err, st.sentinel)
				}
				var apiErr *APIError
				if !errors.As(err, &apiErr) || apiErr.Message != "nope" {
					t.Errorf("api error = %+v", apiErr)
				}
			})
		}
	}
}

func TestSpeechToTextResponseVariants(t *testing.T) {
	tests := []struct {
		name string
		body string
		want func(r SpeechToTextConvertResponse) bool
	}{
		{"chunk", `{"language_code":"en","language_probability":1,"text":"hi","words":[]}`, func(r SpeechToTextConvertResponse) bool {
			return r.Chunk != nil && r.Chunk.Text == "hi" && r.Multichannel == nil && r.Webhook == nil
		}},
		{"multichannel", `{"transcripts":[{"language_code":"en","language_probability":1,"text":"a","words":[],"channel_index":0}],"transcription_id":"t"}`, func(r SpeechToTextConvertResponse) bool {
			return r.Multichannel != nil && len(r.Multichannel.Transcripts) == 1 && *r.Multichannel.TranscriptionID == "t"
		}},
		{"webhook", `{"message":"queued","request_id":"r","transcription_id":"t"}`, func(r SpeechToTextConvertResponse) bool {
			return r.Webhook != nil && r.Webhook.RequestID == "r"
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r SpeechToTextConvertResponse
			if err := json.Unmarshal([]byte(tt.body), &r); err != nil {
				t.Fatal(err)
			}
			if !tt.want(r) {
				t.Errorf("unexpected variant: %+v", r)
			}
			out, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONSubset(t, out, tt.body)
		})
	}
	if out, _ := json.Marshal(SpeechToTextConvertResponse{}); string(out) != "null" {
		t.Errorf("empty marshal = %s", out)
	}
}

// assertJSONSubset checks every key of want is present with the same value in got.
func assertJSONSubset(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotM, wantM map[string]any
	_ = json.Unmarshal(got, &gotM)
	_ = json.Unmarshal([]byte(want), &wantM)
	for k, v := range wantM {
		if !reflect.DeepEqual(gotM[k], v) {
			t.Errorf("key %s = %v, want %v", k, gotM[k], v)
		}
	}
}

func TestSpeechToTextRequiresModel(t *testing.T) {
	client, ts := newTestServer(t, respond(200, contentTypeJSON, `{}`))
	if _, err := client.SpeechToText.Convert(context.Background(), &SpeechToTextConvertRequest{}); err == nil {
		t.Error("expected error without model id")
	}
	if _, err := client.SpeechToText.Convert(context.Background(), nil); err == nil {
		t.Error("expected error for nil request")
	}
	if len(ts.all()) != 0 {
		t.Error("request sent despite validation error")
	}
}

func TestHistoryListAll(t *testing.T) {
	pages := map[string]string{
		"":   `{"history":[{"history_item_id":"h1"},{"history_item_id":"h2"}],"last_history_item_id":"h2","has_more":true}`,
		"h2": `{"history":[{"history_item_id":"h3"}],"last_history_item_id":"h3","has_more":false}`,
	}
	client, ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, ok := pages[r.URL.Query().Get("start_after_history_item_id")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, body)
	})
	var ids []string
	for item, err := range client.History.ListAll(context.Background(), &HistoryListRequest{PageSize: Ptr(2), VoiceID: "v"}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, item.HistoryItemID)
	}
	if strings.Join(ids, ",") != "h1,h2,h3" {
		t.Errorf("ids = %v", ids)
	}
	reqs := ts.all()
	if len(reqs) != 2 || reqs[1].RawQuery != "page_size=2&start_after_history_item_id=h2&voice_id=v" {
		t.Errorf("requests = %+v", reqs)
	}
}

func TestHistoryListAllYieldsError(t *testing.T) {
	client, _ := newTestServer(t, respond(http.StatusUnauthorized, contentTypeJSON, `{}`))
	for _, err := range client.History.ListAll(context.Background(), nil) {
		if !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("error = %v", err)
		}
		return
	}
	t.Fatal("iterator yielded nothing")
}

func TestVoicesSearchAll(t *testing.T) {
	client, ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("next_page_token") == "" {
			_, _ = io.WriteString(w, `{"voices":[{"voice_id":"a"}],"has_more":true,"total_count":2,"next_page_token":"p2"}`)
			return
		}
		_, _ = io.WriteString(w, `{"voices":[{"voice_id":"b"}],"has_more":false,"total_count":2}`)
	})
	var ids []string
	for v, err := range client.Voices.SearchAll(context.Background(), &VoicesSearchRequest{Search: "x"}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, v.VoiceID)
	}
	if strings.Join(ids, ",") != "a,b" {
		t.Errorf("ids = %v", ids)
	}
	if q := ts.last(t).RawQuery; q != "next_page_token=p2&search=x" {
		t.Errorf("second page query = %q", q)
	}
}

func TestNilRequestsSendEmptyBodies(t *testing.T) {
	client, ts := newTestServer(t, respond(200, "audio/mpeg", "a"))
	rc, err := client.TextToSpeech.Convert(context.Background(), "v", nil)
	if err != nil {
		t.Fatal(err)
	}
	rc.Close()
	assertJSON(t, ts.last(t).Body, `{"text":""}`)
}
