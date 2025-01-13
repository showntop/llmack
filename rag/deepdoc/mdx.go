package deepdoc

// MdxDocument ...
type MdxDocument struct {
}

// Mdx ...
func Mdx() *MdxDocument {
	return &MdxDocument{}
}

// Extract ...
func (d *MdxDocument) Extract(filename string, binary []byte) ([]string, error) {
	return []string{string(binary)}, nil
}
