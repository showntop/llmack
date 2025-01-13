package extractor

// NewUnknownExtractor ...
func NewUnknownExtractor() *UnknownExtractor {

	return &UnknownExtractor{}
}

// UnknownExtractor ...
type UnknownExtractor struct {
}

// Extract ...
func (d *UnknownExtractor) Extract(m *Meta, binary []byte) ([]string, error) {
	return nil, nil
}
