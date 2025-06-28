package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/showntop/llmack/llm"
	appiumgo "github.com/showntop/llmack/pkg/appium"
)

type Controller struct {
	model    *llm.Instance
	driver   *appiumgo.WebDriver
	registry *Registry
}

func NewController(driver *appiumgo.WebDriver) *Controller {
	ctrl := &Controller{
		driver:   driver,
		registry: NewRegistry(),
	}
	RegisterAction(ctrl.registry, "press_key_code", "Press key code", ctrl.PressKeyCode)
	RegisterAction(ctrl.registry, "swipe", "Swipe", ctrl.Swipe)
	RegisterAction(ctrl.registry, "tap_by_coordinates", "Tap by coordinates", ctrl.TapByCoordinates)
	return ctrl
}

func (c *Controller) Registry() *Registry {
	return c.registry
}

type ScreenshotAction struct {
	Path string `json:"path"`
}

type PressKeyCodeAction struct {
	KeyCode int `json:"key_code"` // keycode: Android keycode to press(// Common keycodes:\n	- 3: HOME\n	- 4: BACK\n	- 24: VOLUME UP\n	- 25: VOLUME DOWN\n	- 26: POWER\n	- 82: MENU)

}

func (c *Controller) PressKeyCode(ctx context.Context, params PressKeyCodeAction) (*ActionResult, error) {
	if err := c.driver.KeyDown(strconv.Itoa(params.KeyCode)); err != nil {
		return nil, err
	}
	return NewActionResult(), nil
}

type SwipeAction struct {
	StartX   int `json:"start_x"`
	StartY   int `json:"start_y"`
	EndX     int `json:"end_x"`
	EndY     int `json:"end_y"`
	Duration int `json:"duration"`
}

func (c *Controller) Swipe(ctx context.Context, params SwipeAction) (*ActionResult, error) {
	if err := c.driver.Swipe(params.StartX, params.StartY, params.EndX, params.EndY, params.Duration); err != nil {
		return nil, err
	}
	return NewActionResult(), nil
}

type LaunchAppAction struct {
	PackageName  string `json:"package_name"`  // Package name (e.g., "com.android.settings")
	ActivityName string `json:"activity_name"` // Optional activity name
}

func (c *Controller) LaunchApp(ctx context.Context, params LaunchAppAction) (*ActionResult, error) {
	panic("not implemented")
}

type TapByCoordinatesAction struct {
	X        int `json:"x" jsonschema:"description=the percentage of the screen width from the left,required"`
	Y        int `json:"y" jsonschema:"description=the percentage of the screen height from the top,required"`
	Duration int `json:"duration" jsonschema:"description=duration of the tap,required"`
}

func (c *Controller) TapByCoordinates(ctx context.Context, params TapByCoordinatesAction) (*ActionResult, error) {

	if err := c.driver.Tap([]appiumgo.Position{{X: params.X, Y: params.Y}}, params.Duration); err != nil {
		return nil, err
	}
	return NewActionResult(), nil
}

// ExecuteAction TODO
func (c *Controller) ExecuteAction(
	ctx context.Context,
	action map[string]any,
	driver *appiumgo.WebDriver,
	model *llm.Instance,
	sensitiveData map[string]string,
	availableFilePaths []string,
	// context: Context | None,
) (*ActionResult, error) {
	for name, params := range action {
		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(params)
		if err != nil {
			return nil, err
		}
		ab := buffer.Bytes()
		if len(ab) > 0 && ab[len(ab)-1] == '\n' {
			ab = ab[:len(ab)-1]
		}
		result, err := c.Registry().ExecuteAction(ctx, name, string(ab), sensitiveData)
		if err != nil {
			return nil, err
		}
		var actionResult ActionResult
		err = json.Unmarshal([]byte(result), &actionResult)
		if err != nil {
			return nil, err
		}
		return &actionResult, nil
	}
	return NewActionResult(), nil
}

type ActionResult struct {
	IsDone           *bool   `json:"is_done,omitempty"`
	Success          *bool   `json:"success,omitempty"`
	ExtractedContent *string `json:"extracted_content,omitempty"`
	Error            *string `json:"error,omitempty"`
	IncludeInMemory  bool    `json:"include_in_memory"`
}

func NewActionResult() *ActionResult {
	success := true
	return &ActionResult{
		Success:          &success,
		ExtractedContent: nil,
		Error:            nil,
		// IncludeInMemory:  false,
	}
}
