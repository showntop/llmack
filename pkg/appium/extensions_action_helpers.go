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
	"time"

	"github.com/tebeka/selenium"
)

// ActionHelpers provides touch action functionality for Appium WebDriver
type ActionHelpers struct {
	driver *WebDriver
}

// NewActionHelpers creates a new ActionHelpers instance
func NewActionHelpers(driver *WebDriver) *ActionHelpers {
	return &ActionHelpers{
		driver: driver,
	}
}

// Position represents a coordinate position
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Scroll scrolls from one element to another using selenium PerformActions
func (ah *ActionHelpers) Scroll(originEl, destinationEl selenium.WebElement, duration int) error {
	if duration <= 0 {
		duration = 600
	}

	originLoc, err := originEl.Location()
	if err != nil {
		return fmt.Errorf("failed to get origin element location: %w", err)
	}

	originSize, err := originEl.Size()
	if err != nil {
		return fmt.Errorf("failed to get origin element size: %w", err)
	}

	destLoc, err := destinationEl.Location()
	if err != nil {
		return fmt.Errorf("failed to get destination element location: %w", err)
	}

	destSize, err := destinationEl.Size()
	if err != nil {
		return fmt.Errorf("failed to get destination element size: %w", err)
	}

	startX := originLoc.X + originSize.Width/2
	startY := originLoc.Y + originSize.Height/2
	endX := destLoc.X + destSize.Width/2
	endY := destLoc.Y + destSize.Height/2

	return ah.Swipe(startX, startY, endX, endY, duration)
}

// DragAndDrop drags the origin element to the destination element using selenium PerformActions
func (ah *ActionHelpers) DragAndDrop(originEl, destinationEl selenium.WebElement, pause *float64) error {
	originLoc, err := originEl.Location()
	if err != nil {
		return fmt.Errorf("failed to get origin element location: %w", err)
	}

	originSize, err := originEl.Size()
	if err != nil {
		return fmt.Errorf("failed to get origin element size: %w", err)
	}

	destLoc, err := destinationEl.Location()
	if err != nil {
		return fmt.Errorf("failed to get destination element location: %w", err)
	}

	destSize, err := destinationEl.Size()
	if err != nil {
		return fmt.Errorf("failed to get destination element size: %w", err)
	}

	startX := originLoc.X + originSize.Width/2
	startY := originLoc.Y + originSize.Height/2
	endX := destLoc.X + destSize.Width/2
	endY := destLoc.Y + destSize.Height/2

	// Store pointer actions using selenium WebDriver methods
	ah.driver.StorePointerActions("finger1", selenium.TouchPointer,
		selenium.PointerMoveAction(0, selenium.Point{X: startX, Y: startY}, selenium.FromViewport),
		selenium.PointerDownAction(selenium.LeftButton),
	)

	// Add pause if specified
	if pause != nil && *pause > 0 {
		pauseDuration := time.Duration(*pause * float64(time.Second))
		ah.driver.StorePointerActions("finger1", selenium.TouchPointer,
			selenium.PointerPauseAction(pauseDuration),
		)
	}

	// Move to destination and release
	ah.driver.StorePointerActions("finger1", selenium.TouchPointer,
		selenium.PointerMoveAction(200*time.Millisecond, selenium.Point{X: endX, Y: endY}, selenium.FromViewport),
		selenium.PointerUpAction(selenium.LeftButton),
	)

	// Perform the actions
	err = ah.driver.PerformActions()
	if err != nil {
		return fmt.Errorf("failed to perform drag and drop actions: %w", err)
	}

	// Release actions
	return ah.driver.ReleaseActions()
}

