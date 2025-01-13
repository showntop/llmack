package engine

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/showntop/llmack/speech"
)

func transcribeHandler(r *Engine, next Handler) Handler {
	return func(msg *Message) error {
		payload := msg.Payload.([]byte)
		if len(payload) == 0 {
			return nil
		}

		// 开场白中，或者bot在说话，人类说了某些词
		if r.preambleOnGoing.Load() { // 开场白不允许打断。
			return nil
		}

		// human input when bot is speaking, should we ignore it
		// if r.speakingRole == "bot" && r.ignoreUtterance() {
		if r.speakingRole == "bot" && ignoreUtterance(r.Context(), string(payload)) {
			return nil
		}
		if r.speakingRole == "bot" { // 当前bot在说话，人类插话
			return nil // 忽略内容，禁止打断
			// 广播：中断所有消息 TODO 只处理上一轮的就可以。
			for _, m := range r.messages {
				_ = m
				m.Interrupt() // 处理中断shi'jian
			}

			// 处理上一轮
			if len(r.Turns) > 0 {
				turn := r.Turns[r.turnID]
				turn.tts.Complete()
			}

			// 打断bot
			r.outbound.Reset()
		}

		r.speakingRole = "human"
		var vv speech.TranscriptionPayload
		json.Unmarshal(msg.Payload.([]byte), &vv)
		if !vv.Final { // human finish
			return nil
		}

		// catch up final 新的一轮回复
		r.speakingRole = "bot"
		r.turnID = uuid.New().String() //重置当前轮次ID
		turn := NewTurn(r, uuid.NewString())
		r.Turns[r.turnID] = turn

		raw, _ := json.Marshal(speech.AgentPayload{Transcription: vv, TurnID: r.turnID})
		return turn.Handle(r.newMessage(raw).WithTurnID(r.turnID))
	}
}

const (
	ignoreUtteranceThreshold = 3
)

var (
	backchannels = []string{
		"好的",
		"嗯",
		"哦",
		"ok",
		"好的",
		"嗯哼",
		"嗯",
		"嗯嗯",
		"嗯嗯嗯",
	}
)

var bcregex = regexp.MustCompile("[^\\w\\s]")

func ignoreUtterance(ctx context.Context, payload string) bool {
	numWords := len([]rune(string(payload)))
	if numWords < 3 {
		return true
	}

	for _, bc := range backchannels {
		//
		//    cleaned = re.sub("[^\w\s]", "", transcription.message).strip().lower()
		// return any(re.fullmatch(regex, cleaned) for regex in BACKCHANNEL_PATTERNS)

		cleaned := bcregex.ReplaceAllString(string(payload), "")
		cleaned = strings.TrimSpace(cleaned)
		cleaned = strings.ToLower(cleaned)
		if strings.Contains(cleaned, bc) {
			return true
		}
	}
	return false
}
