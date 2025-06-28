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

	"github.com/tebeka/selenium"
)

// Mock element for testing
type mockElement struct {
	x, y, width, height int
}

func (m *mockElement) Click() error {
	return nil
}

func (m *mockElement) SendKeys(keys string) error {
	return nil
}

func (m *mockElement) TagName() (string, error) {
	return "div", nil
}

func (m *mockElement) Text() (string, error) {
	return "test", nil
}

func (m *mockElement) IsEnabled() (bool, error) {
	return true, nil
}

func (m *mockElement) IsSelected() (bool, error) {
	return false, nil
}

func (m *mockElement) IsDisplayed() (bool, error) {
	return true, nil
}

func (m *mockElement) GetAttribute(name string) (string, error) {
	return "", nil
}

func (m *mockElement) GetProperty(name string) (string, error) {
	return "", nil
}

func (m *mockElement) Location() (*selenium.Point, error) {
	return &selenium.Point{X: m.x, Y: m.y}, nil
}

func (m *mockElement) LocationInView() (*selenium.Point, error) {
	return &selenium.Point{X: m.x, Y: m.y}, nil
}

func (m *mockElement) Size() (*selenium.Size, error) {
	return &selenium.Size{Width: m.width, Height: m.height}, nil
}

func (m *mockElement) CSSProperty(name string) (string, error) {
	return "", nil
}

func (m *mockElement) Screenshot(scroll bool) ([]byte, error) {
	return []byte{}, nil
}

func (m *mockElement) Clear() error {
	return nil
}

func (m *mockElement) FindElement(by, value string) (selenium.WebElement, error) {
	return m, nil
}

func (m *mockElement) FindElements(by, value string) ([]selenium.WebElement, error) {
	return []selenium.WebElement{m}, nil
}

func (m *mockElement) MoveTo(xOffset, yOffset int) error {
	return nil
}

func (m *mockElement) Submit() error {
	return nil
}

// Mock WebDriver for testing
type mockWebDriver struct {
	executedCommands []string
	executedParams   []map[string]interface{}
}

func newMockWebDriver() *mockWebDriver {
	return &mockWebDriver{
		executedCommands: []string{},
		executedParams:   []map[string]interface{}{},
	}
}

func (m *mockWebDriver) Execute(command string, params map[string]interface{}) (interface{}, error) {
	m.executedCommands = append(m.executedCommands, command)
	m.executedParams = append(m.executedParams, params)
	return nil, nil
}

// TestActionHelpers tests ActionHelpers creation
func TestActionHelpers(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	// TODO: Fix test setup - driver.Execute assignment issue
	// driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)
	if ah == nil {
		t.Fatal("ActionHelpers should not be nil")
	}

	if ah.driver != driver {
		t.Error("ActionHelpers driver should match the provided driver")
	}
}

// TestActionHelpersNilDriver tests ActionHelpers with nil driver
func TestActionHelpersNilDriver(t *testing.T) {
	ah := NewActionHelpers(nil)

	err := ah.Tap([]Position{{X: 100, Y: 100}}, 100)
	if err == nil {
		t.Error("Expected error for nil driver")
	}
}

// TestTapSingleFinger tests single finger tap using W3C Actions
func TestTapSingleFinger(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)

	positions := []Position{{X: 100, Y: 200}}
	err := ah.Tap(positions, 500)

	if err != nil {
		t.Fatalf("Tap failed: %v", err)
	}

	if len(mockDriver.executedCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mockDriver.executedCommands))
	}

	if mockDriver.executedCommands[0] != "performActions" {
		t.Errorf("Expected performActions command, got %s", mockDriver.executedCommands[0])
	}
}

// TestTapMultipleFinger tests multi-finger tap using W3C Actions
func TestTapMultipleFinger(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)

	positions := []Position{{X: 100, Y: 200}, {X: 300, Y: 400}}
	err := ah.Tap(positions, 500)

	if err != nil {
		t.Fatalf("Multi-finger tap failed: %v", err)
	}

	if len(mockDriver.executedCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mockDriver.executedCommands))
	}

	if mockDriver.executedCommands[0] != "performActions" {
		t.Errorf("Expected performActions command, got %s", mockDriver.executedCommands[0])
	}
}

