package types_test

import (
	"io"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/v2/types"
)

type chunk struct {
	AudioBase64 string `json:"audio_base64"`
	Alignment   *struct {
		Characters []string `json:"characters"`
	} `json:"alignment"`
}

func TestStreamDecoding(t *testing.T) {
	body := "{\"audio_base64\":\"AA==\"}\n\n{\"audio_base64\":\"AQ==\",\"alignment\":{\"characters\":[\"h\"],\"character_start_times_seconds\":[0],\"character_end_times_seconds\":[0.1]}}"
	stream := types.NewStream[chunk](io.NopCloser(strings.NewReader(body)))
	var got []string
	for c, err := range stream.All() {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, c.AudioBase64)
	}
	if strings.Join(got, ",") != "AA==,AQ==" {
		t.Errorf("chunks = %v", got)
	}

	bad := types.NewStream[chunk](io.NopCloser(strings.NewReader("{\"audio_base64\":\"AA==\"}\nnot-json\n")))
	if !bad.Next() || bad.Current().AudioBase64 != "AA==" {
		t.Fatal("first message missing")
	}
	if bad.Next() || bad.Err() == nil {
		t.Error("expected decode error")
	}
	if bad.Next() {
		t.Error("Next after error should be false")
	}
}
