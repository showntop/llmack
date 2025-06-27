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
	"strings"

	"github.com/tebeka/selenium"
)

const (
	// AppiumPrefix is the prefix for Appium-specific capabilities
	AppiumPrefix = "appium:"

	// PlatformName capability key
	PlatformName = "platformName"

	// BrowserName capability key
	BrowserName = "browserName"
)

// W3C capability names that don't need the appium prefix
var W3CCapabilityNames = map[string]bool{
	"acceptInsecureCerts":     true,
	BrowserName:               true,
	"browserVersion":          true,
	PlatformName:              true,
	"pageLoadStrategy":        true,
	"proxy":                   true,
	"setWindowRect":           true,
	"timeouts":                true,
	"unhandledPromptBehavior": true,
}

// OSS to W3C capability conversion map
var OSSToW3CConversion = map[string]string{
	"acceptSslCerts": "acceptInsecureCerts",
	"version":        "browserVersion",
	"platform":       PlatformName,
}

// AppiumOptions represents Appium capabilities and options
type AppiumOptions struct {
	caps             selenium.Capabilities
	ignoreLocalProxy bool
}

// NewAppiumOptions creates a new AppiumOptions instance
func NewAppiumOptions() *AppiumOptions {
	return &AppiumOptions{
		caps:             make(selenium.Capabilities),
		ignoreLocalProxy: false,
	}
}

// SetCapability sets a capability value
func (opts *AppiumOptions) SetCapability(name string, value interface{}) *AppiumOptions {
	w3cName := name
	if !W3CCapabilityNames[name] && !strings.Contains(name, ":") {
		w3cName = AppiumPrefix + name
	}

	if value == nil {
		delete(opts.caps, w3cName)
	} else {
		opts.caps[w3cName] = value
	}

	return opts
}

// GetCapability fetches a capability value or nil if not set
func (opts *AppiumOptions) GetCapability(name string) interface{} {
	if value, exists := opts.caps[name]; exists {
		return value
	}

	appiumName := AppiumPrefix + name
	if value, exists := opts.caps[appiumName]; exists {
		return value
	}

	return nil
}

// LoadCapabilities sets multiple capabilities
func (opts *AppiumOptions) LoadCapabilities(capabilities map[string]interface{}) *AppiumOptions {
	for name, value := range capabilities {
		opts.SetCapability(name, value)
	}
	return opts
}

// AsW3C formats capabilities to a valid W3C session request object
func AsW3C(capabilities map[string]interface{}) map[string]interface{} {
	processedCaps := make(map[string]interface{})

	for k, v := range capabilities {
		key := processKey(k)
		processedCaps[key] = v
	}

	return map[string]interface{}{
		"capabilities": map[string]interface{}{
			"firstMatch":  []map[string]interface{}{{}},
			"alwaysMatch": processedCaps,
		},
	}
}

// ToW3C formats the instance to a valid W3C session request object
func (opts *AppiumOptions) ToW3C() map[string]interface{} {
	return AsW3C(opts.ToCapabilities())
}

// ToCapabilities returns a copy of the capabilities
func (opts *AppiumOptions) ToCapabilities() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range opts.caps {
		result[k] = v
	}
	return result
}

// GetDefaultCapabilities returns default capabilities
func (opts *AppiumOptions) GetDefaultCapabilities() map[string]interface{} {
	return make(map[string]interface{})
}

// processKey processes capability key according to W3C rules
func processKey(k string) string {
	// Check OSS to W3C conversion
	if w3cKey, exists := OSSToW3CConversion[k]; exists {
		k = w3cKey
	}

	// If it's a W3C capability name, return as-is
	if W3CCapabilityNames[k] {
		return k
	}

	// If it already has a prefix (contains ':'), return as-is
	if strings.Contains(k, ":") {
		return k
	}

	// Otherwise, add the appium prefix
	return AppiumPrefix + k
}

// Common capability setters for convenience

// SetPlatformName sets the platform name capability
func (opts *AppiumOptions) SetPlatformName(platform string) *AppiumOptions {
	return opts.SetCapability(PlatformName, platform)
}

// SetDeviceName sets the device name capability
func (opts *AppiumOptions) SetDeviceName(deviceName string) *AppiumOptions {
	return opts.SetCapability("deviceName", deviceName)
}

// SetApp sets the app capability
func (opts *AppiumOptions) SetApp(app string) *AppiumOptions {
	return opts.SetCapability("app", app)
}

// SetAutomationName sets the automation name capability
func (opts *AppiumOptions) SetAutomationName(automationName string) *AppiumOptions {
	return opts.SetCapability("automationName", automationName)
}

// SetUDID sets the UDID capability
func (opts *AppiumOptions) SetUDID(udid string) *AppiumOptions {
	return opts.SetCapability("udid", udid)
}

// SetNoReset sets the noReset capability
func (opts *AppiumOptions) SetNoReset(noReset bool) *AppiumOptions {
	return opts.SetCapability("noReset", noReset)
}

// SetFullReset sets the fullReset capability
func (opts *AppiumOptions) SetFullReset(fullReset bool) *AppiumOptions {
	return opts.SetCapability("fullReset", fullReset)
}

// SetNewCommandTimeout sets the newCommandTimeout capability
func (opts *AppiumOptions) SetNewCommandTimeout(timeout int) *AppiumOptions {
	return opts.SetCapability("newCommandTimeout", timeout)
}