// TestSwipe tests swipe gesture using W3C Actions
func TestSwipe(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)

	err := ah.Swipe(100, 200, 300, 400, 600)

	if err != nil {
		t.Fatalf("Swipe failed: %v", err)
	}

	if len(mockDriver.executedCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mockDriver.executedCommands))
	}

	if mockDriver.executedCommands[0] != "performActions" {
		t.Errorf("Expected performActions command, got %s", mockDriver.executedCommands[0])
	}
}

// TestLongPress tests long press gesture using W3C Actions
func TestLongPress(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)

	err := ah.LongPress(100, 200, 1500)

	if err != nil {
		t.Fatalf("LongPress failed: %v", err)
	}

	if len(mockDriver.executedCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mockDriver.executedCommands))
	}

	if mockDriver.executedCommands[0] != "performActions" {
		t.Errorf("Expected performActions command, got %s", mockDriver.executedCommands[0])
	}
}

// TestPinch tests pinch gesture using W3C Actions
func TestPinch(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)

	err := ah.Pinch(200, 200, 400, 800)

	if err != nil {
		t.Fatalf("Pinch failed: %v", err)
	}

	if len(mockDriver.executedCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mockDriver.executedCommands))
	}

	if mockDriver.executedCommands[0] != "performActions" {
		t.Errorf("Expected performActions command, got %s", mockDriver.executedCommands[0])
	}
}

// TestZoom tests zoom gesture using W3C Actions
func TestZoom(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)

	err := ah.Zoom(200, 200, 400, 800)

	if err != nil {
		t.Fatalf("Zoom failed: %v", err)
	}

	if len(mockDriver.executedCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mockDriver.executedCommands))
	}

	if mockDriver.executedCommands[0] != "performActions" {
		t.Errorf("Expected performActions command, got %s", mockDriver.executedCommands[0])
	}
}

// TestDragAndDrop tests drag and drop gesture using W3C Actions
func TestDragAndDrop(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)

	originEl := &mockElement{x: 100, y: 100, width: 50, height: 50}
	destEl := &mockElement{x: 300, y: 300, width: 50, height: 50}

	pause := 0.5
	err := ah.DragAndDrop(originEl, destEl, &pause)

	if err != nil {
		t.Fatalf("DragAndDrop failed: %v", err)
	}

	if len(mockDriver.executedCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mockDriver.executedCommands))
	}

	if mockDriver.executedCommands[0] != "performActions" {
		t.Errorf("Expected performActions command, got %s", mockDriver.executedCommands[0])
	}
}

// TestScroll tests scroll gesture using W3C Actions
func TestScroll(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)

	originEl := &mockElement{x: 100, y: 100, width: 50, height: 50}
	destEl := &mockElement{x: 100, y: 400, width: 50, height: 50}

	err := ah.Scroll(originEl, destEl, 800)

	if err != nil {
		t.Fatalf("Scroll failed: %v", err)
	}

	if len(mockDriver.executedCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mockDriver.executedCommands))
	}

	if mockDriver.executedCommands[0] != "performActions" {
		t.Errorf("Expected performActions command, got %s", mockDriver.executedCommands[0])
	}
}

// TestW3CActionBuilder tests W3C action builder
func TestW3CActionBuilder(t *testing.T) {
	ah := &ActionHelpers{}

	action := ah.CreateTouchAction("test-finger")
	if action.Type != "pointer" {
		t.Errorf("Expected pointer type, got %s", action.Type)
	}

	if action.ID != "test-finger" {
		t.Errorf("Expected test-finger ID, got %s", action.ID)
	}

	action.AddPointerMove(100, 200, 300).AddPointerDown(nil).AddPause(500).AddPointerUp(nil)

	if len(action.Actions) != 4 {
		t.Errorf("Expected 4 actions, got %d", len(action.Actions))
	}

	// Check first action (pointer move)
	if action.Actions[0].Type != "pointerMove" {
		t.Errorf("Expected pointerMove, got %s", action.Actions[0].Type)
	}

	if *action.Actions[0].X != 100 || *action.Actions[0].Y != 200 {
		t.Errorf("Expected coordinates (100, 200), got (%d, %d)", *action.Actions[0].X, *action.Actions[0].Y)
	}

	// Check second action (pointer down)
	if action.Actions[1].Type != "pointerDown" {
		t.Errorf("Expected pointerDown, got %s", action.Actions[1].Type)
	}

	// Check third action (pause)
	if action.Actions[2].Type != "pause" {
		t.Errorf("Expected pause, got %s", action.Actions[2].Type)
	}

	if *action.Actions[2].Duration != 500 {
		t.Errorf("Expected duration 500, got %d", *action.Actions[2].Duration)
	}

	// Check fourth action (pointer up)
	if action.Actions[3].Type != "pointerUp" {
		t.Errorf("Expected pointerUp, got %s", action.Actions[3].Type)
	}
}

