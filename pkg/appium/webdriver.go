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
	"strings"

	"github.com/tebeka/selenium"
)

// WebDriver represents an Appium WebDriver instance
// This is the main class that corresponds to Python's WebDriver class
type WebDriver struct {
	selenium.WebDriver
	*ActionHelpers
	*Applications

	connection       *AppiumConnection
	clientConfig     *AppiumClientConfig
	extensionManager *ExtensionManager
	capabilities     map[string]interface{}
	sessionID        string
}

// NewWebDriver creates a new Appium WebDriver instance
func NewWebDriver(commandExecutor string, options *AppiumOptions, clientConfig *AppiumClientConfig, extensions []ExtensionBase) (*WebDriver, error) {
	// Set default command executor if empty
	if commandExecutor == "" {
		commandExecutor = "http://127.0.0.1:4723"
	}

	// Create client config if not provided
	if clientConfig == nil {
		clientConfig = NewAppiumClientConfig(commandExecutor)
	}

	// Set server address if not already set
	if clientConfig.RemoteServerAddr == "" {
		clientConfig.RemoteServerAddr = commandExecutor
	}

	// Merge options into client config capabilities
	if options != nil {
		for k, v := range options.ToCapabilities() {
			clientConfig.WithCapability(k, v)
		}
	}

	// Create Appium connection
	connection, err := NewAppiumConnection(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Appium connection: %w", err)
	}

	// Create extension manager
	extensionManager := NewExtensionManager()

	// Create WebDriver instance
	wd := &WebDriver{
		WebDriver:        connection.WebDriver,
		connection:       connection,
		clientConfig:     clientConfig,
		extensionManager: extensionManager,
		capabilities:     make(map[string]interface{}),
	}

	// Initialize ActionHelpers
	wd.ActionHelpers = NewActionHelpers(wd)

	// Add extensions
	for _, ext := range extensions {
		wd.addExtension(ext)
	}

	// Add Appium-specific commands
	wd.addCommands()

	return wd, nil
}

// addExtension adds an extension to the WebDriver
func (wd *WebDriver) addExtension(ext ExtensionBase) {
	// Set the execute function for the extension
	ext.SetExecuteFunc(wd.Execute)

	// Add to extension manager
	wd.extensionManager.AddExtension(ext)

	// In a full implementation, you would also register the command with the command executor
	// For now, we just add it to the extension manager
}

// Execute executes a command with the given parameters
func (wd *WebDriver) Execute(command string, params map[string]interface{}) (interface{}, error) {
	// Delegate to the connection's Execute method
	return wd.connection.Execute(command, params)
}

// DeleteExtensions removes all extensions
func (wd *WebDriver) DeleteExtensions() {
	wd.extensionManager.DeleteExtensions()
}

// updateCommandExecutor updates command executor for directConnect feature
func (wd *WebDriver) updateCommandExecutor(keepAlive bool) error {
	const (
		directProtocol = "directConnectProtocol"
		directHost     = "directConnectHost"
		directPort     = "directConnectPort"
		directPath     = "directConnectPath"
	)

	if wd.capabilities == nil {
		return fmt.Errorf("driver capabilities must be defined")
	}

	// Check if all required direct connect capabilities are present
	requiredKeys := []string{directProtocol, directHost, directPort, directPath}
	for _, key := range requiredKeys {
		if _, exists := wd.capabilities[key]; !exists {
			// Log debug message about missing capabilities
			return fmt.Errorf("missing direct connect capability: %s", key)
		}
	}

	protocol := wd.capabilities[directProtocol].(string)
	hostname := wd.capabilities[directHost].(string)
	port := wd.capabilities[directPort]
	path := wd.capabilities[directPath].(string)

	executor := fmt.Sprintf("%s://%s:%v%s", protocol, hostname, port, path)

	// Update the connection's server address
	wd.clientConfig.RemoteServerAddr = executor

	// Create new connection with updated address
	newConnection, err := NewAppiumConnection(wd.clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create new connection: %w", err)
	}

	wd.connection = newConnection
	wd.WebDriver = newConnection.WebDriver

	// Re-add commands
	wd.addCommands()

	return nil
}

