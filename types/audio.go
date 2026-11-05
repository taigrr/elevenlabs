package types

import "encoding/json"

// OutputFormat is an audio output format: codec_samplerate[_bitrate].
type OutputFormat string

// OutputFormat values.
const (
	OutputFormatALaw8000      OutputFormat = "alaw_8000"
	OutputFormatMP3_22050_32  OutputFormat = "mp3_22050_32"
	OutputFormatMP3_24000_48  OutputFormat = "mp3_24000_48"
	OutputFormatMP3_44100_32  OutputFormat = "mp3_44100_32"
	OutputFormatMP3_44100_64  OutputFormat = "mp3_44100_64"
	OutputFormatMP3_44100_96  OutputFormat = "mp3_44100_96"
	OutputFormatMP3_44100_128 OutputFormat = "mp3_44100_128"
	OutputFormatMP3_44100_192 OutputFormat = "mp3_44100_192"
	OutputFormatMP3_48000_128 OutputFormat = "mp3_48000_128"
	OutputFormatMP3_48000_192 OutputFormat = "mp3_48000_192"
	OutputFormatMP3_48000_240 OutputFormat = "mp3_48000_240"
	OutputFormatMP3_48000_320 OutputFormat = "mp3_48000_320"
	// OutputFormatAuto lets the server pick the format (music endpoints only).
	OutputFormatAuto          OutputFormat = "auto"
	OutputFormatOpus48000_32  OutputFormat = "opus_48000_32"
	OutputFormatOpus48000_64  OutputFormat = "opus_48000_64"
	OutputFormatOpus48000_96  OutputFormat = "opus_48000_96"
	OutputFormatOpus48000_128 OutputFormat = "opus_48000_128"
	OutputFormatOpus48000_192 OutputFormat = "opus_48000_192"
	OutputFormatPCM8000       OutputFormat = "pcm_8000"
	OutputFormatPCM16000      OutputFormat = "pcm_16000"
	OutputFormatPCM22050      OutputFormat = "pcm_22050"
	OutputFormatPCM24000      OutputFormat = "pcm_24000"
	OutputFormatPCM32000      OutputFormat = "pcm_32000"
	OutputFormatPCM44100      OutputFormat = "pcm_44100"
	OutputFormatPCM48000      OutputFormat = "pcm_48000"
	OutputFormatULaw8000      OutputFormat = "ulaw_8000"
	OutputFormatWAV8000       OutputFormat = "wav_8000"
	OutputFormatWAV16000      OutputFormat = "wav_16000"
	OutputFormatWAV22050      OutputFormat = "wav_22050"
	OutputFormatWAV24000      OutputFormat = "wav_24000"
	OutputFormatWAV32000      OutputFormat = "wav_32000"
	OutputFormatWAV44100      OutputFormat = "wav_44100"
	OutputFormatWAV48000      OutputFormat = "wav_48000"
)

// TextNormalization controls text normalization.
type TextNormalization string

// TextNormalization values.
const (
	TextNormalizationAuto TextNormalization = "auto"
	TextNormalizationOn   TextNormalization = "on"
	TextNormalizationOff  TextNormalization = "off"
	// TextNormalizationApplyEnglish is "on" assuming English text; only some endpoints (e.g. Audio Native) accept it.
	TextNormalizationApplyEnglish TextNormalization = "apply_english"
)

// PronunciationDictionaryLocator references a pronunciation dictionary version.
type PronunciationDictionaryLocator struct {
	PronunciationDictionaryID string `json:"pronunciation_dictionary_id"`
	VersionID                 string `json:"version_id,omitempty"`
}

// String returns the JSON encoding of l, the form used by multipart
// endpoints that take locators as JSON strings.
func (l PronunciationDictionaryLocator) String() string {
	b, _ := json.Marshal(l)
	return string(b)
}

// AudioEffects configures post-processing effects applied to generated audio.
type AudioEffects struct {
	FilterPresetID    string   `json:"filter_preset_id,omitempty"`
	Distance          *float64 `json:"distance,omitempty"`
	EnvironmentID     string   `json:"environment_id,omitempty"`
	BackgroundNoiseID string   `json:"background_noise_id,omitempty"`
	SendLevel         *float64 `json:"send_level,omitempty"`
	Pan               *float64 `json:"pan,omitempty"`
}

// SortDirection is a sort order.
type SortDirection string

// SortDirection values.
const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

// CharacterAlignment holds per-character timing information.
type CharacterAlignment struct {
	Characters                 []string  `json:"characters"`
	CharacterStartTimesSeconds []float64 `json:"character_start_times_seconds"`
	CharacterEndTimesSeconds   []float64 `json:"character_end_times_seconds"`
}

// VoiceSettings controls voice synthesis. All fields are optional.
type VoiceSettings struct {
	Stability       *float64 `json:"stability,omitempty"`
	UseSpeakerBoost *bool    `json:"use_speaker_boost,omitempty"`
	SimilarityBoost *float64 `json:"similarity_boost,omitempty"`
	Style           *float64 `json:"style,omitempty"`
	Speed           *float64 `json:"speed,omitempty"`
	CloneStrength   *float64 `json:"clone_strength,omitempty"`
}

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