// TestTapValidation tests input validation for Tap
func TestTapValidation(t *testing.T) {
	ah := NewActionHelpers(nil)

	// Test empty positions
	err := ah.Tap([]Position{}, 100)
	if err == nil {
		t.Error("Expected error for empty positions")
	}

	// Test too many positions (more than 5 fingers)
	positions := make([]Position, 6)
	for i := range positions {
		positions[i] = Position{X: i * 50, Y: i * 50}
	}

	err = ah.Tap(positions, 100)
	if err == nil {
		t.Error("Expected error for more than 5 positions")
	}
}

// TestReleaseActions tests release W3C actions
func TestReleaseActions(t *testing.T) {
	mockDriver := newMockWebDriver()

	driver := &WebDriver{}
	driver.Execute = mockDriver.Execute

	ah := NewActionHelpers(driver)

	err := ah.ReleaseW3CActions()

	if err != nil {
		t.Fatalf("ReleaseW3CActions failed: %v", err)
	}

	if len(mockDriver.executedCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mockDriver.executedCommands))
	}

	if mockDriver.executedCommands[0] != "releaseActions" {
		t.Errorf("Expected releaseActions command, got %s", mockDriver.executedCommands[0])
	}
}

// TestDragAndDropElementValidation tests drag and drop element validation
func TestDragAndDropElementValidation(t *testing.T) {
	ah := NewActionHelpers(nil)

	originEl := &mockElement{x: 100, y: 100, width: 50, height: 50}
	destEl := &mockElement{x: 300, y: 300, width: 50, height: 50}

	// Should return error due to nil driver
	err := ah.DragAndDrop(originEl, destEl, nil)
	if err == nil {
		t.Error("Expected error for nil driver")
	}
}

// TestScrollElementValidation tests scroll element validation
func TestScrollElementValidation(t *testing.T) {
	ah := NewActionHelpers(nil)

	originEl := &mockElement{x: 100, y: 100, width: 50, height: 50}
	destEl := &mockElement{x: 100, y: 400, width: 50, height: 50}

	// Should return error due to nil driver
	err := ah.Scroll(originEl, destEl, 800)
	if err == nil {
		t.Error("Expected error for nil driver")
	}
}

// TestActionBuilderFluency tests the fluent interface of action builder
func TestActionBuilderFluency(t *testing.T) {
	ah := &ActionHelpers{}

	// Test method chaining
	action := ah.CreateTouchAction("fluent-test").
		AddPointerMove(50, 60, 100).
		AddPointerDown(nil).
		AddPause(250).
		AddPointerMove(150, 160, 200).
		AddPointerUp(nil)

	if len(action.Actions) != 5 {
		t.Errorf("Expected 5 chained actions, got %d", len(action.Actions))
	}

	// Verify first move action
	if action.Actions[0].Type != "pointerMove" {
		t.Errorf("Expected first action to be pointerMove, got %s", action.Actions[0].Type)
	}

	if *action.Actions[0].X != 50 || *action.Actions[0].Y != 60 {
		t.Errorf("Expected first move to (50, 60), got (%d, %d)", *action.Actions[0].X, *action.Actions[0].Y)
	}

	// Verify second move action
	if action.Actions[3].Type != "pointerMove" {
		t.Errorf("Expected fourth action to be pointerMove, got %s", action.Actions[3].Type)
	}

	if *action.Actions[3].X != 150 || *action.Actions[3].Y != 160 {
		t.Errorf("Expected second move to (150, 160), got (%d, %d)", *action.Actions[3].X, *action.Actions[3].Y)
	}
}
