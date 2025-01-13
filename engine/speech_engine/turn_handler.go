package engine

import (
	"encoding/json"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/speech"
)

// turnHandler ...
func turnHandler(r *Engine, next Handler) Handler {
	return func(msg *Message) error {
		payload := msg.Payload.([]byte)
		log.InfoContextf(r.Context(), "turn message: %s", string(payload))
		if msg.Interrupted() {
			return nil
		}
		var vv speech.TurnPayload
		json.Unmarshal(payload, &vv)

		return next(msg)
	}
}
