package aliyun

import (
	"fmt"
	"os"

	"github.com/showntop/llmack/speech"
)

// SpeechWsSynthesisListener is the listener of
type SpeechWsSynthesisListener interface {
	OnSynthesizeStarted(*speech.TTSResult)
	OnSentenceEnded(*speech.TTSResult)
	OnSynthesized(*speech.TTSResult)
	OnSentenceStarted(*speech.TTSResult)
	OnSynthesizeCompleted(*speech.TTSResult, error)
	OnSynthesizeFailed(*speech.TTSResult, error)
	OnAudioResult([]byte)
}

type listener struct {
	ff *os.File
	t  *TTS
}

func newListener(t *TTS) *listener {
	f, _ := os.Create("tts.wav")
	return &listener{
		ff: f,
		t:  t,
	}
}

func (l *listener) OnAudioResult(data []byte) {
	l.ff.Write(data)
	l.t.result <- data
}
func (l *listener) OnSentenceEnded(r *speech.TTSResult) {
	fmt.Printf("tts sentence ended, result:%+v\n", r)
}
func (l *listener) OnSentenceStarted(r *speech.TTSResult) {
	fmt.Printf("tts sentence started, result:%+v\n", r)
}
func (l *listener) OnSynthesizeCompleted(_ *speech.TTSResult, _ error) {
	fmt.Printf("tts synthesize completed\n")
}
func (l *listener) OnSynthesizeStarted(_ *speech.TTSResult) {
	fmt.Println("tts synthesize started")
}
func (l *listener) OnSynthesized(_ *speech.TTSResult) {
	fmt.Println("tts sentence synthesized 有新的合成结果返回。")
}
func (l *listener) OnSynthesizeFailed(_ *speech.TTSResult, _ error) {
	fmt.Println("tts OnSynthesizeFailed")
}
