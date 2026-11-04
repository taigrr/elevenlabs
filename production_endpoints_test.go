package elevenlabs

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func productionEndpointCases() []endpointCase {
	start := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return []endpointCase{
		// productions
		{
			name: "productions.orders.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Productions.Orders.List(ctx, &OrdersListRequest{PageSize: Ptr(5), Status: []OrderState{OrderStateOpen, OrderStateDone}, StartDate: &start})
			},
			method: http.MethodGet, path: "/v1/productions/orders",
			query:    "page_size=5&start_date=2026-01-02T03%3A04%3A05Z&status=open&status=done",
			respBody: `{"orders":[{"order_id":"o","name":"n","state":"open","submitted_at":"2026-01-02T03:04:05Z"}]}`,
			check: func(t *testing.T, r any) {
				if o := r.(*OrderList).Orders[0]; o.State != OrderStateOpen || !o.SubmittedAt.Equal(start) {
					t.Errorf("order = %+v", o)
				}
			},
		},
		{
			name: "productions.orders.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Productions.Orders.Create(ctx, &OrderCreateRequest{Sandbox: Ptr(true)})
			},
			method: http.MethodPost, path: "/v1/productions/orders", body: `{"sandbox":true}`, respBody: `{"order_id":"o"}`,
		},
		{
			name:   "productions.orders.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Productions.Orders.Get(ctx, "o") },
			method: http.MethodGet, path: "/v1/productions/orders/o",
			respBody: `{"order_id":"o","name":"n","state":"paid","items":[{"item_id":"i","item":{"kind":"dub","media_id":"m","source_language":"en","destination_languages":["es"],"include_captions":true,"include_source_captions":false}}],"created_at":"2026-01-02T03:04:05Z"}`,
			check: func(t *testing.T, r any) {
				o := r.(*Order)
				if o.Items[0].Item.Kind != OrderItemKindDub || !*o.Items[0].Item.IncludeCaptions {
					t.Errorf("order = %+v", o)
				}
			},
		},
		{
			name:   "productions.orders.update",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Productions.Orders.Update(ctx, "o", "new") },
			method: http.MethodPatch, path: "/v1/productions/orders/o", body: `{"request":{"name":"new"}}`, respBody: `{"name":"new"}`,
		},
		{
			name:   "productions.orders.submit",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Productions.Orders.Submit(ctx, "o") },
			method: http.MethodPost, path: "/v1/productions/orders/o/submit",
			respBody: `{"order_id":"o","state":"submitted","submitted_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			name: "productions.orders.deliverables.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Productions.Orders.Deliverables.List(ctx, "o")
			},
			method: http.MethodGet, path: "/v1/productions/orders/o/deliverables",
			respBody: `{"deliverables":[{"signed_url":"u","content_type":"audio/mpeg","name":"n"}]}`,
		},
		{
			name: "productions.orders.items.upsert",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Productions.Orders.Items.Upsert(ctx, "o", &OrderItemUpsertRequest{Item: OrderItem{Kind: OrderItemKindTranscription, MediaIDs: []string{"m"}, SourceLanguage: "en", Verbatim: Ptr(true)}})
			},
			method: http.MethodPost, path: "/v1/productions/orders/o/items",
			body:     `{"request":{"item":{"kind":"transcription","media_ids":["m"],"source_language":"en","verbatim":true}}}`,
			respBody: `{"item_id":"i","quote":{"amount_usd":12.5}}`,
			check: func(t *testing.T, r any) {
				if q := r.(*OrderItemUpsertResponse).Quote; q.AmountUSD != 12.5 {
					t.Errorf("quote = %+v", q)
				}
			},
		},
		{
			name: "productions.orders.items.remove",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Productions.Orders.Items.Remove(ctx, "o", "i")
			},
			method: http.MethodDelete, path: "/v1/productions/orders/o/items/i", respBody: `{"success":true}`,
		},
		{
			name: "productions.orders.languages.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Productions.Orders.Languages.List(ctx, OrderItemKindSubtitles)
			},
			method: http.MethodGet, path: "/v1/productions/orders/languages/subtitles",
			respBody: `{"kind":"pair","language_pairs":[{"source_language":{"code":"en","label":"English"},"destination_languages":[{"code":"es","label":"Spanish"}]}]}`,
			check: func(t *testing.T, r any) {
				if l := r.(*ProductionLanguages); l.LanguagePairs[0].DestinationLanguages[0].Code != "es" {
					t.Errorf("languages = %+v", l)
				}
			},
		},
		{
			name: "productions.orders.media.register",
			call: func(ctx context.Context, c *Client) (any, error) {
				f := testFile("v.mp4", "mp4")
				return c.Productions.Orders.Media.Register(ctx, "o", &MediaRegisterRequest{DeclaredLanguage: "en", Media: &f})
			},
			method: http.MethodPost, path: "/v1/productions/orders/o/media",
			form:     map[string][]string{"declared_language": {"en"}, "media": {"file:v.mp4:mp4"}},
			respBody: `{"media_id":"m"}`,
		},
		{
			name: "productions.orders.media.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Productions.Orders.Media.Get(ctx, "o", "m")
			},
			method: http.MethodGet, path: "/v1/productions/orders/o/media/m",
			respBody: `{"media_id":"m","name":"n","content_type":"video/mp4","signed_url":"u"}`,
		},
		// dubbing
		{
			name: "dubbing.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.List(ctx, &DubbingListRequest{PageSize: Ptr(2), DubbingStatuses: []DubbingStatus{DubbingStatusDubbed}, DubbingModels: []DubbingModel{DubbingModelV2}, OrderDirection: "ASCENDING"})
			},
			method: http.MethodGet, path: "/v1/dubbing",
			query:    "dubbing_models=dubbing_v2&dubbing_statuses=dubbed&order_direction=ASCENDING&page_size=2",
			respBody: `{"dubs":[{"dubbing_id":"d","name":"n","status":"dubbed","target_languages":["es"],"created_at":"2026-01-02T03:04:05Z"}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if d := r.(*DubbingPage).Dubs[0]; d.Status != DubbingStatusDubbed || d.TargetLanguages[0] != "es" {
					t.Errorf("dub = %+v", d)
				}
			},
		},
		{
			name: "dubbing.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				f := testFile("a.mp3", "mp3")
				return c.Dubbing.Create(ctx, &DubbingCreateRequest{File: &f, TargetLang: "es", NumSpeakers: Ptr(2), Watermark: Ptr(true), Mode: "automatic"})
			},
			method: http.MethodPost, path: "/v1/dubbing",
			form:     map[string][]string{"file": {"file:a.mp3:mp3"}, "target_lang": {"es"}, "num_speakers": {"2"}, "watermark": {"true"}, "mode": {"automatic"}},
			respBody: `{"dubbing_id":"d","expected_duration_sec":12}`,
			check: func(t *testing.T, r any) {
				if d := r.(*DubbingCreateResponse); d.DubbingID != "d" || d.ExpectedDurationSec != 12 {
					t.Errorf("resp = %+v", d)
				}
			},
		},
		{
			name:   "dubbing.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Dubbing.Get(ctx, "d") },
			method: http.MethodGet, path: "/v1/dubbing/d", respBody: `{"dubbing_id":"d","created_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			name:   "dubbing.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Dubbing.Delete(ctx, "d") },
			method: http.MethodDelete, path: "/v1/dubbing/d", respBody: `{"status":"ok"}`,
		},
		{
			name:   "dubbing.audio.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return readAll(c.Dubbing.Audio.Get(ctx, "d", "es")) },
			method: http.MethodGet, path: "/v1/dubbing/d/audio/es", accept: acceptAny, respType: "audio/mpeg", respBody: "dub",
			check: checkString("dub"),
		},
		{
			name: "dubbing.transcript.getTranscriptForDub",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Transcript.GetTranscriptForDub(ctx, "d", "es", TranscriptFormatSRT)
			},
			method: http.MethodGet, path: "/v1/dubbing/d/transcript/es", query: "format_type=srt", accept: acceptAny,
			respType: "text/plain", respBody: "1\n00:00:00,000 --> 00:00:01,000\nhola\n",
			check: func(t *testing.T, r any) {
				if tr := r.(*DubTranscript); tr.JSON != nil || tr.Text != "1\n00:00:00,000 --> 00:00:01,000\nhola\n" {
					t.Errorf("transcript = %+v", tr)
				}
			},
		},
		{
			name: "dubbing.transcripts.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Transcripts.Get(ctx, "d", "es", TranscriptFormatJSON)
			},
			method: http.MethodGet, path: "/v1/dubbing/d/transcripts/es/format/json",
			respBody: `{"transcript_format":"json","json":{"language":"es","utterances":[{"text":"hola","speaker_id":"s"}]}}`,
			check: func(t *testing.T, r any) {
				if tr := r.(*DubbingTranscripts); *tr.JSON.Utterances[0].Text != "hola" {
					t.Errorf("transcripts = %+v", tr)
				}
			},
		},
		{
			name: "dubbing.project.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.List(ctx, &DubbingProjectListRequest{Status: "ready", SortDirection: "DESCENDING"})
			},
			method: http.MethodGet, path: "/v1/dubbing/project", query: "sort_direction=DESCENDING&status=ready",
			respBody: `{"projects":[{"project_id":"p","status":"ready","revision":1,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}]}`,
		},
		{
			name: "dubbing.project.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Create(ctx, &DubbingProjectCreateRequest{SourceURL: "https://v", ModelID: "dubbing_v2", Keyterms: []string{"a", "b"}, WebhookIDs: []string{"w"}})
			},
			method: http.MethodPost, path: "/v1/dubbing/project",
			form:     map[string][]string{"source_url": {"https://v"}, "model_id": {"dubbing_v2"}, "keyterms": {"a", "b"}, "webhook_ids": {"w"}},
			respBody: `{"project_id":"p","status":"queued","revision":0,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			name:   "dubbing.project.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Dubbing.Project.Get(ctx, "p") },
			method: http.MethodGet, path: "/v1/dubbing/project/p",
			respBody: `{"project_id":"p","status":"failed","error":{"message_type":"error","error":"boom"},"revision":0,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
			check: func(t *testing.T, r any) {
				if p := r.(*DubbingProject); p.Error.Error != "boom" {
					t.Errorf("project = %+v", p)
				}
			},
		},
		{
			name:   "dubbing.project.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return voidCall(c.Dubbing.Project.Delete(ctx, "p")) },
			method: http.MethodDelete, path: "/v1/dubbing/project/p",
		},
		{
			name: "dubbing.project.language.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Language.List(ctx, "p", &DubbingLanguageListRequest{PageSize: Ptr(1)})
			},
			method: http.MethodGet, path: "/v1/dubbing/project/p/language", query: "page_size=1", respBody: `{"languages":[]}`,
		},
		{
			name: "dubbing.project.language.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Language.Create(ctx, "p", &DubbingLanguageCreateRequest{TargetLanguage: "es", VoiceSettings: &VoiceSettings{Stability: Ptr(0.4)}})
			},
			method: http.MethodPost, path: "/v1/dubbing/project/p/language",
			body:     `{"target_language":"es","voice_settings":{"stability":0.4}}`,
			respBody: `{"language_id":"l","project_id":"p","target_language":"es","status":"queued","revision":0,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			name: "dubbing.project.language.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Language.Get(ctx, "p", "l")
			},
			method: http.MethodGet, path: "/v1/dubbing/project/p/language/l",
			respBody: `{"language_id":"l","status":"completed","outputs":{"lossless_audio":"u"},"revision":1,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			name: "dubbing.project.language.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return voidCall(c.Dubbing.Project.Language.Delete(ctx, "p", "l"))
			},
			method: http.MethodDelete, path: "/v1/dubbing/project/p/language/l",
		},
		{
			name: "dubbing.project.language.transcript.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Language.Transcript.Get(ctx, "p", "l")
			},
			method: http.MethodGet, path: "/v1/dubbing/project/p/language/l/transcript",
			respBody: `{"target_language":"es","segments":[{"id":"s","speaker_id":"a","start_s":0,"end_s":1,"source_text":"hi","translation":"hola"}],"revision":2}`,
		},
		{
			name: "dubbing.project.language.transcript.updateSegment",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Language.Transcript.UpdateSegment(ctx, "p", "l", "s", &DubbingTargetSegmentUpdate{Translation: Ptr("hola")})
			},
			method: http.MethodPatch, path: "/v1/dubbing/project/p/language/l/transcript/segment/s",
			body: `{"translation":"hola"}`, respBody: `{"segment":{"id":"s"},"revision":3}`,
		},
		{
			name: "dubbing.project.language.transcript.updateSegments",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Language.Transcript.UpdateSegments(ctx, "p", "l", map[string]DubbingTargetSegmentUpdate{"s": {Translation: Ptr("x")}})
			},
			method: http.MethodPatch, path: "/v1/dubbing/project/p/language/l/transcript/segments",
			body: `{"segments":{"s":{"translation":"x"}}}`, respBody: `{"segments":[],"revision":4}`,
		},
		{
			name: "dubbing.project.language.transcript.regenerate",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Language.Transcript.Regenerate(ctx, "p", "l")
			},
			method: http.MethodPost, path: "/v1/dubbing/project/p/language/l/transcript/regenerate",
			respBody: `{"regenerated_segment_ids":["s"],"regenerated_seconds":1.5,"charged_seconds":0,"free_regeneration_seconds_remaining":10}`,
		},
		{
			name:   "dubbing.project.transcript.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Dubbing.Project.Transcript.Get(ctx, "p") },
			method: http.MethodGet, path: "/v1/dubbing/project/p/transcript",
			respBody: `{"language":"en","segments":[{"id":"s","text":"hi","speaker_id":"a","start_s":0,"end_s":1}],"revision":1}`,
		},
		{
			name: "dubbing.project.transcript.deleteSegment",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Transcript.DeleteSegment(ctx, "p", "s")
			},
			method: http.MethodDelete, path: "/v1/dubbing/project/p/transcript/segment/s", respBody: `{"revision":2}`,
		},
		{
			name: "dubbing.project.transcript.updateSegment",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Transcript.UpdateSegment(ctx, "p", "s", &DubbingSegmentUpdate{Text: Ptr("hey"), EndS: Ptr(2.0)})
			},
			method: http.MethodPatch, path: "/v1/dubbing/project/p/transcript/segment/s",
			body: `{"text":"hey","end_s":2}`, respBody: `{"segment":{"id":"s"},"revision":3}`,
		},
		{
			name: "dubbing.project.transcript.updateSegments",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Transcript.UpdateSegments(ctx, "p", nil)
			},
			method: http.MethodPatch, path: "/v1/dubbing/project/p/transcript/segments",
			body: `{"segments":{}}`, respBody: `{"segments":[],"revision":4}`,
		},
		{
			name: "dubbing.project.transcript.createSegment",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Project.Transcript.CreateSegment(ctx, "p", &DubbingSegmentCreate{Text: "t", SpeakerID: "a", StartS: 1, EndS: 2})
			},
			method: http.MethodPost, path: "/v1/dubbing/project/p/transcript/segment",
			body: `{"text":"t","speaker_id":"a","start_s":1,"end_s":2}`, respBody: `{"segment":{"id":"n"},"revision":5}`,
		},
		{
			name:   "dubbing.resource.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Dubbing.Resource.Get(ctx, "d") },
			method: http.MethodGet, path: "/v1/dubbing/resource/d",
			respBody: `{"id":"d","version":3,"source_language":"en","target_languages":["es"],"input":{"src":"s","content_type":"video/mp4","bucket_name":"b","random_path_slug":"r","duration_secs":5,"is_audio":false,"url":"u"},"speaker_tracks":{},"speaker_segments":{"s":{"id":"s","start_time":0,"end_time":1,"text":"hi","subtitles":[],"dubs":{"es":{"start_time":0,"end_time":1,"subtitles":[],"audio_stale":true}}}},"renders":{"r":{"id":"r","version":1,"type":"mp4","status":"complete"}}}`,
			check: func(t *testing.T, r any) {
				res := r.(*DubbingResource)
				if !res.SpeakerSegments["s"].Dubs["es"].AudioStale || *res.Renders["r"].Type != RenderTypeMP4 {
					t.Errorf("resource = %+v", res)
				}
			},
		},
		{
			name: "dubbing.resource.migrateSegments",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.MigrateSegments(ctx, "d", []string{"a"}, "sp")
			},
			method: http.MethodPost, path: "/v1/dubbing/resource/d/migrate-segments",
			body: `{"segment_ids":["a"],"speaker_id":"sp"}`, respBody: `{"version":4}`,
		},
		{
			name: "dubbing.resource.transcribe",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Transcribe(ctx, "d", []string{"a"})
			},
			method: http.MethodPost, path: "/v1/dubbing/resource/d/transcribe", body: `{"segments":["a"]}`, respBody: `{"version":4}`,
		},
		{
			name: "dubbing.resource.translate",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Translate(ctx, "d", []string{"a"}, []string{"es"})
			},
			method: http.MethodPost, path: "/v1/dubbing/resource/d/translate", body: `{"segments":["a"],"languages":["es"]}`, respBody: `{"version":4}`,
		},
		{
			name:   "dubbing.resource.dub",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Dubbing.Resource.Dub(ctx, "d", nil, nil) },
			method: http.MethodPost, path: "/v1/dubbing/resource/d/dub", body: `{"segments":[]}`, respBody: `{"version":4}`,
		},
		{
			name: "dubbing.resource.render",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Render(ctx, "d", "es", &DubbingRenderRequest{RenderType: RenderTypeMP3, NormalizeVolume: Ptr(true)})
			},
			method: http.MethodPost, path: "/v1/dubbing/resource/d/render/es",
			body: `{"render_type":"mp3","normalize_volume":true}`, respBody: `{"version":5,"render_id":"r"}`,
		},
		{
			name: "dubbing.resource.language.add",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Language.Add(ctx, "d", "fr")
			},
			method: http.MethodPost, path: "/v1/dubbing/resource/d/language", body: `{"language":"fr"}`, respBody: `{"version":6}`,
		},
		{
			name: "dubbing.resource.segment.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Segment.Update(ctx, "d", "s", "es", &SegmentUpdateRequest{Text: Ptr("hola"), StartTime: Ptr(0.5)})
			},
			method: http.MethodPatch, path: "/v1/dubbing/resource/d/segment/s/es",
			body: `{"start_time":0.5,"text":"hola"}`, respBody: `{"version":7}`,
		},
		{
			name: "dubbing.resource.segment.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Segment.Delete(ctx, "d", "s")
			},
			method: http.MethodDelete, path: "/v1/dubbing/resource/d/segment/s", respBody: `{"version":8}`,
		},
		{
			name: "dubbing.resource.speaker.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Speaker.Update(ctx, "d", "sp", &SpeakerRequest{VoiceID: "track-clone", Languages: []string{"es"}})
			},
			method: http.MethodPatch, path: "/v1/dubbing/resource/d/speaker/sp",
			body: `{"voice_id":"track-clone","languages":["es"]}`, respBody: `{"version":9}`,
		},
		{
			name: "dubbing.resource.speaker.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Speaker.Create(ctx, "d", &SpeakerRequest{SpeakerName: "Bob", VoiceStability: Ptr(0.3), Languages: []string{"x"}})
			},
			method: http.MethodPost, path: "/v1/dubbing/resource/d/speaker",
			body: `{"speaker_name":"Bob","voice_stability":0.3}`, respBody: `{"version":10,"speaker_id":"sp2"}`,
		},
		{
			name: "dubbing.resource.speaker.findSimilarVoices",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Speaker.FindSimilarVoices(ctx, "d", "sp")
			},
			method: http.MethodGet, path: "/v1/dubbing/resource/d/speaker/sp/similar-voices",
			respBody: `{"voices":[{"voice_id":"v","name":"n","category":"premade"}]}`,
			check: func(t *testing.T, r any) {
				if v := r.(*SimilarVoicesResponse).Voices[0]; v.Category != VoiceCategoryPremade {
					t.Errorf("voice = %+v", v)
				}
			},
		},
		{
			name: "dubbing.resource.speaker.segment.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Dubbing.Resource.Speaker.Segment.Create(ctx, "d", "sp", &SegmentCreateRequest{StartTime: 1, EndTime: 2, Text: Ptr("t")})
			},
			method: http.MethodPost, path: "/v1/dubbing/resource/d/speaker/sp/segment",
			body: `{"start_time":1,"end_time":2,"text":"t"}`, respBody: `{"version":11,"new_segment":"s2"}`,
		},
		// studio
		{
			name: "studio.createPodcast",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.CreatePodcast(ctx, &PodcastCreateRequest{ModelID: "m", Mode: ConversationPodcast("h", "g"), Source: PodcastSources{URLSource("https://x")}, DurationScale: "short"})
			},
			method: http.MethodPost, path: "/v1/studio/podcasts",
			body:     `{"model_id":"m","mode":{"type":"conversation","conversation":{"host_voice_id":"h","guest_voice_id":"g"}},"source":{"type":"url","url":"https://x"},"duration_scale":"short"}`,
			respBody: `{"project":{"project_id":"p","name":"n","state":"creating","access_level":"admin"}}`,
			check: func(t *testing.T, r any) {
				if p := r.(*ProjectResponse).Project; p.ProjectID != "p" || p.State != ProjectStateCreating {
					t.Errorf("project = %+v", p)
				}
			},
		},
		{
			name:   "studio.projects.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Studio.Projects.List(ctx) },
			method: http.MethodGet, path: "/v1/studio/projects", respBody: `{"projects":[{"project_id":"p"}]}`,
		},
		{
			name: "studio.projects.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				f := testFile("b.epub", "epub")
				return c.Studio.Projects.Create(ctx, &ProjectCreateRequest{Name: "n", FromDocument: &f, Genres: []string{"a", "b"}, AutoConvert: Ptr(true),
					PronunciationDictionaryLocators: []string{PronunciationDictionaryLocator{PronunciationDictionaryID: "d"}.String()}})
			},
			method: http.MethodPost, path: "/v1/studio/projects",
			form: map[string][]string{"name": {"n"}, "from_document": {"file:b.epub:epub"}, "genres": {"a", "b"}, "auto_convert": {"true"},
				"pronunciation_dictionary_locators": {`{"pronunciation_dictionary_id":"d"}`}},
			respBody: `{"project":{"project_id":"p"}}`,
		},
		{
			name:   "studio.projects.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Studio.Projects.Get(ctx, "p", "sh") },
			method: http.MethodGet, path: "/v1/studio/projects/p", query: "share_id=sh",
			respBody: `{"project_id":"p","quality_preset":"high","chapters":[{"chapter_id":"c","name":"one","state":"default","can_be_downloaded":true}],"voices":[],"assets":[{"type":"image"}]}`,
			check: func(t *testing.T, r any) {
				p := r.(*ProjectExtended)
				if p.ProjectID != "p" || p.QualityPreset != QualityPresetHigh || p.Chapters[0].ChapterID != "c" || len(p.Assets) != 1 {
					t.Errorf("project = %+v", p)
				}
			},
		},
		{
			name: "studio.projects.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Update(ctx, "p", &ProjectUpdateRequest{Name: "n", DefaultTitleVoiceID: "t", DefaultParagraphVoiceID: "q", Author: "a"})
			},
			method: http.MethodPost, path: "/v1/studio/projects/p",
			body: `{"name":"n","default_title_voice_id":"t","default_paragraph_voice_id":"q","author":"a"}`, respBody: `{"project":{"project_id":"p"}}`,
		},
		{
			name:   "studio.projects.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Studio.Projects.Delete(ctx, "p") },
			method: http.MethodDelete, path: "/v1/studio/projects/p", respBody: `{"status":"ok"}`,
		},
		{
			name:   "studio.projects.convert",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Studio.Projects.Convert(ctx, "p") },
			method: http.MethodPost, path: "/v1/studio/projects/p/convert", respBody: `{"status":"ok"}`,
		},
		{
			name:   "studio.projects.getMutedTracks",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Studio.Projects.GetMutedTracks(ctx, "p") },
			method: http.MethodGet, path: "/v1/studio/projects/p/muted-tracks", respBody: `{"chapter_ids":["c"]}`,
		},
		{
			name:   "studio.projects.chapters.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Studio.Projects.Chapters.List(ctx, "p") },
			method: http.MethodGet, path: "/v1/studio/projects/p/chapters", respBody: `{"chapters":[]}`,
		},
		{
			name: "studio.projects.chapters.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Chapters.Create(ctx, "p", &ChapterCreateRequest{Name: "c", FromURL: "https://x"})
			},
			method: http.MethodPost, path: "/v1/studio/projects/p/chapters",
			body: `{"name":"c","from_url":"https://x"}`, respBody: `{"chapter":{"chapter_id":"c","content":{"blocks":[]}}}`,
		},
		{
			name: "studio.projects.chapters.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Chapters.Get(ctx, "p", "c")
			},
			method: http.MethodGet, path: "/v1/studio/projects/p/chapters/c",
			respBody: `{"chapter_id":"c","name":"n","state":"converting","content":{"blocks":[{"block_id":"b","nodes":[{"type":"tts_node","text":"hi","voice_id":"v","project_voice_ref_id":"r"}]}]}}`,
			check: func(t *testing.T, r any) {
				ch := r.(*ChapterWithContent)
				if ch.ChapterID != "c" || ch.State != ChapterStateConverting || ch.Content.Blocks[0].Nodes[0].Text != "hi" {
					t.Errorf("chapter = %+v", ch)
				}
			},
		},
		{
			name: "studio.projects.chapters.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Chapters.Update(ctx, "p", "c", &ChapterUpdateRequest{Content: &ChapterContent{Blocks: []ChapterContentBlock{{SubType: "p", Nodes: []ChapterContentNode{TTSNode("hi", "v")}}}}})
			},
			method: http.MethodPost, path: "/v1/studio/projects/p/chapters/c",
			body:     `{"content":{"blocks":[{"sub_type":"p","nodes":[{"type":"tts_node","text":"hi","voice_id":"v"}]}]}}`,
			respBody: `{"chapter":{"chapter_id":"c"}}`,
		},
		{
			name: "studio.projects.chapters.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Chapters.Delete(ctx, "p", "c")
			},
			method: http.MethodDelete, path: "/v1/studio/projects/p/chapters/c", respBody: `{"status":"ok"}`,
		},
		{
			name: "studio.projects.chapters.convert",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Chapters.Convert(ctx, "p", "c")
			},
			method: http.MethodPost, path: "/v1/studio/projects/p/chapters/c/convert", respBody: `{"status":"ok"}`,
		},
		{
			name: "studio.projects.chapters.snapshots.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Chapters.Snapshots.List(ctx, "p", "c")
			},
			method: http.MethodGet, path: "/v1/studio/projects/p/chapters/c/snapshots",
			respBody: `{"snapshots":[{"chapter_snapshot_id":"s","project_id":"p","chapter_id":"c","created_at_unix":1,"name":"n"}]}`,
		},
		{
			name: "studio.projects.chapters.snapshots.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Chapters.Snapshots.Get(ctx, "p", "c", "s")
			},
			method: http.MethodGet, path: "/v1/studio/projects/p/chapters/c/snapshots/s",
			respBody: `{"chapter_snapshot_id":"s","character_alignments":[{"characters":["a"],"character_start_times_seconds":[0],"character_end_times_seconds":[1]}]}`,
			check: func(t *testing.T, r any) {
				if s := r.(*ChapterSnapshotExtended); s.ChapterSnapshotID != "s" || s.CharacterAlignments[0].Characters[0] != "a" {
					t.Errorf("snapshot = %+v", s)
				}
			},
		},
		{
			name: "studio.projects.chapters.snapshots.stream",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.Studio.Projects.Chapters.Snapshots.Stream(ctx, "p", "c", "s", &SnapshotStreamRequest{ConvertToMpeg: Ptr(true)}))
			},
			method: http.MethodPost, path: "/v1/studio/projects/p/chapters/c/snapshots/s/stream", accept: acceptAny,
			body: `{"convert_to_mpeg":true}`, respType: "audio/mpeg", respBody: "aud", check: checkString("aud"),
		},
		{
			name: "studio.projects.content.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Content.Update(ctx, "p", &ProjectContentUpdateRequest{FromURL: "https://x", AutoConvert: Ptr(false)})
			},
			method: http.MethodPost, path: "/v1/studio/projects/p/content",
			form: map[string][]string{"from_url": {"https://x"}, "auto_convert": {"false"}}, respBody: `{"project":{"project_id":"p"}}`,
		},
		{
			name: "studio.projects.pronunciationDictionaries.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.PronunciationDictionaries.Create(ctx, "p", &ProjectPronunciationDictionariesRequest{PronunciationDictionaryLocators: []PronunciationDictionaryLocator{{PronunciationDictionaryID: "d", VersionID: "v"}}})
			},
			method: http.MethodPost, path: "/v1/studio/projects/p/pronunciation-dictionaries",
			body: `{"pronunciation_dictionary_locators":[{"pronunciation_dictionary_id":"d","version_id":"v"}]}`, respBody: `{"status":"ok"}`,
		},
		{
			name:   "studio.projects.snapshots.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Studio.Projects.Snapshots.List(ctx, "p") },
			method: http.MethodGet, path: "/v1/studio/projects/p/snapshots", respBody: `{"snapshots":[]}`,
		},
		{
			name: "studio.projects.snapshots.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Studio.Projects.Snapshots.Get(ctx, "p", "s")
			},
			method: http.MethodGet, path: "/v1/studio/projects/p/snapshots/s",
			respBody: `{"project_snapshot_id":"s","audio_duration_secs":12.5,"character_alignments":[]}`,
			check: func(t *testing.T, r any) {
				if s := r.(*ProjectSnapshotExtended); s.ProjectSnapshotID != "s" || s.AudioDurationSecs != 12.5 {
					t.Errorf("snapshot = %+v", s)
				}
			},
		},
		{
			name: "studio.projects.snapshots.stream",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.Studio.Projects.Snapshots.Stream(ctx, "p", "s", nil))
			},
			method: http.MethodPost, path: "/v1/studio/projects/p/snapshots/s/stream", accept: acceptAny,
			body: `{}`, respType: "audio/mpeg", respBody: "aud", check: checkString("aud"),
		},
		{
			name: "studio.projects.snapshots.streamArchive",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.Studio.Projects.Snapshots.StreamArchive(ctx, "p", "s"))
			},
			method: http.MethodPost, path: "/v1/studio/projects/p/snapshots/s/archive", accept: acceptAny,
			respType: "application/zip", respBody: "PK", check: checkString("PK"),
		},
	}
}

func TestDubbingTranscriptJSONVariant(t *testing.T) {
	client, _ := newTestServer(t, respond(http.StatusOK, contentTypeJSON, `{"language":"es","utterances":[{"text":"hola"}]}`))
	tr, err := client.Dubbing.Transcript.GetTranscriptForDub(context.Background(), "d", "es", TranscriptFormatJSON)
	if err != nil || tr.JSON == nil || tr.JSON.Language != "es" || tr.Text != "" {
		t.Fatalf("transcript = %+v, err = %v", tr, err)
	}
	client, _ = newTestServer(t, respond(http.StatusOK, contentTypeJSON, `"WEBVTT\n"`))
	if tr, err = client.Dubbing.Transcript.GetTranscriptForDub(context.Background(), "d", "es", TranscriptFormatWebVTT); err != nil || tr.Text != "WEBVTT\n" {
		t.Fatalf("transcript = %+v, err = %v", tr, err)
	}
}

func TestPodcastSourcesMultiple(t *testing.T) {
	b, err := PodcastSources{TextSource("a"), URLSource("u")}.MarshalJSON()
	if err != nil || string(b) != `[{"type":"text","text":"a"},{"type":"url","url":"u"}]` {
		t.Fatalf("sources = %s, %v", b, err)
	}
	b, _ = BulletinPodcast("h").MarshalJSON()
	if string(b) != `{"bulletin":{"host_voice_id":"h"},"type":"bulletin"}` {
		t.Errorf("mode = %s", b)
	}
}

func TestDubbingListAll(t *testing.T) {
	client, ts := newTestServer(t, sequence(
		respond(http.StatusOK, contentTypeJSON, `{"projects":[{"project_id":"a"}],"next_cursor":"c"}`),
		respond(http.StatusOK, contentTypeJSON, `{"projects":[{"project_id":"b"}]}`),
	))
	var ids []string
	for p, err := range client.Dubbing.Project.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, p.ProjectID)
	}
	if len(ids) != 2 || ts.last(t).RawQuery != "cursor=c" {
		t.Errorf("ids = %v", ids)
	}
}
