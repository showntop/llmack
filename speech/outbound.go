package speech

// Outbound ...
type Outbound interface {
	Write([]byte) error
	Reset() error
	Close() error
}
