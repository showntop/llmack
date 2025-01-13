package engine

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/showntop/llmack/speech"
	"github.com/showntop/llmack/speech/asr/tencent"

	// sasr "github.com/showntop/llmack/speech/asr"

	"github.com/showntop/tencentcloud-speech-sdk-go/asr"
)

// asrHandler ...
func asrHandler(r *Engine, next Handler) Handler {
	ff, _ := os.Create("asr.wav")
	wk := NewAsrWorker(r, next)
	go func() {
		for {
			select {
			case <-r.Context().Done():
				wk.asrH.Close()
				return
			}
		}
	}()

	first := true
	return func(msg *Message) error {
		chunk, _ := msg.Payload.([]byte)
		if first {
			// ff.Write(PcmToWav(chunk, 16000, 1))
			first = false
		}
		ff.Write(chunk)
		return wk.Input(chunk)
	}
}

// AsrWorker ...
type AsrWorker struct {
	r    *Engine
	asrH speech.ASR
	q    chan []byte

	handle Handler
}

// NewAsrWorker ...
func NewAsrWorker(r *Engine, handle Handler) *AsrWorker {
	worker := &AsrWorker{r: r}
	worker.handle = handle
	worker.q = make(chan []byte, 1000)
	go worker.LoopHandle()
	return worker
}

// Input ...
func (h *AsrWorker) Input(chunk []byte) error {
	h.q <- chunk
	return nil
}

// LoopHandle ...
func (h *AsrWorker) LoopHandle() {
	var xxxx map[string]struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
		AppID     int64  `json:"app_id"`
	}
	xxxx = make(map[string]struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
		AppID     int64  `json:"app_id"`
	})
	config := xxxx["tencent"]

	asrh := tencent.NewASR(
		tencent.WithAppID(fmt.Sprintf("%d", config.AppID)),
		tencent.WithSecretID(config.SecretID),
		tencent.WithSecretKey(config.SecretKey),
		tencent.WithListener(&AsrListener{h: h}))
	// for mock
	// asrh := sasr.NewMockASR(func(s string, b bool) error {
	// 	record := speech.TranscriptionPayload{Text: s, Final: b}
	// 	payload, _ := json.Marshal(&record)
	// 	worker.handle(NewMessage("", payload))
	// 	return nil
	// })
	h.asrH = asrh

	for chunk := range h.q {
		// h.Handle(chunk)
		h.asrH.Input(chunk)
	}
}

// Handle ...
// func (h *AsrWorker) Handle(chunk []byte) error {
// 	return h.asrH.Input(chunk)
// }

// AsrListener implementation of SpeechRecognitionListener
type AsrListener struct {
	h *AsrWorker
}

// OnRecognitionStart implementation of SpeechRecognitionListener
func (listener *AsrListener) OnRecognitionStart(response *asr.SpeechRecognitionResponse) {
}

// OnSentenceBegin implementation of SpeechRecognitionListener
func (listener *AsrListener) OnSentenceBegin(response *asr.SpeechRecognitionResponse) {
}

// OnRecognitionResultChange implementation of SpeechRecognitionListener
func (listener *AsrListener) OnRecognitionResultChange(response *asr.SpeechRecognitionResponse) {
	record := speech.TranscriptionPayload{Text: response.Result.VoiceTextStr, Final: false}
	payload, _ := json.Marshal(&record)
	listener.h.handle(listener.h.r.newMessage(payload))
}

// OnSentenceEnd implementation of SpeechRecognitionListener
func (listener *AsrListener) OnSentenceEnd(response *asr.SpeechRecognitionResponse) {
	record := speech.TranscriptionPayload{Text: response.Result.VoiceTextStr, Final: true}
	payload, _ := json.Marshal(&record)
	listener.h.handle(listener.h.r.newMessage(payload))
	// fmt.Printf("%s|%s|OnSentenceEnd: %v\n", time.Now().Format("2006-01-02 15:04:05"), response.VoiceID, response)
}

// OnRecognitionComplete implementation of SpeechRecognitionListener
func (listener *AsrListener) OnRecognitionComplete(response *asr.SpeechRecognitionResponse) {
}

// OnFail implementation of SpeechRecognitionListener
func (listener *AsrListener) OnFail(response *asr.SpeechRecognitionResponse, err error) {
}
