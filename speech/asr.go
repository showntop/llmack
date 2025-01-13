package speech

// ASR ...
type ASR interface {
	Recognize(buffer []byte) (string, error)
	Input(buffer []byte) error // 流入音频数据， async recognize
	Close() error              // 流入音频数据， async recognize
}
