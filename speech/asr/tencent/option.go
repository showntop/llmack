package tencent

import "github.com/showntop/tencentcloud-speech-sdk-go/asr"

// Option ...
type Option func(*Options)

// Options ...
type Options struct {
	AppID     string
	SecretID  string
	SecretKey string

	listener asr.SpeechRecognitionListener
}

// WithAppID ...
func WithAppID(appID string) Option {
	return func(o *Options) {
		o.AppID = appID
	}
}

// WithSecretID ...
func WithSecretID(secretID string) Option {
	return func(o *Options) {
		o.SecretID = secretID
	}
}

// WithSecretKey ...
func WithSecretKey(secretKey string) Option {
	return func(o *Options) {
		o.SecretKey = secretKey
	}
}

// WithListener ...
func WithListener(l asr.SpeechRecognitionListener) Option {
	return func(o *Options) {
		o.listener = l
	}
}
