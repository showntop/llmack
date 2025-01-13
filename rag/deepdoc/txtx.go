package deepdoc

import (
	"strings"
)

// TxtxDocument ...
type TxtxDocument struct {
}

// Txtx ...
func Txtx() *TxtxDocument {
	return &TxtxDocument{}
}

// Extract ...
func (d *TxtxDocument) Extract(filename string, binary []byte) ([]string, error) {
	sections := strings.Split(string(binary), "\n")
	return sections, nil
}
