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

// MobileCommand contains Appium-specific command constants
type MobileCommand struct{}

// Common commands
const (
	GetSession = "getSession"
	GetStatus  = "getStatus"

	// MJSONWP for Selenium v4
	GetLocation = "getLocation"
	SetLocation = "setLocation"

	Clear          = "clear"
	LocationInView = "locationInView"

	Contexts          = "getContexts"
	GetCurrentContext = "getCurrentContext"
	SwitchToContext   = "switchToContext"

	Background    = "background"
	GetAppStrings = "getAppStrings"

	IsLocked             = "isLocked"
	Lock                 = "lock"
	Unlock               = "unlock"
	GetDeviceTimeGet     = "getDeviceTimeGet"
	GetDeviceTimePost    = "getDeviceTimePost"
	InstallApp           = "installApp"
	RemoveApp            = "removeApp"
	IsAppInstalled       = "isAppInstalled"
	TerminateApp         = "terminateApp"
	ActivateApp          = "activateApp"
	QueryAppState        = "queryAppState"
	Shake                = "shake"
	HideKeyboard         = "hideKeyboard"
	PressKeycode         = "pressKeyCode"
	LongPressKeycode     = "longPressKeyCode"
	KeyEvent             = "keyEvent" // Needed for Selendroid
	PushFile             = "pushFile"
	PullFile             = "pullFile"
	PullFolder           = "pullFolder"
	GetClipboard         = "getClipboard"
	SetClipboard         = "setClipboard"
	FingerPrint          = "fingerPrint"
	GetSettings          = "getSettings"
	UpdateSettings       = "updateSettings"
	StartRecordingScreen = "startRecordingScreen"
	StopRecordingScreen  = "stopRecordingScreen"
	CompareImages        = "compareImages"
	IsKeyboardShown      = "isKeyboardShown"

	ExecuteDriver = "executeDriver"

	GetEvents = "getLogEvents"
	LogEvent  = "logCustomEvent"

	// MJSONWP for Selenium v4
	IsElementDisplayed   = "isElementDisplayed"
	GetCapabilities      = "getCapabilities"
	GetScreenOrientation = "getScreenOrientation"
	SetScreenOrientation = "setScreenOrientation"

	// To override selenium commands
	GetLog               = "getLog"
	GetAvailableLogTypes = "getAvailableLogTypes"

	// Android
	OpenNotifications       = "openNotifications"
	GetCurrentActivity      = "getCurrentActivity"
	GetCurrentPackage       = "getCurrentPackage"
	GetSystemBars           = "getSystemBars"
	GetDisplayDensity       = "getDisplayDensity"
	ToggleWifi              = "toggleWiFi"
	ToggleLocationServices  = "toggleLocationServices"
	GetPerformanceDataTypes = "getPerformanceDataTypes"
	GetPerformanceData      = "getPerformanceData"
	GetNetworkConnection    = "getNetworkConnection"
	SetNetworkConnection    = "setNetworkConnection"

	// Android Emulator
	SendSMS          = "sendSms"
	MakeGSMCall      = "makeGsmCall"
	SetGSMSignal     = "setGsmSignal"
	SetGSMVoice      = "setGsmVoice"
	SetNetworkSpeed  = "setNetworkSpeed"
	SetPowerCapacity = "setPowerCapacity"
	SetPowerAC       = "setPowerAc"

	// iOS
	TouchID                 = "touchId"
	ToggleTouchIDEnrollment = "toggleTouchIdEnrollment"
)
