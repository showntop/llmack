package speech

import "context"

// StreamTTS ...
type StreamTTS interface {
	Prepare() error
	Input(string) error
	Complete() error
	Terminate() error
	StreamResult() chan []byte
}

// TextTTS is the struct for text to speech
type TextTTS interface {
	Synthesize(ctx context.Context, text string) (*TextTTSResult, error)
}

// RealtimeTTS is the struct for text to speech
type RealtimeTTS interface {
	Synthesize(ctx context.Context, text string) (*TTSResult, error)
}

// TTSResult is the basic struct for the TTS response.
type TTSResult struct {
	SessionID string `json:"session_id"` //音频流唯一 id，由客户端在握手阶段生成并赋值在调用参数中
	// RequestId string             `json:"request_id"` //音频流唯一 id，由服务端在握手阶段自动生成
	MessageID string      `json:"message_id"` //本 message 唯一 id
	Data      chan []byte `json:"data"`       //最新语音合成文本结果/音频流数据
	Audios    []byte
	Subtitles []Subtitle `json:"subtitles"`
}

func NewTTSResult() *TTSResult {
	ttr := &TTSResult{Data: make(chan []byte, 100)}
	return ttr
}

type TextTTSResult struct {
	Audio     string
	Subtitles []Subtitle
}

type Subtitle struct {
	// ⽂本信息。
	Text string `json:"text,omitnil,omitempty" name:"text"`

	// ⽂本对应tts语⾳开始时间戳，单位ms。
	BeginTime int64 `json:"BeginTime,omitnil,omitempty" name:"BeginTime"`

	// ⽂本对应tts语⾳结束时间戳，单位ms。
	EndTime int64 `json:"EndTime,omitnil,omitempty" name:"EndTime"`

	// 该文本在时间戳数组中的开始位置，从0开始。
	BeginIndex int64 `json:"BeginIndex,omitnil,omitempty" name:"BeginIndex"`

	// 该文本在时间戳数组中的结束位置，从0开始。
	EndIndex int64 `json:"EndIndex,omitnil,omitempty" name:"EndIndex"`

	// 该字的音素。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Phoneme string `json:"Phoneme,omitnil,omitempty" name:"Phoneme"`
}
