package minmax

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/speech"
)

const url = "https://api.minimax.chat/v1/t2a_v2"

// Options ...
type Options struct {
	apiKey       string
	gouupID      int64
	model        string
	voiceSetting map[string]any
	audioSetting map[string]any
}

// Option ...
type Option func(*Options)

// WithAPIKey ...
func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.apiKey = apiKey
	}
}

// WithGroupID ...
func WithGroupID(gouupID int64) Option {
	return func(o *Options) {
		o.gouupID = gouupID
	}
}

// WithModel ...
func WithModel(model string) Option {
	return func(o *Options) {
		o.model = model
	}
}

// WithVoiceSetting ...
func WithVoiceSetting(voiceSetting map[string]any) Option {
	return func(o *Options) {
		o.voiceSetting = voiceSetting
	}
}

// WithAudioSetting ...
func WithAudioSetting(audioSetting map[string]any) Option {
	return func(o *Options) {
		o.audioSetting = audioSetting
	}
}

// RealtimeTTS ...
type RealtimeTTS struct {
	sync.Mutex
	options Options

	client *http.Client
}

// NewRealtimeTTS ...
func NewRealtimeTTS(opts ...Option) (speech.RealtimeTTS, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	if options.apiKey == "" || options.gouupID == 0 {
		return nil, errors.New("apiKey and gouupID are required")
	}

	tts := new(RealtimeTTS)
	tts.options = options

	tts.client = http.DefaultClient
	return tts, nil
}

// Synthesize ...
func (t *RealtimeTTS) Synthesize(ctx context.Context, text string) (*speech.TTSResult, error) {
	// result := new(speech.TTSResult)
	result := speech.NewTTSResult()

	body := map[string]any{
		"model":         t.options.model,
		"text":          text,
		"stream":        true,
		"voice_setting": t.options.voiceSetting,
		"audio_setting": t.options.audioSetting,
		// "voice_setting": map[string]any{
		// 	"voice_id": "kefu_female_ai_mix_8",
		// 	"speed":    1.10,
		// 	"vol":      1.0,
		// 	"pitch":    0,
		// },
		// "pronunciation_dict": map[string]any{
		// 	"tone": []string{
		// 		"处理/(chu3)(li3)", "危险/dangerous",
		// 	},
		// },
		// "audio_setting": map[string]any{
		// 	"sample_rate": 16000,
		// 	"format":      "pcm",
		// 	"channel":     1,
		// 	// "bitrate":     128000,
		// },
	}
	payload, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.options.apiKey))
	req.Header.Set("X-Group-ID", fmt.Sprintf("%d", t.options.gouupID))

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, %s", resp.StatusCode, resp.Status)
	}

	go func() {
		defer resp.Body.Close()
		defer close(result.Data)

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 4096*1024), 4096*1024)

		for scanner.Scan() {
			chunk := scanner.Text()
			log.InfoContextf(ctx, "minimax tts chunk: %d", len(chunk))
			if len(chunk) == 0 {
				log.WarnContextf(ctx, "minimax tts chunk is empty %s", chunk)
				continue
			}
			if chunk[:5] != "data:" {
				log.WarnContextf(ctx, "minimax tts chunk not start with data: %s", chunk)
				continue
			}
			payload := struct {
				Data struct {
					Status int    `json:"status"`
					Audio  string `json:"audio"`
				} `json:"data"`
				TraceID string `json:"trace_id"`
			}{}
			if err := json.Unmarshal([]byte(chunk[5:]), &payload); err != nil {
				log.ErrorContextf(ctx, "unmarshal tts data error: %v, payload: %s trace_id: %s", err, chunk[5:], payload.TraceID)
				continue
			}
			if payload.Data.Status != 1 {
				continue
			}
			log.InfoContextf(ctx, "minimax tts chunk: %d, trace_id: %s", len(chunk), payload.TraceID)
			audio, _ := hex.DecodeString(payload.Data.Audio)
			result.Data <- audio
		}

		if err := scanner.Err(); err != nil {
			log.ErrorContextf(ctx, "scanner error: %v", err)
		}
	}()
	return result, nil
}
