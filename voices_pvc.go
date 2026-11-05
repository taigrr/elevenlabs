package elevenlabs

import (
	"context"
	"net/http"
)

// VoicesPVCService groups professional voice cloning endpoints.
type VoicesPVCService struct {
	cfg *config

	// Samples groups PVC sample endpoints.
	Samples *PVCSamplesService
	// Verification groups PVC verification endpoints.
	Verification *PVCVerificationService
}

func newVoicesPVCService(cfg *config) *VoicesPVCService {
	return &VoicesPVCService{
		cfg: cfg,
		Samples: &PVCSamplesService{
			cfg:      cfg,
			Audio:    &PVCSamplesAudioService{cfg: cfg},
			Speakers: &PVCSamplesSpeakersService{cfg: cfg, Audio: &PVCSpeakersAudioService{cfg: cfg}},
			Waveform: &PVCSamplesWaveformService{cfg: cfg},
		},
		Verification: &PVCVerificationService{cfg: cfg, Captcha: &PVCCaptchaService{cfg: cfg}},
	}
}

// PVCCreateRequest is the body of VoicesPVCService.Create.
type PVCCreateRequest struct {
	// Name is the voice's display name. Required.
	Name string `json:"name"`
	// Language is the language of the samples. Required.
	Language string `json:"language"`
	// Description describes the voice.
	Description string `json:"description,omitempty"`
	// Labels are key/value labels (language, accent, gender, age).
	Labels map[string]string `json:"labels,omitempty"`
}

