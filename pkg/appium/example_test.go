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
	"testing"
	"time"
)

// ExampleBasicUsage demonstrates basic usage of the Appium Go client
func ExampleBasicUsage() {
	// Create Appium options
	options := NewAppiumOptions().
		SetPlatformName("Android").
		SetDeviceName("emulator-5554").
		SetApp("/path/to/your/app.apk").
		SetAutomationName("UIAutomator2")

	// Create client configuration
	clientConfig := NewAppiumClientConfig("http://127.0.0.1:4723").
		WithTimeout(30 * time.Second).
		WithKeepAlive(true)

	// Create WebDriver
	driver, err := NewWebDriver("", options, clientConfig, nil)
	if err != nil {
		panic(err)
	}
	defer driver.Close()

	// Start session
	err = driver.StartSession(options.ToCapabilities())
	if err != nil {
		panic(err)
	}

	// Get server status
	status, err := driver.GetStatus()
	if err != nil {
		panic(err)
	}

	// Print status
	println("Server status:", status)

	// Get orientation
	orientation, err := driver.GetOrientation()
	if err != nil {
		panic(err)
	}

	println("Current orientation:", orientation)
}

// ExampleWithApplications demonstrates using the Applications extension
func ExampleWithApplications() {
	// Create options and driver (same as above)
	options := NewAppiumOptions().
		SetPlatformName("Android").
		SetDeviceName("emulator-5554").
		SetAutomationName("UIAutomator2")

	driver, err := NewWebDriver("", options, nil, nil)
	if err != nil {
		panic(err)
	}
	defer driver.Close()

	// Create Applications extension
	apps := NewApplications(driver)

	// Check if app is installed
	bundleID := "com.example.app"
	installed, err := apps.IsAppInstalled(bundleID)
	if err != nil {
		panic(err)
	}

	if !installed {
		// Install app if not installed
		err = apps.InstallApp("/path/to/app.apk", map[string]interface{}{
			"replace": true,
		})
		if err != nil {
			panic(err)
		}
	}

	// Activate app
	err = apps.ActivateApp(bundleID)
	if err != nil {
		panic(err)
	}

	// Put app in background for 5 seconds
	err = apps.BackgroundApp(5)
	if err != nil {
		panic(err)
	}

	// Query app state
	state, err := apps.QueryAppState(bundleID)
	if err != nil {
		panic(err)
	}

	println("App state:", state)
}

// ExampleWithCustomExtension demonstrates creating and using a custom extension
func ExampleWithCustomExtension() {
	// Create custom extension
	customExt := &ExampleExtension{}

	// Create driver with custom extension
	options := NewAppiumOptions().
		SetPlatformName("iOS").
		SetDeviceName("iPhone Simulator").
		SetAutomationName("XCUITest")

	driver, err := NewWebDriver("", options, nil, []ExtensionBase{customExt})
	if err != nil {
		panic(err)
	}
	defer driver.Close()

	// Use custom extension method
	result, err := customExt.CustomMethodName()
	if err != nil {
		panic(err)
	}

	println("Custom extension result:", result)
}

// ExampleWithDirectConnection demonstrates using direct connection feature
func ExampleWithDirectConnection() {
	// Create client config with direct connection enabled
	clientConfig := NewAppiumClientConfig("http://127.0.0.1:4723").
		WithDirectConnection(true).
		WithKeepAlive(true)

	options := NewAppiumOptions().
		SetPlatformName("Android").
		SetDeviceName("emulator-5554").
		SetAutomationName("UIAutomator2")

	driver, err := NewWebDriver("", options, clientConfig, nil)
	if err != nil {
		panic(err)
	}
	defer driver.Close()

	// Start session - this will use direct connection if server supports it
	err = driver.StartSession(options.ToCapabilities())
	if err != nil {
		panic(err)
	}

	println("Session started with direct connection")
}

// TestOptionsBuilder tests the options builder pattern
func TestOptionsBuilder(t *testing.T) {
	options := NewAppiumOptions().
		SetPlatformName("Android").
		SetDeviceName("emulator-5554").
		SetApp("/path/to/app.apk").
		SetAutomationName("UIAutomator2").
		SetUDID("device-udid").
		SetNoReset(true).
		SetFullReset(false).
		SetNewCommandTimeout(60)

	caps := options.ToCapabilities()

	// Check that capabilities are set correctly
	if caps["platformName"] != "Android" {
		t.Errorf("Expected platformName to be Android, got %v", caps["platformName"])
	}

	if caps["appium:deviceName"] != "emulator-5554" {
		t.Errorf("Expected deviceName to be emulator-5554, got %v", caps["appium:deviceName"])
	}

	if caps["appium:app"] != "/path/to/app.apk" {
		t.Errorf("Expected app path, got %v", caps["appium:app"])
	}
}

// TestW3CConversion tests W3C capability conversion
func TestW3CConversion(t *testing.T) {
	capabilities := map[string]interface{}{
		"platformName":   "iOS",
		"deviceName":     "iPhone 12",
		"automationName": "XCUITest",
		"acceptSslCerts": true,   // Should be converted to acceptInsecureCerts
		"version":        "14.5", // Should be converted to browserVersion
	}

	w3c := AsW3C(capabilities)

	// Check structure
	if _, exists := w3c["capabilities"]; !exists {
		t.Error("Expected capabilities key in W3C format")
	}

	caps := w3c["capabilities"].(map[string]interface{})
	alwaysMatch := caps["alwaysMatch"].(map[string]interface{})

	// Check conversions
	if alwaysMatch["acceptInsecureCerts"] != true {
		t.Error("Expected acceptSslCerts to be converted to acceptInsecureCerts")
	}

	if alwaysMatch["browserVersion"] != "14.5" {
		t.Error("Expected version to be converted to browserVersion")
	}

	// Check appium prefix
	if alwaysMatch["appium:deviceName"] != "iPhone 12" {
		t.Error("Expected deviceName to have appium prefix")
	}
}
