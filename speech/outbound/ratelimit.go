package outbound

import (
	"time"
)

// RateLimitOutbound ...
type RateLimitOutbound struct {
	*BaseOutbound
	handle func([]byte) error
}

// NewRateLimitOutbound ...
func NewRateLimitOutbound() *RateLimitOutbound {
	out := &RateLimitOutbound{}
	out.BaseOutbound = NewBaseOutbound()
	go out.Loop()
	return out
}

// Loop ...
func (o *RateLimitOutbound) Loop() error {
	for chunk := range o.q {
		startTime := time.Now()
		// 置换 chunk 状态
		chunkSizePerDuration := float64(o.sampleRate) * 2.00 / float64(time.Second)
		durations := time.Duration(float64(len(chunk)) / chunkSizePerDuration)
		o.handle(chunk)
		endTime := time.Now()

		x := durations - endTime.Sub(startTime) - 10*time.Millisecond
		if x > 0 {
			time.Sleep(x)
		}

	}
	return nil
}