// StartSession creates a new session with the desired capabilities
func (wd *WebDriver) StartSession(capabilities map[string]interface{}) error {
	if capabilities == nil {
		return fmt.Errorf("capabilities must not be nil")
	}

	// Convert capabilities to W3C format
	w3cCaps := AsW3C(capabilities)

	// Execute NEW_SESSION command
	response, err := wd.Execute("newSession", w3cCaps)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Parse response
	if responseMap, ok := response.(map[string]interface{}); ok {
		// Extract session ID
		if sessionID, exists := responseMap["sessionId"]; exists {
			wd.sessionID = sessionID.(string)
		} else if value, exists := responseMap["value"]; exists {
			if valueMap, ok := value.(map[string]interface{}); ok {
				if sessionID, exists := valueMap["sessionId"]; exists {
					wd.sessionID = sessionID.(string)
				}
			}
		}

		// Extract capabilities
		if caps, exists := responseMap["capabilities"]; exists {
			wd.capabilities = caps.(map[string]interface{})
		} else if value, exists := responseMap["value"]; exists {
			if valueMap, ok := value.(map[string]interface{}); ok {
				if caps, exists := valueMap["capabilities"]; exists {
					wd.capabilities = caps.(map[string]interface{})
				}
			}
		}
	}

	if wd.sessionID == "" {
		return fmt.Errorf("failed to extract session ID from response")
	}

	// Update command executor if direct connection is enabled
	if wd.clientConfig.GetDirectConnection() {
		return wd.updateCommandExecutor(wd.clientConfig.KeepAlive)
	}

	return nil
}

// GetStatus gets the Appium server status
func (wd *WebDriver) GetStatus() (map[string]interface{}, error) {
	result, err := wd.Execute(GetStatus, nil)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		if value, exists := resultMap["value"]; exists {
			return value.(map[string]interface{}), nil
		}
		return resultMap, nil
	}

	return nil, fmt.Errorf("unexpected response format")
}

// GetOrientation gets the current orientation of the device
func (wd *WebDriver) GetOrientation() (string, error) {
	result, err := wd.Execute(GetScreenOrientation, nil)
	if err != nil {
		return "", err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		if value, exists := resultMap["value"]; exists {
			return value.(string), nil
		}
	}

	return "", fmt.Errorf("unexpected response format")
}

// SetOrientation sets the current orientation of the device
func (wd *WebDriver) SetOrientation(orientation string) error {
	allowedValues := []string{"LANDSCAPE", "PORTRAIT"}
	upperOrientation := strings.ToUpper(orientation)

	allowed := false
	for _, val := range allowedValues {
		if upperOrientation == val {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("you can only set the orientation to 'LANDSCAPE' and 'PORTRAIT'")
	}

	params := map[string]interface{}{
		"orientation": orientation,
	}

	_, err := wd.Execute(SetScreenOrientation, params)
	return err
}

// AssertExtensionExists verifies if the given extension is not present in the list of absent extensions
func (wd *WebDriver) AssertExtensionExists(extName string) error {
	return wd.extensionManager.AssertExtensionExists(extName)
}

// MarkExtensionAbsence marks the given extension as absent
func (wd *WebDriver) MarkExtensionAbsence(extName string) *WebDriver {
	wd.extensionManager.MarkExtensionAbsence(extName)
	return wd
}

// addCommands adds Appium-specific commands to the command executor
func (wd *WebDriver) addCommands() {
	// This method would add Appium-specific command mappings
	// In a full implementation, this would register commands with the HTTP client

	// Touch action commands
	// These would be registered with the underlying HTTP client
	// For now, we define the command mappings that would be used

	// Touch Actions (legacy)
	// POST /session/:sessionId/touch/perform - performTouchAction
	// POST /session/:sessionId/touch/multi/perform - performMultiAction

	// W3C Actions
	// POST /session/:sessionId/actions - performActions
	// DELETE /session/:sessionId/actions - releaseActions

	// Other Appium-specific commands would be registered here
}

// GetCapabilities returns the current session capabilities
func (wd *WebDriver) GetCapabilities() map[string]interface{} {
	return wd.capabilities
}

// GetSessionID returns the current session ID
func (wd *WebDriver) GetSessionID() string {
	return wd.sessionID
}

// Close closes the WebDriver session
func (wd *WebDriver) Close() error {
	if wd.connection != nil {
		return wd.connection.Close()
	}
	return nil
}
