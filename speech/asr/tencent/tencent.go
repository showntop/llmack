package tencent

import (
	"context"
	"time"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/speech"
	"github.com/showntop/tencentcloud-speech-sdk-go/asr"
	"github.com/showntop/tencentcloud-speech-sdk-go/common"
)

// ASR ...
type ASR struct {
	options    Options
	recognizer *asr.SpeechRecognizer
}

// NewASR 新建一个ASR
func NewASR(opts ...Option) speech.ASR {
	t := &ASR{}

	for i := 0; i < len(opts); i++ {
		opts[i](&t.options)
	}

	if t.options.listener == nil {
		t.options.listener = &DefaultListener{
			t:  t,
			ID: 1,
		}
	}

	EngineModelType := "16k_zh"
	credential := common.NewCredential(t.options.SecretID, t.options.SecretKey)
	recognizer := asr.NewSpeechRecognizer(t.options.AppID, credential, EngineModelType, t.options.listener)

	recognizer.VoiceFormat = asr.AudioFormatPCM
	// recognizer.VoiceFormat = asr.AudioFormatWav
	// recognizer.NeedVad = 1
	recognizer.NoiseThreshold = 0.8

	// recognizer.VoiceFormat = asr.AudioFormatWav
	if err := recognizer.Start(); err != nil {
		panic(err)
	}
	log.InfoContextf(context.Background(), "asr engine start success")
	t.recognizer = recognizer
	// fmt.Println(t.recognizer)
	return t
}

// Input 异步实现语音识别，转录为文字
func (t *ASR) Input(content []byte) error {
	if len(content) == 0 {
		return nil
	}

	startTime := time.Now()
	// 置换 chunk 状态
	chunkSizePerDuration := float64(48000) * 2.00 / float64(time.Second)
	durations := time.Duration(float64(len(content)) / chunkSizePerDuration)
	if err := t.recognizer.Write(content); err != nil {
		panic(err)
	}
	endTime := time.Now()

	x := durations - endTime.Sub(startTime)
	// fmt.Printf("size:%d, chunk: %f durations: %s, sleep x=%s \n", len(content), chunkSizePerDuration, durations, x)
	if x > 0 {
		time.Sleep(x)
	}
	return nil
}

// Recognize 实现语音识别，转录为文字
func (t *ASR) Recognize(content []byte) (string, error) {
	panic("implement me")
}

// Close 关闭
func (t *ASR) Close() error {
	return t.recognizer.Stop()
}
