package engine

import (
	"encoding/json"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/speech"
)

// TTSWorker TODO
type TTSWorker struct {
	first    bool
	tts      speech.StreamTTS
	ttsTable map[string]speech.StreamTTS
	r        *Engine
	q        chan string
	done     chan struct{}
}

func ttsHandler(turn *Turn, next Handler) Handler {
	// wk := NewTTSWorker(t.tts)
	// wk.r = r
	// wk.ttsTable = make(map[string]speech.TTS)
	return func(msg *Message) error {
		log.InfoContextf(turn.engine.Context(), "tts message: %s %v", msg, msg.Interrupted())
		if msg.Interrupted() {
			return nil
		}
		var vv speech.ResponsePayload
		json.Unmarshal(msg.Payload.([]byte), &vv)

		if vv.Final { // 到达大模型最后一个chunk，本轮结束
			if err := turn.tts.Complete(); err != nil {
				return err
			}
			return nil
		}

		if !vv.First { // 中间chunk
			if err := turn.tts.Input(vv.Text); err != nil {
				return err
			}
			return nil
		}

		// first chunk
		// log.InfoContextf(msg.Context(), "tts message is first chunk: %v", firstChunk)
		// new message and send to next handler
		payload := &speech.TTSResultPayload{
			Stream: turn.tts.StreamResult(),
			Text:   vv.Text,
		}
		dst := turn.engine.newMessage(payload).WithTurnID(msg.TurnID).WithOnOut(msg.OnOut())
		return next(dst)
	}
}

// NewTTSWorker TODO
func NewTTSWorker(tts speech.StreamTTS) *TTSWorker {
	h := &TTSWorker{first: true}
	h.tts = tts
	h.q = make(chan string, 10)
	// h.result = make(chan []byte, 10)
	// go h.LoopHandle()

	return h
}

// StreamResult TODO
func (h *TTSWorker) StreamResult() chan []byte {
	result := make(chan []byte, 10)
	go func() {
		first := true
		for c := range h.tts.StreamResult() {
			if first {
				// c = PcmToWav(c, 1, 16000)
				first = false
			}
			result <- c
		}
		close(result)
	}()
	// 当前 tts 的 result
	return result
}

// Input TODO
// func (h *TTSWorker) Input(words string) error {
// 	// h.q <- words

// 	return nil
// }

// LoopHandle TODO
// func (h *TTSWorker) LoopHandle() {
// 	for chunk := range h.q {
// 		h.tts.Input(chunk)
// 	}
// }

// Stop TODO
func (h *TTSWorker) Stop() error {
	close(h.q)
	h.done <- struct{}{}
	return nil
}
