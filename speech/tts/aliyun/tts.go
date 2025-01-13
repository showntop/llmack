package aliyun

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/showntop/llmack/speech"
	"github.com/showntop/llmack/speech/tts"
)

// TTS ...
type TTS struct {
	tts.TTS

	options   Options
	sessionID string
	listener  SpeechWsSynthesisListener
	client    *WsClient
	result    chan []byte
}

// NewTTS ...
func NewTTS(opts ...Option) (speech.StreamTTS, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return newTTS(options)
}

func newTTS(options Options) (*TTS, error) {
	t := &TTS{options: options}
	t.listener = newListener(t)

	t.result = make(chan []byte, 1000)
	cli, err := NewWsClient(options.Token)
	if err != nil {
		return nil, err
	}
	t.client = cli
	return t, nil
}

// Start ...
func (t *TTS) Start() error {

	t.sessionID = Get32UUID()
	fmt.Println("tts start ", t.sessionID)
	payload := map[string]any{
		"enable_subtitle": true,
		"format":          "wav",
		"pitch_rate":      0,
		"platform":        "javascript",
		"sample_rate":     16000,
		"speech_rate":     0,
		"voice":           "zhixiaoxia",
		"volume":          100,
	}
	request := Request{
		Header:  newHeader(t.sessionID, "StartSynthesis"),
		Payload: payload,
	}

	if err := t.client.WriteJSON(request); err != nil {
		return err
	}

	// wait for synthesis started
	_, resp, err := t.client.ws.ReadMessage()
	if err != nil {
		return err
	}
	var vv Response
	json.Unmarshal(resp, &vv)
	if vv.Header.Name == "SynthesisStarted" {
		t.listener.OnSynthesizeStarted(&speech.TTSResult{SessionID: vv.Header.TaskID})
	}
	go t.listen()
	// log.InfoContextf(context.Background(), "sessionID:%s, request:%+v, response:%+v, err:%v", c.sessionID, request, resp, err)
	return nil
}

// StreamResult ...
func (t *TTS) StreamResult() chan []byte {
	return t.result
}

// Synthesize ...
func (t *TTS) Synthesize(data string) (chan []byte, error) {
	panic("implement me")
}

// Complete ...
func (t *TTS) Complete() error {
	request := Request{
		Header: newHeader(t.sessionID, "StopSynthesis"),
	}
	return t.client.WriteJSON(request)
}

// Prepare ...
func (t *TTS) Prepare() error {

	// stop put in
	if t.client.ws != nil {
		t.Terminate()
		// t.client.close()
	}

	// connect
	if err := t.client.reconnect(t.options.Token); err != nil {
		panic(err)
	}

	if err := t.Start(); err != nil {
		return err
	}
	// t.client.CloseHandler()

	return nil
}

// Input ...
func (t *TTS) Input(data string) error {

	if data == "" {
		return nil
	}
	header := map[string]any{
		"task_id":    t.sessionID,
		"message_id": strings.ReplaceAll(uuid.NewString(), "-", ""),
		"namespace":  "FlowingSpeechSynthesizer",
		"name":       "RunSynthesis",
		"appkey":     "N3CyrsLkHXD2yD3c",
	}
	payload := map[string]any{"text": data}

	if err := t.client.WriteJSON(map[string]any{
		"header":  header,
		"payload": payload,
	}); err != nil {
		panic(err)
	}
	return nil
}

// Terminate ...
func (t *TTS) Terminate() error {
	fmt.Println("tts stop ", t.sessionID)
	request := Request{
		Header: newHeader(t.sessionID, "StopSynthesis"),
	}
	return t.client.WriteJSON(request)
}

// Close ...
func (t *TTS) Close() error {

	return t.client.close()
}

func (t *TTS) listen() error {
	for {
		// t.Lock()
		// defer t.Unlock() // TODO lock
		mtype, resp, err := t.client.ws.ReadMessage()
		if err != nil {
			return err
		}

		if mtype == websocket.TextMessage {
			var vv Response
			json.Unmarshal(resp, &vv)
			if "SentenceEnd" == vv.Header.Name {
				t.listener.OnSentenceEnded(&speech.TTSResult{SessionID: vv.Header.TaskID})
			}
			if "SentenceSynthesis" == vv.Header.Name {
				t.listener.OnSynthesized(&speech.TTSResult{SessionID: vv.Header.TaskID})
			}
			if "SentenceBegin" == vv.Header.Name {
				t.listener.OnSentenceStarted(&speech.TTSResult{SessionID: vv.Header.TaskID})
			}
			if "SynthesisCompleted" == vv.Header.Name {
				t.listener.OnSynthesizeCompleted(&speech.TTSResult{SessionID: vv.Header.TaskID}, nil)
			}
			if "TaskFailed" == vv.Header.Name {
				fmt.Println(string(resp))
				t.Close()
				t.listener.OnSynthesizeFailed(&speech.TTSResult{SessionID: vv.Header.TaskID}, nil)
			}
		} else { // audio
			t.listener.OnAudioResult(resp)
		}
	}
}

// Options ...
type Options struct {
	Token    string
	Listener SpeechWsSynthesisListener
}

// Option ...
type Option func(*Options)

// WithToken ...
func WithToken(token string) Option {
	return func(o *Options) {
		o.Token = token
	}
}

// WithListener ...
func WithListener(listener SpeechWsSynthesisListener) Option {
	return func(o *Options) {
		o.Listener = listener
	}
}
