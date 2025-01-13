package outbound

type handle func([]byte) error

// BaseOutbound ...
type BaseOutbound struct {
	q             chan []byte
	sampleRate    int
	audioEncoding int
}

// NewBaseOutbound ...
func NewBaseOutbound() *BaseOutbound {
	return &BaseOutbound{
		sampleRate:    16000,
		audioEncoding: 1,
		q:             make(chan []byte, 0),
	}
}

// Write ...
func (o *BaseOutbound) Write(data []byte) error {
	o.q <- data
	return nil
}

// Reset ...
func (o *BaseOutbound) Reset() error {
	return nil
}
