package tencent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/showntop/llmack/speech"

	"github.com/google/uuid"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	stts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tts/v20190823"
)

// TextTTS ...
type TextTTS struct {
	sync.Mutex

	options Options
	client  *stts.Client
}

// NewTextTTS  ...
func NewTextTTS(opts ...Option) (speech.TextTTS, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	if options.secretID == "" || options.secretKey == "" {
		return nil, errors.New("secretID, secretKey are required")
	}

	tencent := new(TextTTS)
	tencent.options = options
	credential := common.NewCredential(options.secretID, options.secretKey)
	cpf := profile.NewClientProfile()
	cli, err := stts.NewClient(credential, regions.Guangzhou, cpf)
	if err != nil {
		return nil, err
	}
	tencent.client = cli
	return tencent, nil
}

// Terminate ...
func (t *TextTTS) Terminate() error {
	return nil
}

// Synthesize ...
func (t *TextTTS) Synthesize(ctx context.Context, text string) (*speech.TextTTSResult, error) {
	if !t.options.longText {
		return t.Text2Voice(ctx, text)
	} else {
		return t.LongText2Voice(ctx, text)
	}
}

// LongText2Voice ...
func (t *TextTTS) LongText2Voice(ctx context.Context, text string) (*speech.TextTTSResult, error) {
	request := stts.NewCreateTtsTaskRequest()
	request.BaseRequest.SetDomain("tts.internal.tencentcloudapi.com")
	request.Text = common.StringPtr(text)
	request.VoiceType = common.Int64Ptr(501005)
	request.Volume = common.Float64Ptr(5)
	request.Speed = common.Float64Ptr(0)
	// request.Codec = "mp3"
	request.EnableSubtitle = common.BoolPtr(true)
	// request.SampleRate = 16000
	// request.Speed = -1
	//request.EmotionCategory = "happy"
	//request.EmotionIntensity = 200
	//request.Debug = true
	//request.DebugFunc = func(message string) { fmt.Println(message) }
	resp, err := t.client.CreateTtsTaskWithContext(ctx, request)
	if err != nil {
		return nil, err
	}

	if *resp.Response.Data.TaskId == "" {
		return nil, fmt.Errorf("ff")
	}

	for {
		request2 := stts.NewDescribeTtsTaskStatusRequest()
		request2.TaskId = resp.Response.Data.TaskId
		resp2, err := t.client.DescribeTtsTaskStatusWithContext(ctx, request2)
		if err != nil {
			return nil, err
		}
		if *resp2.Response.Data.Status == 0 {
			continue
		}
		if *resp2.Response.Data.Status == 1 {
			continue
		}
		if *resp2.Response.Data.Status == 3 {
			return nil, fmt.Errorf("tts failed")
		}
		if *resp2.Response.Data.Status == 2 {
			data := resp2.Response.Data
			subtitles := make([]speech.Subtitle, 0, len(data.Subtitles))
			for i := 0; i < len(data.Subtitles); i++ {
				subtitles = append(subtitles, speech.Subtitle{
					BeginTime:  *data.Subtitles[i].BeginTime,
					EndTime:    *data.Subtitles[i].EndTime,
					BeginIndex: *data.Subtitles[i].BeginIndex,
					EndIndex:   *data.Subtitles[i].EndIndex,
					Phoneme:    *data.Subtitles[i].Phoneme,
					Text:       *data.Subtitles[i].Text,
				})
			}
			fmt.Println(resp2.Response.Data)
			return &speech.TextTTSResult{
				Audio:     "",
				Subtitles: subtitles,
			}, nil
		}
	}
}

// Text2Voice ...
func (t *TextTTS) Text2Voice(ctx context.Context, text string) (*speech.TextTTSResult, error) {
	request := stts.NewTextToVoiceRequest()
	request.BaseRequest.SetDomain("tts.internal.tencentcloudapi.com")
	request.SessionId = common.StringPtr(uuid.NewString())
	request.Text = common.StringPtr(text)
	request.VoiceType = common.Int64Ptr(301040)
	request.Volume = common.Float64Ptr(5)
	request.Speed = common.Float64Ptr(0)
	// request.Codec = "mp3"
	request.EnableSubtitle = common.BoolPtr(true)
	// request.SampleRate = 16000
	// request.Speed = -1
	//request.EmotionCategory = "happy"
	//request.EmotionIntensity = 200
	//request.Debug = true
	//request.DebugFunc = func(message string) { fmt.Println(message) }
	resp, err := t.client.TextToVoice(request)
	if err != nil {
		return nil, err
	}

	subtitles := make([]speech.Subtitle, 0, len(resp.Response.Subtitles))
	for i := 0; i < len(resp.Response.Subtitles); i++ {
		subtitles = append(subtitles, speech.Subtitle{
			BeginTime:  *resp.Response.Subtitles[i].BeginTime,
			EndTime:    *resp.Response.Subtitles[i].EndTime,
			BeginIndex: *resp.Response.Subtitles[i].BeginIndex,
			EndIndex:   *resp.Response.Subtitles[i].EndIndex,
			Phoneme:    *resp.Response.Subtitles[i].Phoneme,
			Text:       *resp.Response.Subtitles[i].Text,
		})
	}

	return &speech.TextTTSResult{
		Audio:     *resp.Response.Audio,
		Subtitles: subtitles,
	}, nil
}