// Tap taps on particular places with up to five fingers using selenium PerformActions
func (ah *ActionHelpers) Tap(positions []Position, duration int) error {
	if len(positions) == 0 {
		return fmt.Errorf("positions cannot be empty")
	}

	if len(positions) > 5 {
		return fmt.Errorf("maximum 5 fingers supported, got %d", len(positions))
	}

	if duration <= 0 {
		duration = 100
	}

	tapDuration := time.Duration(duration) * time.Millisecond

	// Store actions for each finger
	for i, pos := range positions {
		fingerID := fmt.Sprintf("finger%d", i+1)
		ah.driver.StorePointerActions(fingerID, selenium.TouchPointer,
			selenium.PointerMoveAction(0, selenium.Point{X: pos.X, Y: pos.Y}, selenium.FromViewport),
			selenium.PointerDownAction(selenium.LeftButton),
			selenium.PointerPauseAction(tapDuration),
			selenium.PointerUpAction(selenium.LeftButton),
		)
	}

	// Perform the actions
	err := ah.driver.PerformActions()
	if err != nil {
		return fmt.Errorf("failed to perform tap actions: %w", err)
	}

	// Release actions
	return ah.driver.ReleaseActions()
}

// Swipe swipes from one point to another point using selenium PerformActions
func (ah *ActionHelpers) Swipe(startX, startY, endX, endY, duration int) error {
	if duration <= 0 {
		duration = 300
	}

	moveDuration := time.Duration(duration) * time.Millisecond

	// Store swipe actions
	ah.driver.StorePointerActions("finger1", selenium.TouchPointer,
		selenium.PointerMoveAction(0, selenium.Point{X: startX, Y: startY}, selenium.FromViewport),
		selenium.PointerDownAction(selenium.LeftButton),
		selenium.PointerMoveAction(moveDuration, selenium.Point{X: endX, Y: endY}, selenium.FromViewport),
		selenium.PointerUpAction(selenium.LeftButton),
	)

	// Perform the actions
	err := ah.driver.PerformActions()
	if err != nil {
		return fmt.Errorf("failed to perform swipe actions: %w", err)
	}

	// Release actions
	return ah.driver.ReleaseActions()
}

// Flick flicks from one point to another point using selenium PerformActions
func (ah *ActionHelpers) Flick(startX, startY, endX, endY int) error {
	// Store flick actions (fast movement)
	ah.driver.StorePointerActions("finger1", selenium.TouchPointer,
		selenium.PointerMoveAction(0, selenium.Point{X: startX, Y: startY}, selenium.FromViewport),
		selenium.PointerDownAction(selenium.LeftButton),
		selenium.PointerMoveAction(50*time.Millisecond, selenium.Point{X: endX, Y: endY}, selenium.FromViewport),
		selenium.PointerUpAction(selenium.LeftButton),
	)

	// Perform the actions
	err := ah.driver.PerformActions()
	if err != nil {
		return fmt.Errorf("failed to perform flick actions: %w", err)
	}

	// Release actions
	return ah.driver.ReleaseActions()
}

// LongPress performs a long press on the specified coordinates using selenium PerformActions
func (ah *ActionHelpers) LongPress(x, y int, duration int) error {
	if duration <= 0 {
		duration = 1000
	}

	pressDuration := time.Duration(duration) * time.Millisecond

	// Store long press actions
	ah.driver.StorePointerActions("finger1", selenium.TouchPointer,
		selenium.PointerMoveAction(0, selenium.Point{X: x, Y: y}, selenium.FromViewport),
		selenium.PointerDownAction(selenium.LeftButton),
		selenium.PointerPauseAction(pressDuration),
		selenium.PointerUpAction(selenium.LeftButton),
	)

	// Perform the actions
	err := ah.driver.PerformActions()
	if err != nil {
		return fmt.Errorf("failed to perform long press actions: %w", err)
	}

	// Release actions
	return ah.driver.ReleaseActions()
}

// LongPressElement performs a long press on the specified element using selenium PerformActions
func (ah *ActionHelpers) LongPressElement(element selenium.WebElement, duration int) error {
	loc, err := element.Location()
	if err != nil {
		return fmt.Errorf("failed to get element location: %w", err)
	}

	size, err := element.Size()
	if err != nil {
		return fmt.Errorf("failed to get element size: %w", err)
	}

	centerX := loc.X + size.Width/2
	centerY := loc.Y + size.Height/2

	return ah.LongPress(centerX, centerY, duration)
}

