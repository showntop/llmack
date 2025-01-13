package tencent

import (
	"fmt"
	"time"

	"github.com/showntop/tencentcloud-speech-sdk-go/asr"
)

// DefaultListener implementation of SpeechRecognitionListener
type DefaultListener struct {
	t  *ASR
	ID int
}

// OnRecognitionStart implementation of SpeechRecognitionListener
func (listener *DefaultListener) OnRecognitionStart(response *asr.SpeechRecognitionResponse) {
	fmt.Printf("%s|%s|OnRecognitionStart\n", time.Now().Format("2006-01-02 15:04:05"), response.VoiceID)
}

// OnSentenceBegin implementation of SpeechRecognitionListener
func (listener *DefaultListener) OnSentenceBegin(response *asr.SpeechRecognitionResponse) {
	fmt.Printf("%s|%s|OnSentenceBegin: %v\n", time.Now().Format("2006-01-02 15:04:05"), response.VoiceID, response)
}

// OnRecognitionResultChange implementation of SpeechRecognitionListener
func (listener *DefaultListener) OnRecognitionResultChange(response *asr.SpeechRecognitionResponse) {
	// fmt.Printf("OnRecognitionResultChange|%s|: %+v\n", response.MessageID, response)
	fmt.Printf("%s|%s|OnRecognitionResultChange: %v\n", time.Now().Format("2006-01-02 15:04:05"), response.VoiceID, response)
}

// OnSentenceEnd implementation of SpeechRecognitionListener
func (listener *DefaultListener) OnSentenceEnd(response *asr.SpeechRecognitionResponse) {
	fmt.Printf("OnSentenceEnd|%s|: %+v\n", response.VoiceID, response)
}

// OnRecognitionComplete implementation of SpeechRecognitionListener
func (listener *DefaultListener) OnRecognitionComplete(response *asr.SpeechRecognitionResponse) {
	fmt.Printf("%s|%s|OnRecognitionComplete\n", time.Now().Format("2006-01-02 15:04:05"), response.VoiceID)
}

// OnFail implementation of SpeechRecognitionListener
func (listener *DefaultListener) OnFail(response *asr.SpeechRecognitionResponse, err error) {
	fmt.Printf("OnFail:%s|%s| %v\n", time.Now().Format("2006-01-02 15:04:05"), response.VoiceID, err)
}
