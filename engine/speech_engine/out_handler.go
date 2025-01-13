package engine

import (
	"fmt"
	"os"

	"github.com/showntop/llmack/speech"
)

func outHandler(r *Engine, _ Handler) Handler {
	q := make(chan *Message, 10)

	go func() {
		i := 1
		for msg := range q {
			file, _ := os.Create(fmt.Sprint(i) + "000.wav")
			i++
			payload := msg.Payload.(*speech.TTSResultPayload)
			for chunk := range payload.Stream {
				// file.Write([]byte(fmt.Sprintf("%x", md5.Sum(chunk))))
				file.Write(chunk)
				if msg.Interrupted() {
					break
				}
				if err := r.outbound.Write(chunk); err != nil {
					panic(err)
				}
			}
			file.Close()
		}
	}()

	return func(msg *Message) error {
		if msg.onOut != nil {
			msg.onOut(msg)
		}
		if msg.Interrupted() {
			return nil
		}
		q <- msg // sequential
		return nil
	}
}