// Pinch performs a pinch gesture (zoom out) using selenium PerformActions
func (ah *ActionHelpers) Pinch(centerX, centerY, distance int, duration int) error {
	if duration <= 0 {
		duration = 600
	}

	moveDuration := time.Duration(duration) * time.Millisecond

	startX1 := centerX - distance/2
	startY1 := centerY
	endX1 := centerX - distance/4
	endY1 := centerY

	startX2 := centerX + distance/2
	startY2 := centerY
	endX2 := centerX + distance/4
	endY2 := centerY

	// Store actions for first finger
	ah.driver.StorePointerActions("finger1", selenium.TouchPointer,
		selenium.PointerMoveAction(0, selenium.Point{X: startX1, Y: startY1}, selenium.FromViewport),
		selenium.PointerDownAction(selenium.LeftButton),
		selenium.PointerMoveAction(moveDuration, selenium.Point{X: endX1, Y: endY1}, selenium.FromViewport),
		selenium.PointerUpAction(selenium.LeftButton),
	)

	// Store actions for second finger
	ah.driver.StorePointerActions("finger2", selenium.TouchPointer,
		selenium.PointerMoveAction(0, selenium.Point{X: startX2, Y: startY2}, selenium.FromViewport),
		selenium.PointerDownAction(selenium.LeftButton),
		selenium.PointerMoveAction(moveDuration, selenium.Point{X: endX2, Y: endY2}, selenium.FromViewport),
		selenium.PointerUpAction(selenium.LeftButton),
	)

	// Perform the actions
	err := ah.driver.PerformActions()
	if err != nil {
		return fmt.Errorf("failed to perform pinch actions: %w", err)
	}

	// Release actions
	return ah.driver.ReleaseActions()
}

// Zoom performs a zoom gesture (zoom in) using selenium PerformActions
func (ah *ActionHelpers) Zoom(centerX, centerY, distance int, duration int) error {
	if duration <= 0 {
		duration = 600
	}

	moveDuration := time.Duration(duration) * time.Millisecond

	startX1 := centerX - distance/4
	startY1 := centerY
	endX1 := centerX - distance/2
	endY1 := centerY

	startX2 := centerX + distance/4
	startY2 := centerY
	endX2 := centerX + distance/2
	endY2 := centerY

	// Store actions for first finger
	ah.driver.StorePointerActions("finger1", selenium.TouchPointer,
		selenium.PointerMoveAction(0, selenium.Point{X: startX1, Y: startY1}, selenium.FromViewport),
		selenium.PointerDownAction(selenium.LeftButton),
		selenium.PointerMoveAction(moveDuration, selenium.Point{X: endX1, Y: endY1}, selenium.FromViewport),
		selenium.PointerUpAction(selenium.LeftButton),
	)

	// Store actions for second finger
	ah.driver.StorePointerActions("finger2", selenium.TouchPointer,
		selenium.PointerMoveAction(0, selenium.Point{X: startX2, Y: startY2}, selenium.FromViewport),
		selenium.PointerDownAction(selenium.LeftButton),
		selenium.PointerMoveAction(moveDuration, selenium.Point{X: endX2, Y: endY2}, selenium.FromViewport),
		selenium.PointerUpAction(selenium.LeftButton),
	)

	// Perform the actions
	err := ah.driver.PerformActions()
	if err != nil {
		return fmt.Errorf("failed to perform zoom actions: %w", err)
	}

	// Release actions
	return ah.driver.ReleaseActions()
}

// Legacy TouchAction support for backward compatibility
type TouchAction struct {
	Action   string                 `json:"action"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Element  string                 `json:"element,omitempty"`
	X        *int                   `json:"x,omitempty"`
	Y        *int                   `json:"y,omitempty"`
	Duration *int                   `json:"ms,omitempty"`
}

type MultiTouchAction struct {
	Actions [][]TouchAction `json:"actions"`
}
