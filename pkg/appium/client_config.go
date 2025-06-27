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
	"time"

	"github.com/tebeka/selenium"
)

// AppiumClientConfig contains configuration for Appium client
// This struct extends functionality similar to selenium ClientConfig
type AppiumClientConfig struct {
	// RemoteServerAddr is the address of the Appium server
	RemoteServerAddr string

	// DirectConnection enables directConnect feature
	// https://github.com/appium/python-client?tab=readme-ov-file#direct-connect-urls
	DirectConnection bool

	// KeepAlive controls HTTP keep-alive
	KeepAlive bool

	// Timeout for HTTP requests
	Timeout time.Duration

	// Additional capabilities
	Capabilities selenium.Capabilities
}

// NewAppiumClientConfig creates a new AppiumClientConfig with default values
func NewAppiumClientConfig(remoteServerAddr string) *AppiumClientConfig {
	return &AppiumClientConfig{
		RemoteServerAddr: remoteServerAddr,
		DirectConnection: false,
		KeepAlive:        true,
		Timeout:          30 * time.Second,
		Capabilities:     selenium.Capabilities{},
	}
}

// WithDirectConnection enables direct connection feature
func (c *AppiumClientConfig) WithDirectConnection(enable bool) *AppiumClientConfig {
	c.DirectConnection = enable
	return c
}

// WithKeepAlive sets the keep-alive setting
func (c *AppiumClientConfig) WithKeepAlive(keepAlive bool) *AppiumClientConfig {
	c.KeepAlive = keepAlive
	return c
}

// WithTimeout sets the HTTP timeout
func (c *AppiumClientConfig) WithTimeout(timeout time.Duration) *AppiumClientConfig {
	c.Timeout = timeout
	return c
}

// WithCapability adds a capability
func (c *AppiumClientConfig) WithCapability(key string, value interface{}) *AppiumClientConfig {
	if c.Capabilities == nil {
		c.Capabilities = selenium.Capabilities{}
	}
	c.Capabilities[key] = value
	return c
}

// GetDirectConnection returns whether directConnect is enabled
func (c *AppiumClientConfig) GetDirectConnection() bool {
	return c.DirectConnection
}
