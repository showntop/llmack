package deepdoc

import (
	"github.com/jeremywohl/flatten"
)

// JsonDocument ...
type JsonDocument struct {
}

// Json ...
//
// nolint:ignore
func Json() *JsonDocument {
	return &JsonDocument{}
}

// Extract ...
func (d *JsonDocument) Extract(filename string, binary []byte) ([]string, error) {

	flat, err := flatten.FlattenString(string(binary), "", flatten.DotStyle)
	if err != nil {
		return nil, err
	}

	return []string{flat}, nil
}
