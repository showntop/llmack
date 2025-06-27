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
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tebeka/selenium"
)

const (
	// PrefixHeader is the user agent prefix for Appium
	PrefixHeader = "appium/"

	// HeaderIdempotencyKey is the header for idempotency
	HeaderIdempotencyKey = "X-Idempotency-Key"

	// Version represents the library version
	Version = "1.0.0"
)

// AppiumConnection wraps selenium.Remote with Appium-specific functionality
type AppiumConnection struct {
	selenium.WebDriver
	clientConfig *AppiumClientConfig
	extraHeaders map[string]string
	userAgent    string
}

// NewAppiumConnection creates a new Appium connection
func NewAppiumConnection(clientConfig *AppiumClientConfig) (*AppiumConnection, error) {
	if clientConfig == nil {
		clientConfig = NewAppiumClientConfig("http://127.0.0.1:4723")
	}

	// Set up capabilities
	caps := selenium.Capabilities{
		// "browserName": "",
	}

	// Merge user capabilities
	for k, v := range clientConfig.Capabilities {
		caps[k] = v
	}

	// Create selenium WebDriver
	wd, err := selenium.NewRemote(caps, clientConfig.RemoteServerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create selenium remote: %w", err)
	}

	conn := &AppiumConnection{
		WebDriver:    wd,
		clientConfig: clientConfig,
		extraHeaders: make(map[string]string),
		userAgent:    fmt.Sprintf("%s%s", PrefixHeader, Version),
	}

	return conn, nil
}

// GetRemoteConnectionHeaders returns headers for remote connection
// This method mimics the Python implementation for controlling extra headers
func (ac *AppiumConnection) GetRemoteConnectionHeaders(url string, keepAlive bool) map[string]string {
	headers := make(map[string]string)

	// Set user agent
	headers["User-Agent"] = ac.userAgent

	// Handle idempotency key for session creation
	if strings.HasSuffix(url, "/session") {
		// Add idempotency key for new session requests
		// https://github.com/appium/appium-base-driver/pull/400
		ac.extraHeaders[HeaderIdempotencyKey] = uuid.New().String()
	} else {
		// Remove idempotency key for non-session requests
		delete(ac.extraHeaders, HeaderIdempotencyKey)
	}

	// Merge extra headers
	for k, v := range ac.extraHeaders {
		headers[k] = v
	}

	// Set keep-alive
	if keepAlive {
		headers["Connection"] = "keep-alive"
	}

	return headers
}

// Execute executes a command with the given parameters
func (ac *AppiumConnection) Execute(command string, params map[string]interface{}) (interface{}, error) {
	// This is a simplified version - in a full implementation, you would need to
	// handle the command execution similar to how selenium.WebDriver does it
	// For now, we delegate to the underlying WebDriver

	// Handle special Appium commands here if needed
	switch command {
	case GetStatus:
		return ac.executeGetStatus()
	default:
		// For other commands, we would need to implement the command execution
		// This is a simplified placeholder
		return nil, fmt.Errorf("command %s not implemented", command)
	}
}

// executeGetStatus executes the getStatus command
func (ac *AppiumConnection) executeGetStatus() (interface{}, error) {
	// This would need to be implemented to make an actual HTTP request to /status
	// For now, this is a placeholder
	return map[string]interface{}{
		"build": map[string]interface{}{
			"version": Version,
		},
	}, nil
}

// SetHTTPClient allows setting a custom HTTP client
func (ac *AppiumConnection) SetHTTPClient(client *http.Client) {
	// This would be used to set a custom HTTP client for the connection
	// Implementation depends on how the underlying selenium driver handles this
}

// GetClientConfig returns the client configuration
func (ac *AppiumConnection) GetClientConfig() *AppiumClientConfig {
	return ac.clientConfig
}

// Close closes the connection
func (ac *AppiumConnection) Close() error {
	if ac.WebDriver != nil {
		return ac.WebDriver.Quit()
	}
	return nil
}
