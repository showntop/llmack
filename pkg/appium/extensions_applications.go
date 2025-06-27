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

// Applications provides application management functionality
// This corresponds to the Python Applications extension
type Applications struct {
	wd *WebDriver
}

// NewApplications creates a new Applications extension
func NewApplications(wd *WebDriver) *Applications {
	return &Applications{wd: wd}
}

// BackgroundApp puts the application in the background on the device for a certain duration
func (app *Applications) BackgroundApp(seconds int) error {
	extName := "mobile: backgroundApp"
	args := map[string]interface{}{
		"seconds": seconds,
	}

	// Try the new extension first
	if err := app.wd.AssertExtensionExists(extName); err == nil {
		_, err := app.wd.ExecuteScript(extName, args)
		return err
	}

	// Fallback to old command
	app.wd.MarkExtensionAbsence(extName)
	_, err := app.wd.Execute(Background, args)
	return err
}

// IsAppInstalled checks whether the application specified by bundleID is installed on the device
func (app *Applications) IsAppInstalled(bundleID string) (bool, error) {
	extName := "mobile: isAppInstalled"
	args := map[string]interface{}{
		"bundleId": bundleID,
		"appId":    bundleID,
	}

	// Try the new extension first
	if err := app.wd.AssertExtensionExists(extName); err == nil {
		result, err := app.wd.ExecuteScript(extName, args)
		if err != nil {
			return false, err
		}
		if boolResult, ok := result.(bool); ok {
			return boolResult, nil
		}
		return false, fmt.Errorf("unexpected response type")
	}

	// Fallback to old command
	app.wd.MarkExtensionAbsence(extName)
	result, err := app.wd.Execute(IsAppInstalled, map[string]interface{}{
		"bundleId": bundleID,
	})
	if err != nil {
		return false, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		if value, exists := resultMap["value"]; exists {
			if boolValue, ok := value.(bool); ok {
				return boolValue, nil
			}
		}
	}

	return false, fmt.Errorf("unexpected response format")
}

// InstallApp installs the application found at appPath on the device
func (app *Applications) InstallApp(appPath string, options map[string]interface{}) error {
	extName := "mobile: installApp"
	args := map[string]interface{}{
		"app":     appPath,
		"appPath": appPath,
	}

	// Merge options
	if options != nil {
		for k, v := range options {
			args[k] = v
		}
	}

	// Try the new extension first
	if err := app.wd.AssertExtensionExists(extName); err == nil {
		_, err := app.wd.ExecuteScript(extName, args)
		return err
	}

	// Fallback to old command
	app.wd.MarkExtensionAbsence(extName)
	data := map[string]interface{}{
		"appPath": appPath,
	}
	if options != nil {
		data["options"] = options
	}

	_, err := app.wd.Execute(InstallApp, data)
	return err
}

// RemoveApp removes the specified application from the device
func (app *Applications) RemoveApp(appID string, options map[string]interface{}) error {
	extName := "mobile: removeApp"
	args := map[string]interface{}{
		"appId":    appID,
		"bundleId": appID,
	}

	// Merge options
	if options != nil {
		for k, v := range options {
			args[k] = v
		}
	}

	// Try the new extension first
	if err := app.wd.AssertExtensionExists(extName); err == nil {
		_, err := app.wd.ExecuteScript(extName, args)
		return err
	}

	// Fallback to old command
	app.wd.MarkExtensionAbsence(extName)
	data := map[string]interface{}{
		"appId": appID,
	}
	if options != nil {
		data["options"] = options
	}

	_, err := app.wd.Execute(RemoveApp, data)
	return err
}

