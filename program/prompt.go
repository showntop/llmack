package program

import (
	"fmt"
	"reflect"
)

// Field represents an input or output field in a signature
type Field struct {
	Name        string
	Type        reflect.Kind
	Description string
	Marker      string
	Kind        string // input or output
}

// Promptx defines the instruction and input/output about a program
type Promptx struct {
	Name         string
	Description  string
	Prompt       string
	Instruction  string
	InputFields  map[string]*Field
	OutputFields map[string]*Field
}

// ValidateInputs checks if the provided inputs match the signature
func (m *Promptx) ValidateInputs(inputs map[string]any) error {
	for name := range m.InputFields {
		if _, ok := inputs[name]; !ok {
			return fmt.Errorf("missing required input field: %s", name)
		}
	}
	return nil
}

// ValidateOutputs checks if the outputs match the signature
func (m *Promptx) ValidateOutputs(outputs map[string]any) error {
	for name := range m.OutputFields {
		if _, ok := outputs[name]; !ok {
			return fmt.Errorf("missing required output field: %s", name)
		}
	}
	return nil
}
