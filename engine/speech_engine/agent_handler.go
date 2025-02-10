package engine

import (
	"encoding/json"
	"fmt"

	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/speech"
)

func agentHandler(r *Turn, next Handler) Handler {
	return func(msg *Message) error {
		payload := msg.Payload.([]byte)
		if msg.Interrupted() {
			return nil
		}
		var vv speech.AgentPayload
		json.Unmarshal(payload, &vv)
		// new simple agent
		agt := agent.FunAgent{}
		result := agt.Run(r.Context())

		resultx, ok := result.(*llm.Stream)
		if !ok {
			return fmt.Errorf("can not convert to llm stream")
		}
		go func() {
			first := true
			for chunk := resultx.Next(); chunk != nil; chunk = resultx.Next() {
				payload := speech.ResponsePayload{Text: chunk.Delta.Message.Content(), First: first}
				raw, _ := json.Marshal(payload)
				next(r.engine.newMessage(raw).WithTurnID(msg.TurnID))
				first = false
			}
			payload := speech.ResponsePayload{Text: "", Final: true}
			raw, _ := json.Marshal(payload)
			next(r.engine.newMessage(raw).WithTurnID(msg.TurnID))
		}()

		return nil
	}

}