// TerminateApp terminates the application if it is running
func (app *Applications) TerminateApp(appID string, options map[string]interface{}) (bool, error) {
	extName := "mobile: terminateApp"
	args := map[string]interface{}{
		"appId":    appID,
		"bundleId": appID,
	}

	// Merge options
	if options != nil {
		for k, v := range options {
			args[k] = v
		}
	}

	// Try the new extension first
	if err := app.wd.AssertExtensionExists(extName); err == nil {
		result, err := app.wd.ExecuteScript(extName, args)
		if err != nil {
			return false, err
		}
		if boolResult, ok := result.(bool); ok {
			return boolResult, nil
		}
		return false, fmt.Errorf("unexpected response type")
	}

	// Fallback to old command
	app.wd.MarkExtensionAbsence(extName)
	data := map[string]interface{}{
		"appId": appID,
	}
	if options != nil {
		data["options"] = options
	}

	result, err := app.wd.Execute(TerminateApp, data)
	if err != nil {
		return false, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		if value, exists := resultMap["value"]; exists {
			if boolValue, ok := value.(bool); ok {
				return boolValue, nil
			}
		}
	}

	return false, fmt.Errorf("unexpected response format")
}

// ActivateApp activates the application if it is not running or is running in the background
func (app *Applications) ActivateApp(appID string) error {
	extName := "mobile: activateApp"
	args := map[string]interface{}{
		"appId":    appID,
		"bundleId": appID,
	}

	// Try the new extension first
	if err := app.wd.AssertExtensionExists(extName); err == nil {
		_, err := app.wd.ExecuteScript(extName, args)
		return err
	}

	// Fallback to old command
	app.wd.MarkExtensionAbsence(extName)
	_, err := app.wd.Execute(ActivateApp, map[string]interface{}{
		"appId": appID,
	})
	return err
}

// QueryAppState queries the state of the application
func (app *Applications) QueryAppState(appID string) (int, error) {
	extName := "mobile: queryAppState"
	args := map[string]interface{}{
		"appId":    appID,
		"bundleId": appID,
	}

	// Try the new extension first
	if err := app.wd.AssertExtensionExists(extName); err == nil {
		result, err := app.wd.ExecuteScript(extName, args)
		if err != nil {
			return 0, err
		}
		if intResult, ok := result.(int); ok {
			return intResult, nil
		}
		if floatResult, ok := result.(float64); ok {
			return int(floatResult), nil
		}
		return 0, fmt.Errorf("unexpected response type")
	}

	// Fallback to old command
	app.wd.MarkExtensionAbsence(extName)
	result, err := app.wd.Execute(QueryAppState, map[string]interface{}{
		"appId": appID,
	})
	if err != nil {
		return 0, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		if value, exists := resultMap["value"]; exists {
			if intValue, ok := value.(int); ok {
				return intValue, nil
			}
			if floatValue, ok := value.(float64); ok {
				return int(floatValue), nil
			}
		}
	}

	return 0, fmt.Errorf("unexpected response format")
}

// GetAppStrings returns the application strings from the device for the specified language
func (app *Applications) GetAppStrings(language string, stringFile string) (map[string]string, error) {
	extName := "mobile: getAppStrings"
	args := map[string]interface{}{}

	if language != "" {
		args["language"] = language
	}
	if stringFile != "" {
		args["stringFile"] = stringFile
	}

	// Try the new extension first
	if err := app.wd.AssertExtensionExists(extName); err == nil {
		result, err := app.wd.ExecuteScript(extName, args)
		if err != nil {
			return nil, err
		}
		if mapResult, ok := result.(map[string]string); ok {
			return mapResult, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	}

	// Fallback to old command
	app.wd.MarkExtensionAbsence(extName)
	params := map[string]interface{}{}
	if language != "" {
		params["language"] = language
	}
	if stringFile != "" {
		params["stringFile"] = stringFile
	}

	result, err := app.wd.Execute(GetAppStrings, params)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		if value, exists := resultMap["value"]; exists {
			if stringMap, ok := value.(map[string]string); ok {
				return stringMap, nil
			}
		}
	}

	return nil, fmt.Errorf("unexpected response format")
}

// ExecuteScript is a placeholder for script execution
// In a full implementation, this would call the WebDriver's ExecuteScript method
func (wd *WebDriver) ExecuteScript(script string, args map[string]interface{}) (interface{}, error) {
	// This is a placeholder implementation
	// In reality, this would execute JavaScript/mobile scripts on the device
	return nil, fmt.Errorf("ExecuteScript not implemented")
}
