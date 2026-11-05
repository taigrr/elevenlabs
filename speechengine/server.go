package speechengine

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/taigrr/elevenlabs/v2/internal/ws"
	"github.com/taigrr/elevenlabs/v2/realtime"
)

// AuthHeader carries the JWT ElevenLabs signs when it connects
// to a speech engine WebSocket server.
const AuthHeader = "X-Elevenlabs-Speech-Engine-Authorization"

const (
	issuer  = "https://api.elevenlabs.io/convai/speech-engine"
	subject = "convai_speech_engine_upstream"
	leeway  = 60 * time.Second
)

var residencyKeySuffix = regexp.MustCompile(`_residency_[a-z0-9]+$`)

// Speech engine errors.
var (
	// ErrAuth is wrapped by every speech engine token
	// verification failure.
	ErrAuth = errors.New("elevenlabs: speech engine authorization failed")
	// ErrNoTranscript is returned by Session.SendResponse when the
	// context does not come from a Callbacks.OnTranscript call.
	ErrNoTranscript = errors.New("elevenlabs: response sent outside of a transcript handler")
)

// Claims are the verified claims of a speech engine JWT.
type Claims map[string]any

// VerifyToken verifies a speech engine JWT (optionally prefixed
// with "Bearer ") signed with the SHA-256 of apiKey, checking issuer, subject,
// exp and iat against now with a 60 second leeway. Errors wrap
// ErrAuth.
func VerifyToken(token, apiKey string, now time.Time) (Claims, error) {
	token = strings.TrimSpace(token)
	if len(token) >= 7 && strings.EqualFold(token[:7], "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: invalid jwt: expected 3 parts", ErrAuth)
	}
	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid jwt: failed to decode payload", ErrAuth)
	}
	var claims Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("%w: invalid jwt: failed to decode payload", ErrAuth)
	}
	secret := sha256.Sum256([]byte(residencyKeySuffix.ReplaceAllString(strings.TrimSpace(apiKey), "")))
	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(parts[0] + "." + parts[1]))
	sig, err := base64URLDecode(parts[2])
	if err != nil || !hmac.Equal(mac.Sum(nil), sig) {
		return nil, fmt.Errorf("%w: invalid jwt: signature mismatch; the server API key must match the one used to create the engine", ErrAuth)
	}
	if iss, _ := claims["iss"].(string); iss != issuer {
		return nil, fmt.Errorf("%w: invalid jwt: expected issuer %q, got %q", ErrAuth, issuer, claims["iss"])
	}
	if sub, _ := claims["sub"].(string); sub != subject {
		return nil, fmt.Errorf("%w: invalid jwt: expected subject %q, got %q", ErrAuth, subject, claims["sub"])
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("%w: invalid jwt: missing exp claim", ErrAuth)
	}
	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, fmt.Errorf("%w: invalid jwt: missing iat claim", ErrAuth)
	}
	leeway := leeway.Seconds()
	unix := float64(now.Unix())
	if exp+leeway < unix {
		return nil, fmt.Errorf("%w: invalid jwt: token has expired", ErrAuth)
	}
	if iat-leeway > unix {
		return nil, fmt.Errorf("%w: invalid jwt: iat is in the future", ErrAuth)
	}
	return claims, nil
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
}

// VerifyRequest checks that r is a genuine connection from the ElevenLabs
// Speech Engine API using the client's API key. Only needed when you accept
// the WebSocket yourself; Handler verifies automatically.
func (s *Service) VerifyRequest(r *http.Request) error {
	if s.cfg.APIKey == "" {
		return fmt.Errorf("%w: no api key configured on the client", ErrAuth)
	}
	token := r.Header.Get(AuthHeader)
	if token == "" {
		return fmt.Errorf("%w: missing %s header", ErrAuth, AuthHeader)
	}
	_, err := VerifyToken(token, s.cfg.APIKey, s.cfg.Now())
	return err
}

