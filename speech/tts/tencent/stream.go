package tencent

import (
	"sync"

	mtts "github.com/showntop/llmack/speech/tts"
	"github.com/showntop/tencentcloud-speech-sdk-go/tts"
)

// StreamTTS ...
type StreamTTS struct {
	sync.Mutex

	mtts.TTS
	options   Options
	client    *tts.SpeechWsv2Synthesizer
	sessionID string
	buffer    chan []byte
}

func NewStreamTTS(opts ...Option) *StreamTTS {
	return &StreamTTS{}
}

func (s *StreamTTS) Complete() error {
	return nil
}

func (s *StreamTTS) Prepare() error {
	return nil
}

func (s *StreamTTS) Input(text string) error {
	return nil
}

func (s *StreamTTS) Synthesize() error {
	return nil
}

func (s *StreamTTS) Terminate() error {
	return nil
}

// StreamResult() chan []byte
func (s *StreamTTS) StreamResult() chan []byte {
	return nil
}
