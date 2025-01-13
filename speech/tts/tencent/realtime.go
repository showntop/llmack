package tencent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/showntop/llmack/speech"
	mtts "github.com/showntop/llmack/speech/tts"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/common"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/tts"
)

// RealtimeTTS ...
type RealtimeTTS struct {
	mtts.TTS
	options Options
}

// NewRealtimeTTS ...
func NewRealtimeTTS(opts ...Option) (speech.RealtimeTTS, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	if options.secretID == "" || options.secretKey == "" || options.appID == 0 {
		return nil, errors.New("secretID, secretKey and appID are required")
	}

	tencent := &RealtimeTTS{}
	tencent.options = options
	return tencent, nil
}

// Synthesize ...
func (t *RealtimeTTS) Synthesize(ctx context.Context, text string) (*speech.TTSResult, error) {
	if text == "" {
		return nil, fmt.Errorf("text empty")
	}
	result := (speech.NewTTSResult())
	credential := common.NewCredential(t.options.secretID, t.options.secretKey)
	synth := tts.NewSpeechWsSynthesizer(t.options.appID, credential, &realtimeTTSListener{result: result})
	synth.VoiceType = 1001
	// synth.Codec = "mp3"
	// synth.Timestamp = time.Now().Unix()
	// synth.Expired = time.Now().Unix() + 120
	ssid := uuid.NewString()
	fmt.Println(ssid)
	synth.SessionId = ssid
	synth.Text = text
	synth.EnableSubtitle = true
	synth.Codec = "mp3"
	// synth.VoiceType =
	// synth.Volume =
	// synth.Speed = 0
	if err := synth.Synthesis(); err != nil {
		return nil, err
	}

	if err := synth.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

type realtimeTTSListener struct {
	sync.Mutex
	result    *speech.TTSResult
	SessionId string
	Index     int
}

func (l *realtimeTTSListener) OnSynthesisStart(r *tts.SpeechWsSynthesisResponse) {
	fmt.Printf("%s|OnSynthesisStart,sessionId:%s response: %s\n", time.Now().Format("2006-01-02 15:04:05"), l.SessionId, r.ToString())
}

func (l *realtimeTTSListener) OnSynthesisEnd(r *tts.SpeechWsSynthesisResponse) {
	// fileName := fmt.Sprintf("test.mp3")
	// tts.WriteFile(path.Join("./", fileName), l.Data.Audio)
	close(l.result.Data)
	// fmt.Printf("%s|OnSynthesisEnd,sessionId:%s response: %s\n", time.Now().Format("2006-01-02 15:04:05"), l.SessionId, r.ToString())
}

func (l *realtimeTTSListener) OnAudioResult(data []byte) {
	fmt.Printf("%s|OnAudioResult,sessionId:%s index:%d\n", time.Now().Format("2006-01-02 15:04:05"), l.SessionId, l.Index)
	l.Index = l.Index + 1
	// l.result.Data <- data
	l.Lock()
	defer l.Unlock()
	l.result.Audios = append(l.result.Audios, data...)
}

func (l *realtimeTTSListener) OnTextResult(r *tts.SpeechWsSynthesisResponse) {
	l.Lock()
	l.Unlock()

	subtitles := make([]speech.Subtitle, 0, len(r.Result.Subtitles))
	for i := 0; i < len(r.Result.Subtitles); i++ {
		subtitles = append(subtitles, speech.Subtitle{
			BeginTime:  r.Result.Subtitles[i].BeginTime,
			EndTime:    r.Result.Subtitles[i].EndTime,
			BeginIndex: int64(r.Result.Subtitles[i].BeginIndex),
			EndIndex:   int64(r.Result.Subtitles[i].EndIndex),
			Phoneme:    r.Result.Subtitles[i].Phoneme,
			Text:       r.Result.Subtitles[i].Text,
		})
	}
	l.result.Subtitles = subtitles
	// fmt.Printf("%s|OnTextResult,sessionId:%s response: %s\n", time.Now().Format("2006-01-02 15:04:05"), l.SessionId, r.ToString())
}

func (l *realtimeTTSListener) OnSynthesisFail(r *tts.SpeechWsSynthesisResponse, err error) {
	close(l.result.Data)
	fmt.Printf("%s|OnSynthesisFail,sessionId:%s response: %s err:%s\n", time.Now().Format("2006-01-02 15:04:05"), l.SessionId, r.ToString(), err.Error())
}
