//go:build integration

package e2e

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/taigrr/elevenlabs/v2/convai"
	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/types"
)

var userUtterances = []string{
	"Hi! Could I get a medium oat milk latte, please?",
	"Do you have any pastries today?",
	"Great, I will take a croissant too. Thank you!",
}

const (
	agentAudioFormat = "pcm_16000"
	agentIdle        = 2 * time.Second
	turnTimeout      = 45 * time.Second
)

type segment struct {
	speaker string
	text    string
	pcm     []byte
}

type agentTurn struct {
	text        []string
	pcm         []byte
	firstAudio  time.Time
	lastAudio   time.Time
	gotResponse bool
}

func (t *agentTurn) done(now time.Time) bool {
	return t.gotResponse && len(t.pcm) > 0 && now.Sub(t.lastAudio) > agentIdle
}

// playbackEnd is when a real-time client would finish playing the turn.
func (t *agentTurn) playbackEnd() time.Time {
	d := time.Duration(float64(len(t.pcm)) / bytesPerSecond * float64(time.Second))
	return t.firstAudio.Add(d)
}

type recvResult struct {
	ev  *convai.ConversationEvent
	err error
}

func areaConversation(ctx context.Context, h *Harness) (string, error) {
	userPCM := make([][]byte, len(userUtterances))
	for i, text := range userUtterances {
		pcm, err := h.tts(ctx, text, types.OutputFormatPCM16000)
		if err != nil {
			return "", fmt.Errorf("tts user %d: %w", i+1, err)
		}
		userPCM[i] = pcm
		if err := h.pcmToMP3(ctx, pcm, fmt.Sprintf("convo_user_%02d.mp3", i+1)); err != nil {
			return "", err
		}
	}

	agentID, err := h.createAgent(ctx)
	if err != nil {
		return "", err
	}
	return h.converse(ctx, agentID, userPCM)
}

// createAgent creates the temporary barista agent. When the key may not
// create agents, ELEVENLABS_AGENT_ID selects an existing agent instead.
func (h *Harness) createAgent(ctx context.Context) (string, error) {
	created, err := h.Client.ConvAI.Agents.Create(ctx, &agents.CreateRequest{
		Name: ResourcePrefix + "barista-" + h.Stamp,
		ConversationConfig: agents.ConversationalConfig{
			ASR: &agents.ASRConfig{UserInputAudioFormat: agentAudioFormat},
			TTS: &agents.AgentTTSConfig{ModelID: "eleven_flash_v2", VoiceID: VoiceID, AgentOutputAudioFormat: agentAudioFormat},
			Agent: &agents.AgentConfig{
				FirstMessage: "Hi, welcome to Go Beans! What can I get you?",
				Language:     "en",
				Prompt: &agents.AgentPrompt{
					Prompt: "You are a friendly barista at a coffee shop called Go Beans. Always answer in exactly one short sentence.",
				},
			},
		},
	})
	if err != nil {
		if existing := os.Getenv("ELEVENLABS_AGENT_ID"); existing != "" && IsPlanLimit(err) {
			h.Logf("convai: cannot create agent, using ELEVENLABS_AGENT_ID")
			return existing, nil
		}
		return "", fmt.Errorf("create agent: %w", err)
	}
	agentID := created.AgentID
	h.Defer(func() {
		if err := h.Client.ConvAI.Agents.Delete(context.Background(), agentID, nil); err != nil {
			h.Logf("cleanup agent %s: %v", agentID, err)
		}
	})
	return agentID, nil
}

