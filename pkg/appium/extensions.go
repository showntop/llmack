// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package appiumgo

import (
	"fmt"
)

// ExecuteFunc represents a function that can execute commands
type ExecuteFunc func(command string, params map[string]interface{}) (interface{}, error)

// ExtensionBase provides the base functionality for creating custom extensions
// This mimics the Python ExtensionBase class functionality
type ExtensionBase interface {
	// MethodName returns the method name that will be available on the driver
	MethodName() string

	// AddCommand returns the HTTP method and URL for the command
	AddCommand() (string, string)

	// SetExecuteFunc sets the execute function for this extension
	SetExecuteFunc(execute ExecuteFunc)
}

// BaseExtension provides a default implementation of ExtensionBase
type BaseExtension struct {
	executeFunc ExecuteFunc
}

// Execute executes the command with optional parameters
func (be *BaseExtension) Execute(parameters map[string]interface{}) (interface{}, error) {
	if be.executeFunc == nil {
		return nil, fmt.Errorf("execute function not set")
	}

	params := make(map[string]interface{})
	if parameters != nil {
		params = parameters
	}

	return be.executeFunc(be.MethodName(), params)
}

// SetExecuteFunc sets the execute function
func (be *BaseExtension) SetExecuteFunc(execute ExecuteFunc) {
	be.executeFunc = execute
}

// MethodName must be implemented by concrete extensions
func (be *BaseExtension) MethodName() string {
	panic("MethodName must be implemented by concrete extension")
}

// AddCommand must be implemented by concrete extensions
func (be *BaseExtension) AddCommand() (string, string) {
	panic("AddCommand must be implemented by concrete extension")
}

// Extension manager handles extensions for the WebDriver
type ExtensionManager struct {
	extensions       []ExtensionBase
	absentExtensions map[string]bool
}

// NewExtensionManager creates a new extension manager
func NewExtensionManager() *ExtensionManager {
	return &ExtensionManager{
		extensions:       make([]ExtensionBase, 0),
		absentExtensions: make(map[string]bool),
	}
}

// AddExtension adds an extension to the manager
func (em *ExtensionManager) AddExtension(ext ExtensionBase) {
	em.extensions = append(em.extensions, ext)
}

// GetExtensions returns all registered extensions
func (em *ExtensionManager) GetExtensions() []ExtensionBase {
	return em.extensions
}

// AssertExtensionExists verifies if the given extension is not in the absent list
func (em *ExtensionManager) AssertExtensionExists(extName string) error {
	if em.absentExtensions[extName] {
		return fmt.Errorf("extension %s is marked as absent", extName)
	}
	return nil
}

// MarkExtensionAbsence marks the given extension as absent
func (em *ExtensionManager) MarkExtensionAbsence(extName string) {
	em.absentExtensions[extName] = true
}

// DeleteExtensions removes all extensions (for cleanup)
func (em *ExtensionManager) DeleteExtensions() {
	em.extensions = make([]ExtensionBase, 0)
	em.absentExtensions = make(map[string]bool)
}

// Example custom extension implementation
type ExampleExtension struct {
	BaseExtension
}

// MethodName returns the method name for this extension
func (ee *ExampleExtension) MethodName() string {
	return "customMethodName"
}

// AddCommand returns the HTTP method and URL for this extension
func (ee *ExampleExtension) AddCommand() (string, string) {
	return "GET", "session/$sessionId/path/to/your/custom/url"
}

// CustomMethodName is the actual method that will be called
func (ee *ExampleExtension) CustomMethodName() (interface{}, error) {
	// Generally the response of Appium follows { 'value': { data } } format
	result, err := ee.Execute(nil)
	if err != nil {
		return nil, err
	}

	// Extract value from result if it's a map
	if resultMap, ok := result.(map[string]interface{}); ok {
		if value, exists := resultMap["value"]; exists {
			return value, nil
		}
	}

	return result, nil
}