// TranscriptMessage is one turn of the conversation transcript.
type TranscriptMessage struct {
	// Role is "user" or "agent".
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Callbacks receive the events of a speech engine session. All
// fields are optional.
type Callbacks struct {
	// DisableAuth skips JWT verification of incoming connections. Insecure:
	// only use behind another layer such as an IP allowlist.
	DisableAuth bool
	// AcceptOptions customizes the WebSocket upgrade.
	AcceptOptions *websocket.AcceptOptions

	// OnInit is called once the conversation ID is known.
	OnInit func(s *Session, conversationID string)
	// OnTranscript is called in its own goroutine for each user transcript.
	// ctx is cancelled when a newer transcript arrives or the session ends;
	// pass it to SendResponse. A returned error that is not a context
	// cancellation is reported to OnError.
	OnTranscript func(ctx context.Context, s *Session, transcript []TranscriptMessage) error
	// OnClose is called when ElevenLabs cleanly ends the conversation.
	OnClose func(s *Session)
	// OnDisconnect is called when the WebSocket connection ends.
	OnDisconnect func(s *Session)
	// OnError is called on protocol, WebSocket and handler errors.
	OnError func(s *Session, err error)
}

// Handler returns an http.Handler that verifies incoming Speech Engine
// connections, upgrades them to WebSockets and runs a session per
// connection. Mount it on your mux at the path configured as the engine's
// ws_url.
func (s *Service) Handler(cb Callbacks) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cb.DisableAuth {
			if err := s.VerifyRequest(r); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		conn, err := websocket.Accept(w, r, cb.AcceptOptions)
		if err != nil {
			return
		}
		NewSession(conn, cb).Run(r.Context())
	})
}

type eventKey struct{}

type event struct {
	id    *int
	valid bool
}

// Session is one conversation with the ElevenLabs Speech Engine
// API over an accepted WebSocket connection.
type Session struct {
	conn *websocket.Conn
	cb   Callbacks

	writeMu sync.Mutex
	mu      sync.Mutex
	cancel  context.CancelFunc
	eventID *int
	convID  string
	closed  bool
	wg      sync.WaitGroup
}

// NewSession wraps an accepted WebSocket connection. Call Run
// to process messages. Only needed when you accept connections yourself.
func NewSession(conn *websocket.Conn, cb Callbacks) *Session {
	conn.SetReadLimit(ws.ReadLimit)
	return &Session{conn: conn, cb: cb}
}

// ConversationID returns the conversation ID, empty until initialized.
func (s *Session) ConversationID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.convID
}

// IsOpen reports whether the session has not been closed.
func (s *Session) IsOpen() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.closed
}

// Close cancels any in-flight transcript handler and closes the connection.
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.abortLocked()
	s.mu.Unlock()
	return s.conn.Close(websocket.StatusNormalClosure, "")
}

func (s *Session) abortLocked() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

func (s *Session) emitError(err error) {
	if s.cb.OnError != nil {
		s.cb.OnError(s, err)
	}
}

type incoming struct {
	Type           string              `json:"type"`
	ConversationID string              `json:"conversation_id"`
	UserTranscript []TranscriptMessage `json:"user_transcript"`
	EventID        *int                `json:"event_id"`
	Message        string              `json:"message"`
}

// Run reads messages until the connection ends or ctx is cancelled, then
// waits for running transcript handlers to return.
func (s *Session) Run(ctx context.Context) {
	defer s.wg.Wait()
	for {
		_, data, err := s.conn.Read(ctx)
		if err != nil {
			s.mu.Lock()
			wasClosed := s.closed
			s.closed = true
			s.abortLocked()
			s.mu.Unlock()
			if !wasClosed {
				_ = s.conn.CloseNow()
			}
			if s.cb.OnDisconnect != nil {
				s.cb.OnDisconnect(s)
			}
			return
		}
		var msg incoming
		if err := json.Unmarshal(data, &msg); err != nil {
			s.emitError(fmt.Errorf("%w: %w", realtime.ErrUnexpectedMessage, err))
			continue
		}
		s.handle(ctx, &msg)
	}
}

