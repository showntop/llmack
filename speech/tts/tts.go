package tts

import "context"

// TTS is the struct for text to speech
type TTS struct {
}

// newTTS creates a new TTS
func newTTS() *TTS {
	return &TTS{}
}

// Speak speaks the given text
func (t *TTS) Speak(text string) error {
	return nil
}

func (t *TTS) getCached() {
	data, _ := Cache(nil).Get(context.Background(), "")
	if len(data) <= 0 {
		return
	}
}
