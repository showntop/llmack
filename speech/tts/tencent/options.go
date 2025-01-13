package tencent

import "github.com/showntop/tencentcloud-speech-sdk-go/tts"

// Options ...
type Options struct {
	secretID  string
	secretKey string
	appID     int64
	longText  bool
	Listener  tts.SpeechWsv2SynthesisListener
}

// Option ...
type Option func(*Options)

// WithSecretID ...
func WithSecretID(secretID string) Option {
	return func(o *Options) {
		o.secretID = secretID
	}
}

// WithSecretKey ...
func WithSecretKey(secretKey string) Option {
	return func(o *Options) {
		o.secretKey = secretKey
	}
}

// WithLongText ...
func WithLongText(longText bool) Option {
	return func(o *Options) {
		o.longText = longText
	}
}

// WithAppID ...
func WithAppID(appID int64) Option {
	return func(o *Options) {
		o.appID = appID
	}
}

// WithListener ...
func WithListener(listener tts.SpeechWsv2SynthesisListener) Option {
	return func(o *Options) {
		o.Listener = listener
	}
}