func (s *Session) handle(ctx context.Context, msg *incoming) {
	switch msg.Type {
	case "init":
		s.mu.Lock()
		s.convID = msg.ConversationID
		s.mu.Unlock()
		if s.cb.OnInit != nil {
			s.cb.OnInit(s, msg.ConversationID)
		}
	case "user_transcript":
		s.mu.Lock()
		if msg.EventID != nil && s.eventID != nil && *msg.EventID == *s.eventID && s.cancel != nil {
			s.mu.Unlock()
			return
		}
		s.abortLocked()
		turnCtx, cancel := context.WithCancel(context.WithValue(ctx, eventKey{}, event{id: msg.EventID, valid: true}))
		s.cancel = cancel
		s.eventID = msg.EventID
		s.mu.Unlock()
		if s.cb.OnTranscript == nil {
			return
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			err := s.cb.OnTranscript(turnCtx, s, msg.UserTranscript)
			if err != nil && !errors.Is(err, context.Canceled) {
				s.emitError(err)
			}
		}()
	case "ping":
		_ = s.send(ctx, map[string]string{"type": "pong"})
	case "close":
		s.mu.Lock()
		s.abortLocked()
		s.mu.Unlock()
		if s.cb.OnClose != nil {
			s.cb.OnClose(s)
		}
	case "error":
		s.emitError(&realtime.Error{MessageType: "error", Message: msg.Message})
	}
}

type agentResponse struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	EventID *int   `json:"event_id"`
	IsFinal bool   `json:"is_final"`
}

func (s *Session) send(ctx context.Context, v any) error {
	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return realtime.ErrSocketClosed
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if err := s.conn.Write(ctx, websocket.MessageText, b); err != nil {
		return fmt.Errorf("%w: %w", realtime.ErrSocketClosed, err)
	}
	return nil
}

func (s *Session) turn(ctx context.Context) (*int, error) {
	ev, ok := ctx.Value(eventKey{}).(event)
	if !ok || !ev.valid {
		return nil, ErrNoTranscript
	}
	return ev.id, ctx.Err()
}

func (s *Session) sendChunk(ctx context.Context, id *int, text string, final bool) error {
	return s.send(context.WithoutCancel(ctx), agentResponse{Type: "agent_response", Content: text, EventID: id, IsFinal: final})
}

// SendResponse sends a complete agent response for the transcript whose
// OnTranscript ctx is passed in.
func (s *Session) SendResponse(ctx context.Context, text string) error {
	id, err := s.turn(ctx)
	if err != nil {
		return err
	}
	if err := s.sendChunk(ctx, id, text, false); err != nil {
		return err
	}
	return s.sendChunk(ctx, id, "", true)
}

// SendResponseStream streams an agent response chunk by chunk, stopping
// without a final message when ctx is cancelled (the user interrupted).
// Use ExtractLLMText to turn raw LLM stream events into text.
func (s *Session) SendResponseStream(ctx context.Context, chunks iter.Seq2[string, error]) error {
	id, err := s.turn(ctx)
	if err != nil {
		return err
	}
	for text, err := range chunks {
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		if ctx.Err() != nil {
			return nil
		}
		if text == "" {
			continue
		}
		if err := s.sendChunk(ctx, id, text, false); err != nil {
			return err
		}
	}
	if ctx.Err() != nil {
		return nil
	}
	return s.sendChunk(ctx, id, "", true)
}

// ExtractLLMText extracts the text delta from a JSON-encoded LLM stream
// event: a JSON string, an OpenAI Responses "response.output_text.delta",
// an OpenAI Chat Completions chunk, an Anthropic "content_block_delta" or a
// Gemini candidate. It reports false when the event carries no text.
func ExtractLLMText(event []byte) (string, bool) {
	var str string
	if json.Unmarshal(event, &str) == nil {
		return str, true
	}
	var ev struct {
		Type    string          `json:"type"`
		Delta   json.RawMessage `json:"delta"`
		Choices []struct {
			Delta struct {
				Content *string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text *string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if json.Unmarshal(event, &ev) != nil {
		return "", false
	}
	if ev.Type == "response.output_text.delta" && json.Unmarshal(ev.Delta, &str) == nil {
		return str, true
	}
	if len(ev.Choices) > 0 && ev.Choices[0].Delta.Content != nil {
		return *ev.Choices[0].Delta.Content, true
	}
	if ev.Type == "content_block_delta" {
		var d struct {
			Type string  `json:"type"`
			Text *string `json:"text"`
		}
		if json.Unmarshal(ev.Delta, &d) == nil && d.Type == "text_delta" && d.Text != nil {
			return *d.Text, true
		}
	}
	if len(ev.Candidates) > 0 && len(ev.Candidates[0].Content.Parts) > 0 && ev.Candidates[0].Content.Parts[0].Text != nil {
		return *ev.Candidates[0].Content.Parts[0].Text, true
	}
	return "", false
}