// PVCUpdateRequest is the body of VoicesPVCService.Update. All fields are optional.
type PVCUpdateRequest struct {
	Name        string            `json:"name,omitempty"`
	Language    string            `json:"language,omitempty"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// PVCTrainRequest is the body of VoicesPVCService.Train.
type PVCTrainRequest struct {
	// ModelID is the model to train for.
	ModelID string `json:"model_id,omitempty"`
}

// Create creates a new professional voice clone with no samples.
func (s *VoicesPVCService) Create(ctx context.Context, req *PVCCreateRequest, opts ...Option) (*VoiceIDResponse, error) {
	return doValue[VoiceIDResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/voices/pvc", body: orEmpty(req), opts: opts})
}

// Update edits a professional voice clone's metadata.
func (s *VoicesPVCService) Update(ctx context.Context, voiceID string, req *PVCUpdateRequest, opts ...Option) (*VoiceIDResponse, error) {
	return doValue[VoiceIDResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/pvc/%s", voiceID), body: orEmpty(req), opts: opts})
}

// Train starts training a professional voice clone.
func (s *VoicesPVCService) Train(ctx context.Context, voiceID string, req *PVCTrainRequest, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/pvc/%s/train", voiceID), body: orEmpty(req), opts: opts})
}

// PVCSamplesService groups PVC sample endpoints.
type PVCSamplesService struct {
	cfg *config

	// Audio groups sample audio previews.
	Audio *PVCSamplesAudioService
	// Speakers groups speaker separation endpoints.
	Speakers *PVCSamplesSpeakersService
	// Waveform groups sample waveform endpoints.
	Waveform *PVCSamplesWaveformService
}

// PVCAddSamplesRequest is the body of PVCSamplesService.Create.
type PVCAddSamplesRequest struct {
	// Files are the audio samples. Required.
	Files []File `form:"files"`
	// RemoveBackgroundNoise runs audio isolation on the samples.
	RemoveBackgroundNoise *bool `form:"remove_background_noise"`
}

// PVCUpdateSampleRequest is the body of PVCSamplesService.Update.
type PVCUpdateSampleRequest struct {
	// RemoveBackgroundNoise runs audio isolation on the sample.
	RemoveBackgroundNoise *bool `json:"remove_background_noise,omitempty"`
	// SelectedSpeakerIDs are the speakers to train on; each call replaces the previous selection.
	SelectedSpeakerIDs []string `json:"selected_speaker_ids,omitempty"`
	// TrimStartTime is the training start offset in milliseconds.
	TrimStartTime *int64 `json:"trim_start_time,omitempty"`
	// TrimEndTime is the training end offset in milliseconds.
	TrimEndTime *int64 `json:"trim_end_time,omitempty"`
	// FileName renames the sample.
	FileName string `json:"file_name,omitempty"`
}

// Create adds audio samples to a professional voice clone.
func (s *PVCSamplesService) Create(ctx context.Context, voiceID string, req *PVCAddSamplesRequest, opts ...Option) ([]VoiceSample, error) {
	var out []VoiceSample
	err := doJSON(ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/pvc/%s/samples", voiceID), form: orEmpty(req), opts: opts}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Update edits a PVC sample (trimming, speaker selection, noise removal).
func (s *PVCSamplesService) Update(ctx context.Context, voiceID, sampleID string, req *PVCUpdateSampleRequest, opts ...Option) (*VoiceIDResponse, error) {
	return doValue[VoiceIDResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/pvc/%s/samples/%s", voiceID, sampleID), body: orEmpty(req), opts: opts})
}

// Delete removes a sample from a professional voice clone.
func (s *PVCSamplesService) Delete(ctx context.Context, voiceID, sampleID string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/voices/pvc/%s/samples/%s", voiceID, sampleID), opts: opts})
}

// PVCSamplesAudioService groups PVC sample audio endpoints.
type PVCSamplesAudioService struct{ cfg *config }

// PVCSampleAudioRequest holds the optional parameters of PVCSamplesAudioService.Get.
type PVCSampleAudioRequest struct {
	// RemoveBackgroundNoise returns the noise-isolated audio.
	RemoveBackgroundNoise *bool `query:"remove_background_noise"`
}

// VoiceSamplePreview is a base64-encoded audio preview of a sample.
type VoiceSamplePreview struct {
	AudioBase64  string   `json:"audio_base_64"`
	VoiceID      string   `json:"voice_id"`
	SampleID     string   `json:"sample_id"`
	MediaType    string   `json:"media_type"`
	DurationSecs *float64 `json:"duration_secs,omitempty"`
}

// Get returns a base64 preview of a PVC sample's audio. req may be nil.
func (s *PVCSamplesAudioService) Get(ctx context.Context, voiceID, sampleID string, req *PVCSampleAudioRequest, opts ...Option) (*VoiceSamplePreview, error) {
	return doValue[VoiceSamplePreview](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/pvc/%s/samples/%s/audio", voiceID, sampleID), query: req, opts: opts})
}

// PVCSamplesSpeakersService groups PVC speaker separation endpoints.
type PVCSamplesSpeakersService struct {
	cfg *config

	// Audio groups separated speaker audio endpoints.
	Audio *PVCSpeakersAudioService
}

// Get returns the speaker separation status and detected speakers of a sample.
func (s *PVCSamplesSpeakersService) Get(ctx context.Context, voiceID, sampleID string, opts ...Option) (*SpeakerSeparation, error) {
	return doValue[SpeakerSeparation](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/pvc/%s/samples/%s/speakers", voiceID, sampleID), opts: opts})
}

// Separate starts speaker separation for a sample.
func (s *PVCSamplesSpeakersService) Separate(ctx context.Context, voiceID, sampleID string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/pvc/%s/samples/%s/separate-speakers", voiceID, sampleID), opts: opts})
}

// PVCSpeakersAudioService groups separated speaker audio endpoints.
type PVCSpeakersAudioService struct{ cfg *config }

// SpeakerAudio is base64-encoded audio of one separated speaker.
type SpeakerAudio struct {
	AudioBase64  string  `json:"audio_base_64"`
	MediaType    string  `json:"media_type"`
	DurationSecs float64 `json:"duration_secs"`
}

// Get returns the isolated audio of one speaker in a sample.
func (s *PVCSpeakersAudioService) Get(ctx context.Context, voiceID, sampleID, speakerID string, opts ...Option) (*SpeakerAudio, error) {
	return doValue[SpeakerAudio](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/pvc/%s/samples/%s/speakers/%s/audio", voiceID, sampleID, speakerID), opts: opts})
}

// PVCSamplesWaveformService groups PVC sample waveform endpoints.
type PVCSamplesWaveformService struct{ cfg *config }

// VoiceSampleWaveform is the visual waveform of a sample.
type VoiceSampleWaveform struct {
	SampleID       string    `json:"sample_id"`
	VisualWaveform []float64 `json:"visual_waveform"`
}

// Get returns the visual waveform of a PVC sample.
func (s *PVCSamplesWaveformService) Get(ctx context.Context, voiceID, sampleID string, opts ...Option) (*VoiceSampleWaveform, error) {
	return doValue[VoiceSampleWaveform](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/pvc/%s/samples/%s/waveform", voiceID, sampleID), opts: opts})
}

// PVCVerificationService groups PVC verification endpoints.
type PVCVerificationService struct {
	cfg *config

	// Captcha groups captcha verification endpoints.
	Captcha *PVCCaptchaService
}

// PVCManualVerificationRequest is the body of PVCVerificationService.Request.
type PVCManualVerificationRequest struct {
	// Files are the verification documents. Required.
	Files []File `form:"files"`
	// ExtraText is additional context for the verification.
	ExtraText string `form:"extra_text"`
}

// Request requests manual verification of a professional voice clone.
func (s *PVCVerificationService) Request(ctx context.Context, voiceID string, req *PVCManualVerificationRequest, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/pvc/%s/verification", voiceID), form: orEmpty(req), opts: opts})
}

// PVCCaptchaService groups PVC captcha verification endpoints.
type PVCCaptchaService struct{ cfg *config }

// Get requests a verification captcha for a professional voice clone.
func (s *PVCCaptchaService) Get(ctx context.Context, voiceID string, opts ...Option) error {
	return doJSON(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/voices/pvc/%s/captcha", voiceID), opts: opts}, nil)
}

// PVCVerifyCaptchaRequest is the body of PVCCaptchaService.Verify.
type PVCVerifyCaptchaRequest struct {
	// Recording is the user's recording of the captcha text. Required.
	Recording File `form:"recording"`
}

// Verify submits a captcha recording to verify a professional voice clone.
func (s *PVCCaptchaService) Verify(ctx context.Context, voiceID string, req *PVCVerifyCaptchaRequest, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/voices/pvc/%s/captcha", voiceID), form: orEmpty(req), opts: opts})
}