func (h *Harness) converse(ctx context.Context, agentID string, userPCM [][]byte) (string, error) {
	sess, err := h.Client.ConvAI.Connect(ctx, &convai.ConnectRequest{AgentID: agentID})
	if err != nil {
		return "", fmt.Errorf("connect: %w", err)
	}
	defer sess.Close()
	meta, err := sess.Recv(ctx)
	if err != nil {
		return "", fmt.Errorf("recv metadata: %w", err)
	}
	if m := meta.ConversationInitiationMetadata; meta.Type != convai.ConversationEventInitiationMetadata || m == nil {
		return "", fmt.Errorf("first event = %s, want metadata", meta.Type)
	} else if m.AgentOutputAudioFormat != agentAudioFormat || m.UserInputAudioFormat != agentAudioFormat {
		return "", fmt.Errorf("%w: agent formats in=%s out=%s, harness needs %s", ErrSkip, m.UserInputAudioFormat, m.AgentOutputAudioFormat, agentAudioFormat)
	}

	sessCtx, stop := context.WithCancel(ctx)
	defer stop()

	events := make(chan recvResult, 256)
	go func() {
		defer close(events)
		for {
			ev, err := sess.Recv(sessCtx)
			select {
			case events <- recvResult{ev, err}:
			case <-sessCtx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()

	// The microphone: streams queued utterances in real time and silence otherwise.
	utterances := make(chan []byte)
	sendErr := make(chan error, 1)
	go func() {
		quiet := make([]byte, chunkBytes)
		ticker := time.NewTicker(chunkInterval)
		defer ticker.Stop()
		var pending []byte
		for {
			select {
			case <-sessCtx.Done():
				sendErr <- nil
				return
			case u := <-utterances:
				pending = u
				continue
			case <-ticker.C:
			}
			chunk := quiet
			if len(pending) > 0 {
				n := min(chunkBytes, len(pending))
				chunk, pending = pending[:n], pending[n:]
			}
			if err := sess.SendUserAudio(sessCtx, chunk); err != nil {
				if sessCtx.Err() == nil {
					sendErr <- err
				} else {
					sendErr <- nil
				}
				return
			}
		}
	}()

	var (
		segments     []segment
		transcript   []string
		agentReplies int
	)
	addAgent := func(t *agentTurn) error {
		agentReplies++
		n := agentReplies
		text := strings.Join(t.text, " ")
		if err := h.pcmToMP3(ctx, t.pcm, fmt.Sprintf("convo_agent_%02d.mp3", n)); err != nil {
			return err
		}
		segments = append(segments, segment{"agent", text, t.pcm})
		transcript = append(transcript, fmt.Sprintf("[agent %02d] %s", n, text))
		return nil
	}

	// waitAgent consumes events until the agent finished a turn. userIdx is
	// the user turn being answered, or -1 for the first message.
	waitAgent := func(userIdx int) (*agentTurn, error) {
		turn := &agentTurn{}
		gotUser := userIdx < 0
		deadline := time.After(turnTimeout)
		tick := time.NewTicker(200 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-deadline:
				return nil, fmt.Errorf("turn %d timed out (user transcript=%v, agent response=%v, audio=%d bytes)", userIdx+1, gotUser, turn.gotResponse, len(turn.pcm))
			case err := <-sendErr:
				return nil, fmt.Errorf("send audio: %w", err)
			case <-tick.C:
				if gotUser && turn.done(time.Now()) {
					return turn, nil
				}
			case r, ok := <-events:
				if !ok {
					return nil, errors.New("session closed")
				}
				if r.err != nil {
					return nil, fmt.Errorf("recv: %w", r.err)
				}
				ev := r.ev
				switch ev.Type {
				case convai.ConversationEventUserTranscript:
					if userIdx >= 0 && ev.UserTranscript != nil && strings.TrimSpace(ev.UserTranscript.UserTranscript) != "" {
						gotUser = true
						text := ev.UserTranscript.UserTranscript
						segments[len(segments)-1].text = text
						transcript = append(transcript, fmt.Sprintf("[user %02d] %s (sent: %q)", userIdx+1, text, userUtterances[userIdx]))
					}
				case convai.ConversationEventAudio:
					if ev.Audio != nil && len(ev.Audio.Audio) > 0 {
						now := time.Now()
						if turn.firstAudio.IsZero() {
							turn.firstAudio = now
						}
						turn.lastAudio = now
						turn.pcm = append(turn.pcm, ev.Audio.Audio...)
					}
				case convai.ConversationEventAgentResponse:
					turn.gotResponse = true
					turn.text = append(turn.text, ev.AgentResponse.AgentResponse)
				case convai.ConversationEventAgentResponseCorrection:
					if c := ev.AgentResponseCorrection; c != nil && len(turn.text) > 0 {
						turn.text[len(turn.text)-1] = c.CorrectedAgentResponse
					}
				case convai.ConversationEventInterruption:
					h.Logf("convai: interruption during turn %d", userIdx+1)
				}
			}
		}
	}

	first, err := waitAgent(-1)
	if err != nil {
		return "", fmt.Errorf("first message: %w", err)
	}
	if err := addAgent(first); err != nil {
		return "", err
	}
	last := first
	for i, pcm := range userPCM {
		time.Sleep(time.Until(last.playbackEnd().Add(500 * time.Millisecond)))
		segments = append(segments, segment{"user", userUtterances[i], pcm})
		utterances <- pcm
		turn, err := waitAgent(i)
		if err != nil {
			return "", err
		}
		if err := addAgent(turn); err != nil {
			return "", err
		}
		last = turn
	}
	conversationID := sess.ConversationID()
	stop()
	_ = sess.Close()

	gap := silence(gapSilence)
	var full []byte
	for i, s := range segments {
		if i > 0 {
			full = append(full, gap...)
		}
		full = append(full, s.pcm...)
	}
	if err := h.pcmToMP3(ctx, full, "convo_full.mp3"); err != nil {
		return "", err
	}
	header := fmt.Sprintf("conversation %s with agent %s\n\n", conversationID, agentID)
	if err := os.WriteFile(h.path("transcript.txt"), []byte(header+strings.Join(transcript, "\n")+"\n"), 0o644); err != nil {
		return "", err
	}
	for i := range userUtterances {
		if overlap(userUtterances[i], segments[2*i+1].text) < 0.5 {
			return "", fmt.Errorf("user turn %d transcribed as %q", i+1, segments[2*i+1].text)
		}
	}
	return fmt.Sprintf("conversation %s: %d user turns, %d agent replies", conversationID, len(userUtterances), agentReplies), nil
}
