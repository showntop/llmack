package engine

import (
	"context"
	"encoding/base64"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/speech"
	tencent2 "github.com/showntop/llmack/speech/tts/tencent"
)

// Engine TODO
type Engine struct {
	*engine.BotEngine
	messages []*Message

	turnID string // 当前轮次ID
	Turns  map[string]*Turn

	speakingRole string // human or bot

	preambleOnGoing atomic.Bool

	inbound  any
	outbound speech.Outbound

	// simple pipline -> graph
	pipline Handler
	// pipline []HandlerConstructor
}

// NewEngine TODO
func NewEngine(ctx context.Context, out speech.Outbound) *Engine {
	r := &Engine{}

	r.BotEngine = engine.NewBotEngine().WithContext(ctx)
	r.outbound = out
	r.Turns = make(map[string]*Turn)

	r.pipline = asrHandler(r, transcribeHandler(r, turnHandler(r, nil)))
	// r.pipline = []HandlerConstructor{asrHandler, transcribeHandler, agentHandler, ttsHandler, outHandler}
	// send preamble to outbound
	go r.sendPreamble(ctx)
	return r
}

// Destroy TODO
func (r *Engine) Destroy() error {
	// return r.BotEngine.Destroy()
	// TODO 回收资源
	r.outbound.Close()
	return nil
}

func (r *Engine) warmUpTTS() (speech.StreamTTS, error) {

	var xxxx map[string]struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
		AppID     int64  `json:"app_id"`
	}

	ttsConfig := xxxx["tencent"]

	xx := tencent2.NewStreamTTS(
		tencent2.WithSecretID(ttsConfig.SecretID),
		tencent2.WithSecretKey(ttsConfig.SecretKey),
		tencent2.WithAppID(ttsConfig.AppID),
	)
	// if err != nil {
	// 	return nil, err
	// }
	if err := xx.Prepare(); err != nil {
		return nil, err
	}
	return xx, nil
}

func (r *Engine) newMessage(payload any) *Message {
	msg := NewMessage(uuid.NewString(), payload)
	r.messages = append(r.messages, msg)
	return msg
}

func (r *Engine) transmitMessage(o *Message, payload any) *Message {
	msg := NewMessage(uuid.NewString(), payload)
	msg.SessionID = r.SessionID
	r.messages = append(r.messages, msg)
	return msg
}

// Stream ... return channel
func (r *Engine) Stream(ctx context.Context, input engine.Input) <-chan engine.Event {
	resultChan := make(chan engine.Event, 1)
	dst, _ := base64.StdEncoding.DecodeString(input.Query)
	msg := r.newMessage(dst)
	msg.ConversationID = "1"
	// msg.SessionID = uuid.NewString()
	// ctx = context.WithValue(ctx, "session_id", uuid.NewString())
	msg.Question = "question"
	msg.Answer = "answer"
	r.pipline(msg)
	return resultChan
}

// Invoke ... return channel
func (r *Engine) Invoke(ctx context.Context, input engine.Input) (any, error) {
	panic("can not use blocking invoke")
}

func (r *Engine) sendPreamble(ctx context.Context) error {
	r.preambleOnGoing.Store(true)
	time.Sleep(1 * time.Second)
	// 可打断
	r.Settings.Preamble = "您好，请问您是住家家政的负责人嘛？"
	if r.Settings.Preamble == "" {
		return nil
	}
	// new tts once
	// ttsh, err := tts.NewAliyunTTS()
	var xxxx map[string]struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
		AppID     int64  `json:"app_id"`
	}

	ttsConfig := xxxx["tencent"]

	ttsh, err := tencent2.NewTextTTS(
		tencent2.WithSecretID(ttsConfig.SecretID),
		tencent2.WithSecretKey(ttsConfig.SecretKey),
		tencent2.WithAppID(ttsConfig.AppID),
	)
	if err != nil {
		return err
	}
	result, err := ttsh.Synthesize(ctx, r.Settings.Preamble)
	if err != nil {
		return err
	}
	chunks, _ := base64.RawStdEncoding.DecodeString(result.Audio)
	if err := r.outbound.Write(chunks); err != nil {
		return err
	}
	r.preambleOnGoing.Store(false)

	return nil
}
