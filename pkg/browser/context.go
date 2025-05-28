package browser

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/playwright-community/playwright-go"
	"github.com/showntop/llmack/pkg/browser/dom"
	"github.com/showntop/llmack/pkg/structx"
)

type CachedStateClickableElementsHashes struct {
	Url    string
	Hashes []string
}

type BrowserSession struct {
	ActiveTab                          playwright.Page
	Context                            playwright.BrowserContext
	CachedState                        *BrowserState
	CachedStateClickableElementsHashes *CachedStateClickableElementsHashes
}

func NewSession(context playwright.BrowserContext, cachedState *BrowserState) *BrowserSession {

	browserSession := BrowserSession{
		ActiveTab:                          nil,
		Context:                            context,
		CachedState:                        cachedState,
		CachedStateClickableElementsHashes: nil,
	}

	browserSession.Context.OnPage(func(page playwright.Page) {
		initScript := `
			(() => {
				if (!window.getEventListeners) {
					window.getEventListeners = function (node) {
						return node.__listeners || {};
					};

					// Save the original addEventListener
					const originalAddEventListener = Element.prototype.addEventListener;

					const eventProxy = {
						addEventListener: function (type, listener, options = {}) {
							// Initialize __listeners if not exists
							const defaultOptions = { once: false, passive: false, capture: false };
							if(typeof options === 'boolean') {
								options = { capture: options };
							}
							options = { ...defaultOptions, ...options };
							if (!this.__listeners) {
								this.__listeners = {};
							}

							// Initialize array for this event type if not exists
							if (!this.__listeners[type]) {
								this.__listeners[type] = [];
							}
							

							// Add the listener to __listeners
							this.__listeners[type].push({
								listener: listener,
								type: type,
								...options
							});

							// Call original addEventListener using the saved reference
							return originalAddEventListener.call(this, type, listener, options);
						}
					};

					Element.prototype.addEventListener = eventProxy.addEventListener;
				}
			})()`
		page.AddInitScript(playwright.Script{Content: &initScript})
	})

	return &browserSession
}

// State of the browser context
type BrowserContextState struct {
	TargetId *string `json:"target_id,omitempty"`
}

type BrowserContext struct {
	ContextId        string
	Config           BrowserConfig
	Browser          *Browser
	Session          *BrowserSession
	State            *BrowserContextState
	ActiveTab        playwright.Page
	pageEventHandler func(page playwright.Page)
}

func (bc *BrowserContext) ConvertSimpleXpathToCssSelector(xpath string) string {
	return dom.ConvertSimpleXpathToCssSelector(xpath)
}

func (bc *BrowserContext) EnhancedCssSelectorForElement(element *dom.DOMElementNode, includeDynamicAttributes bool) string {
	return dom.EnhancedCssSelectorForElement(element, includeDynamicAttributes)
}

/*
# GetState Get the current state of the browser

cache_clickable_elements_hashes: bool
If True, cache the clickable elements hashes for the current state. This is used to calculate which elements are new to the llm (from last message) -> reduces token usage.
*/
func (bc *BrowserContext) GetState(cacheClickableElementsHashes bool) *BrowserState {
	bc.waitForPageAndFramesLoad(nil)
	page := bc.GetCurrentPage()

	session := bc.GetSession()
	updatedState := bc.getUpdatedState(page)

	if cacheClickableElementsHashes {
		clickableElementProcessor := &dom.ClickableElementProcessor{}
		if session.CachedStateClickableElementsHashes != nil && session.CachedStateClickableElementsHashes.Url == updatedState.Url {
			updatedStateClickableElements := clickableElementProcessor.GetClickableElements(updatedState.ElementTree)

			for _, domElement := range updatedStateClickableElements {
				domElement.IsNew = playwright.Bool(!slices.Contains(session.CachedStateClickableElementsHashes.Hashes, clickableElementProcessor.HashDomElement(domElement)))
			}
		}
		session.CachedStateClickableElementsHashes = &CachedStateClickableElementsHashes{
			Url:    updatedState.Url,
			Hashes: clickableElementProcessor.GetClickableElementsHashes(updatedState.ElementTree),
		}
	}
	session.CachedState = updatedState

	// TODO(MID): Save cookies if a file is specified
	// if bc.Config.CookiesFile != "" {
	// 	bc.SaveCookies()
	// }

	return updatedState
}

func (bc *BrowserContext) getUpdatedState(page playwright.Page) *BrowserState {
	domService := dom.NewDomService(page)
	focus_element := -1 // default
	content, err := domService.GetClickableElements(
		structx.GetDefaultValue(bc.Config, "highlight_elements", true),
		focus_element,
		structx.GetDefaultValue(bc.Config, "viewport_expansion", 0),
	)
	if err != nil {
		log.Warnf("Failed to get clickable elements: %s", err)
	}

	tabsInfo := bc.GetTabsInfo()

	screenshot, err := bc.TakeScreenshot(false)
	if err != nil {
		log.Warnf("Failed to take screenshot: %s", err)
	}
	pixelsAbove, pixelsBelow, err := bc.GetScrollInfo(page)
	if err != nil {
		log.Warnf("Failed to get scroll info: %s", err)
	}

	title, _ := page.Title()
	// updated_state
	currentState := BrowserState{
		ElementTree:   content.ElementTree,
		SelectorMap:   content.SelectorMap,
		Url:           page.URL(),
		Title:         title,
		Tabs:          tabsInfo,
		Screenshot:    screenshot,
		PixelAbove:    pixelsAbove,
		PixelBelow:    pixelsBelow,
		BrowserErrors: []string{},
	}
	return &currentState
}

// Returns a base64 encoded screenshot of the current page.
func (bc *BrowserContext) TakeScreenshot(fullPage bool) (*string, error) {
	page := bc.GetCurrentPage()

	err := page.BringToFront()
	if err != nil {
		return nil, err
	}

	err = page.WaitForLoadState()
	if err != nil {
		return nil, err
	}

	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{FullPage: playwright.Bool(fullPage), Animations: playwright.ScreenshotAnimationsDisabled})
	if err != nil {
		return nil, err
	}

	screenshotBase64 := base64.StdEncoding.EncodeToString(screenshot)
	return &screenshotBase64, nil
}

// Get scroll position information for the current page.
func (bc *BrowserContext) GetScrollInfo(page playwright.Page) (int, int, error) {
	scrollY, err := page.Evaluate("() => window.scrollY")
	if err != nil {
		return 0, 0, err
	}
	viewportHeight, err := page.Evaluate("() => window.innerHeight")
	if err != nil {
		return 0, 0, err
	}
	totalHeight, err := page.Evaluate("() => document.documentElement.scrollHeight")
	if err != nil {
		return 0, 0, err
	}
	pixelsAbove, err := ParseNumberToInt(scrollY)
	if err != nil {
		return 0, 0, err
	}
	totalHeightInt, err := ParseNumberToInt(totalHeight)
	if err != nil {
		return 0, 0, err
	}
	viewportHeightInt, err := ParseNumberToInt(viewportHeight)
	if err != nil {
		return 0, 0, err
	}
	pixelsBelow := totalHeightInt - (pixelsAbove + viewportHeightInt)
	return pixelsAbove, pixelsBelow, nil
}

func (bc *BrowserContext) GetSession() *BrowserSession {
	if bc.Session == nil {
		session, err := bc.initializeSession()
		if err != nil {
			panic(err)
		}
		return session
	}
	return bc.Session
}

// Get the current page
func (bc *BrowserContext) GetCurrentPage() playwright.Page {
	session := bc.GetSession()
	return bc.getCurrentPage(session)
}

func (bc *BrowserContext) Close() {
	if bc.Session == nil {
		return
	}
	if bc.pageEventHandler != nil && bc.Session.Context != nil {
		bc.Session.Context.RemoveListener("page", bc.pageEventHandler)
		bc.pageEventHandler = nil
	}

	if bc.Config["cookies_file"] != nil {
		go bc.SaveCookies()
	}

	if keepAlive, ok := bc.Config["keep_alive"].(bool); (ok && !keepAlive) || !ok {
		err := bc.Session.Context.Close()
		if err != nil {
			log.Debugf("🪨  Failed to close browser context: %s", err)
		}
	}

	// Dereference everything
	bc.Session = nil
	bc.ActiveTab = nil
	bc.pageEventHandler = nil
}

func (bc *BrowserContext) GetSelectorMap() *dom.SelectorMap {
	session := bc.GetSession()
	if session.CachedState == nil {
		return nil
	}
	return session.CachedState.SelectorMap
}

func (bc *BrowserContext) GetDomElementByIndex(index int) (*dom.DOMElementNode, error) {
	selectorMap := bc.GetSelectorMap()
	if selectorMap == nil || (*selectorMap)[index] == nil {
		return nil, fmt.Errorf("element with index %d does not exist - retry or use alternative actions", index)
	}
	return (*selectorMap)[index], nil
}

// Check if element or its children are file uploaders
func (bc *BrowserContext) IsFileUploader(element *dom.DOMElementNode, maxDepth int, currentDepth int) bool {
	if currentDepth > maxDepth {
		return false
	}
	// reflect.TypeOf(element).Elem().Name() != "DOMElementNode"
	if element == nil {
		return false
	}
	// Check current element
	isUploader := false
	// Check for file input attributes
	if element.TagName == "input" {
		isUploader = element.Attributes["type"] == "file" || element.Attributes["accept"] != ""
	}
	if isUploader {
		return true
	}
	// Recursively check children
	if element.Children != nil && currentDepth < maxDepth {
		for _, child := range element.Children {
			if child, ok := child.(*dom.DOMElementNode); ok {
				if bc.IsFileUploader(child, maxDepth, currentDepth+1) {
					return true
				}
			}
		}
	}
	return false
}

// sync DOMElementNode with Playwright
func (bc *BrowserContext) GetLocateElement(element *dom.DOMElementNode) playwright.Locator {
	currentPage := bc.GetCurrentPage()
	var currentFrame playwright.FrameLocator = nil

	// Start with the target element and collect all parents
	parents := []*dom.DOMElementNode{}
	current := element
	for current.Parent != nil {
		parent := current.Parent
		parents = append(parents, parent)
		current = parent
	}

	// Reverse the parents list to process from top to bottom
	slices.Reverse(parents)

	// Process all iframe parents in sequence
	iframes := []*dom.DOMElementNode{}
	for _, item := range parents {
		if item.TagName == "iframe" {
			iframes = append(iframes, item)
		}
	}
	includeDynamicAttributes := structx.GetDefaultValue(bc.Config, "include_dynamic_attributes", true)
	for _, parent := range iframes {
		cssSelector := bc.EnhancedCssSelectorForElement(parent, includeDynamicAttributes)
		if currentFrame != nil {
			currentFrame = currentFrame.FrameLocator(cssSelector)
		} else {
			currentFrame = currentPage.FrameLocator(cssSelector)
		}
	}
	cssSelector := bc.EnhancedCssSelectorForElement(element, includeDynamicAttributes)
	if currentFrame != nil {
		return currentFrame.Locator(cssSelector)
	} else {
		return currentPage.Locator(cssSelector)
	}
}

func (bc *BrowserContext) NavigateTo(url string) error {
	if !bc.isUrlAllowed(url) {
		return &BrowserError{Message: "Navigation to non-allowed URL: " + url}
	}

	page := bc.GetCurrentPage()
	page.Goto(url)
	page.WaitForLoadState()
	return nil
}

func (bc *BrowserContext) ClickElementNode(elementNode *dom.DOMElementNode) (*string, error) {
	// Optimized method to click an element using xpath.
	page := bc.GetCurrentPage()

	elementLocator := bc.GetLocateElement(elementNode)
	if elementLocator == nil {
		return nil, &BrowserError{Message: fmt.Sprintf("Failed to click element - Element: %s not found", elementNode.Xpath)}
	}

	// Performs the actual click, handling both download and navigation scenarios.
	performClick := func(clickFunc func() error) (*string, error) {
		saveDownloadPath, ok := bc.Config["save_downloads_path"].(string)
		if ok {
			downloadInfo, err := page.ExpectDownload(clickFunc, playwright.PageExpectDownloadOptions{Timeout: playwright.Float(3000)})
			if err != nil {
				if strings.HasPrefix(err.Error(), "timeout:") {
					log.Debug("No download triggered within timeout. Checking navigation...")
					page.WaitForLoadState()
					bc.checkAndHandleNavigation(page)
					return nil, nil
				}
				return nil, err
			} else {
				suggestedFilename := downloadInfo.SuggestedFilename()
				uniqueFilename := bc.getUniqueFilename(saveDownloadPath, suggestedFilename)
				downloadPath := filepath.Join(saveDownloadPath, uniqueFilename)
				err := downloadInfo.SaveAs(downloadPath)
				if err != nil {
					return nil, err
				}
				log.Debugf("⬇️  Download triggered. Saved file to: %s", downloadPath)
				return &downloadPath, nil
			}
		} else {
			newPage, err := bc.GetSession().Context.ExpectPage(func() error {
				return clickFunc()
			}, playwright.BrowserContextExpectPageOptions{Timeout: playwright.Float(1500)})
			if err != nil {
				if strings.HasPrefix(err.Error(), "timeout:") {
					page.WaitForLoadState()
					bc.checkAndHandleNavigation(page)
					return nil, nil
				}
				log.Errorf("Failed to click element: %s", err)
				return nil, err
			}
			newPage.WaitForLoadState()
			bc.checkAndHandleNavigation(newPage)
			return nil, nil
		}
	}

	return performClick(func() error {
		// Use First() to handle cases where the locator matches multiple elements
		return elementLocator.First().Click(playwright.LocatorClickOptions{Timeout: playwright.Float(1500)})
	})
}

func (bc *BrowserContext) InputTextElementNode(elementNode *dom.DOMElementNode, text string) error {
	/*
		Input text into an element with proper error handling and state management.
		Handles different types of input fields and ensures proper element state before input.
	*/
	locator := bc.GetLocateElement(elementNode)

	if locator == nil {
		return &BrowserError{Message: "Element: " + elementNode.Xpath + " not found"}
	}

	// Ensure element is ready for input
	selectorState := playwright.WaitForSelectorState("visible")
	locator.WaitFor(playwright.LocatorWaitForOptions{State: &selectorState, Timeout: playwright.Float(1000)})
	isHidden, err := locator.IsHidden()
	if err != nil {
		return &BrowserError{Message: "Failed to check if element is hidden: " + elementNode.Xpath}
	}
	if !isHidden {
		locator.ScrollIntoViewIfNeeded(playwright.LocatorScrollIntoViewIfNeededOptions{Timeout: playwright.Float(1000)})
	}

	// Get element properties to determine input method
	tagNameAny, _ := locator.Evaluate("el => el.tagName.toLowerCase()", nil)
	tagName := tagNameAny.(string)

	if tagName == "input" || tagName == "textarea" {
		locator.Evaluate("el => { el.textContent = ''; el.value = ''; }", nil)
		err := locator.Fill(text)

		if err != nil {
			return &BrowserError{Message: "Failed to fill element: " + elementNode.Xpath}
		}

		value, err := locator.InputValue()
		if err != nil {
			return &BrowserError{Message: "Failed to get input value: " + elementNode.Xpath}
		}
		if value != text {
			return &BrowserError{Message: "Input value does not match: " + elementNode.Xpath}
		}
	} else {
		log.Warnf("Element: %s is not editable.", elementNode.Xpath)
		locator.Fill(text)
	}

	return nil
}

func (bc *BrowserContext) initializeSession() (*BrowserSession, error) {
	log.Debugf("🌎  Initializing new browser context with id: %s", bc.ContextId)
	pwBrowser := bc.Browser.GetPlaywrightBrowser()

	context, err := bc.createContext(pwBrowser)
	if err != nil {
		return nil, err
	}
	bc.pageEventHandler = nil

	pages := context.Pages()
	bc.Session = &BrowserSession{
		Context:     context,
		CachedState: nil,
	}

	var activePage playwright.Page = nil
	if bc.Browser.Config["cdp_url"] != nil {
		// If we have a saved target ID, try to find and activate it
		if bc.State.TargetId != nil {
			targets := bc.getCdpTargets()
			for _, target := range targets {
				if target["targetId"] == *bc.State.TargetId {
					// Find matching page by URL
					for _, page := range pages {
						if page.URL() == target["url"] {
							activePage = page
							break
						}
					}
					break
				}
			}
		}
	}

	if activePage == nil {
		if len(pages) > 0 && !strings.HasPrefix(pages[0].URL(), "chrome://") && !strings.HasPrefix(pages[0].URL(), "chrome-extension://") {
			activePage = pages[0]
			log.Debugf("🔍  Using existing page: %s", activePage.URL())
		} else {
			activePage, err = context.NewPage()
			if err != nil {
				return nil, err
			}
			activePage.Goto("about:blank")
			log.Debugf("🆕  Created new page: %s", activePage.URL())
		}

		// Get target ID for the active page
		if bc.Browser.Config["cdp_url"] != nil {
			targets := bc.getCdpTargets()
			for _, target := range targets {
				if target["url"] == activePage.URL() {
					bc.State.TargetId = playwright.String(activePage.URL())
					break
				}
			}
		}
	}
	log.Debugf("🫨  Bringing tab to front: %s", activePage.URL())
	activePage.BringToFront()
	activePage.WaitForLoadState() // 'load'

	bc.ActiveTab = activePage

	return bc.Session, nil
}

func (bc *BrowserContext) onPage(page playwright.Page) {
	if bc.Browser.Config["cdp_url"] != nil {
		page.Reload()
	}
	page.WaitForLoadState()
	log.Debugf("📑  New page opened: %s", page.URL())

	if !strings.HasPrefix(page.URL(), "chrome-extension://") && !strings.HasPrefix(page.URL(), "chrome://") {
		bc.ActiveTab = page
	}

	if bc.Session != nil {
		bc.State.TargetId = nil
	}
}

func (bc *BrowserContext) getCdpTargets() []map[string]interface{} {
	if bc.Browser.Config["cdp_url"] == nil || bc.Session == nil {
		return []map[string]interface{}{}
	}
	pages := bc.Session.Context.Pages()
	if len(pages) == 0 {
		return []map[string]interface{}{}
	}

	cdpSession, err := pages[0].Context().NewCDPSession(pages[0])
	if err != nil {
		return []map[string]interface{}{}
	}
	result, err := cdpSession.Send("Target.getTargets", map[string]interface{}{})
	if err != nil {
		return []map[string]interface{}{}
	}
	err = cdpSession.Detach()
	if err != nil {
		return []map[string]interface{}{}
	}
	return result.(map[string]interface{})["targetInfos"].([]map[string]interface{})
}

func (bc *BrowserContext) addNewPageListener(context playwright.BrowserContext) {
	bc.pageEventHandler = bc.onPage
	context.OnPage(bc.pageEventHandler)
}

// Generate a unique filename by appending (1), (2), etc., if a file already exists.
func (bc *BrowserContext) getUniqueFilename(directory, filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	newFilename := filename
	counter := 1

	for {
		fullPath := filepath.Join(directory, newFilename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			break
		}
		newFilename = fmt.Sprintf("%s (%d)%s", base, counter, ext)
		counter++
	}
	return newFilename
}

// Check if a URL is allowed based on the whitelist configuration
func (bc *BrowserContext) isUrlAllowed(url string) bool {
	allowedDomainsText, ok := bc.Config["allowed_domains"].(string)
	if !ok || allowedDomainsText == "" {
		return true
	}

	allowedDomains := strings.Split(allowedDomainsText, ",")
	for i := range allowedDomains {
		allowedDomains[i] = strings.ToLower(strings.TrimSpace(allowedDomains[i]))
	}

	// Special case: Allow 'about:blank' explicitly
	if url == "about:blank" {
		return true
	}

	parsedUrl, err := neturl.Parse(url)
	if err != nil {
		log.Printf("⛔️  Error checking URL allowlist: %v", err)
		return false
	}

	domain := strings.ToLower(parsedUrl.Host)
	// Remove port number if present
	if colonIdx := strings.Index(domain, ":"); colonIdx != -1 {
		domain = domain[:colonIdx]
	}

	// Check if domain matches any allowed domain pattern
	for _, allowedDomain := range allowedDomains {
		if domain == allowedDomain || strings.HasSuffix(domain, "."+allowedDomain) {
			return true
		}
	}
	return false
}

// Check if current page URL is allowed and handle if not.
func (bc *BrowserContext) checkAndHandleNavigation(page playwright.Page) error {
	if !bc.isUrlAllowed(page.URL()) {
		log.Warnf("⛔️  Navigation to non-allowed URL detected: %s", page.URL())
		err := bc.GoBack()
		if err != nil {
			log.Errorf("⛔️  Failed to go back after detecting non-allowed URL: %s", err)
		}
		return errors.New("Navigation to non-allowed URL: " + page.URL())
	}
	return nil
}

// TODO(MID): implement waitForPageAndFramesLoad
func (bc *BrowserContext) waitForPageAndFramesLoad(timeoutOverwrite *float64) error {
	// maxTime := 0.25
	// if timeoutOverwrite != nil {
	// 	maxTime = *timeoutOverwrite
	// }
	// log.Debug("🪨  Waiting for page and frames to load for %f seconds", maxTime)
	bc.waitForStableNetwork()
	page := bc.GetCurrentPage()
	bc.checkAndHandleNavigation(page)
	return nil
}

func (bc *BrowserContext) LoadCookies(context playwright.BrowserContext) error {
	cookiesFile, ok := bc.Config["cookies_file"].(string)
	if !ok {
		return nil
	}
	stat, err := os.Stat(cookiesFile)
	if err != nil {
		return nil
	}
	if stat.IsDir() {
		return nil
	}

	f, err := os.Open(cookiesFile)
	if err != nil {
		return err
	}
	defer f.Close()
	cookies := make([]playwright.OptionalCookie, 0)
	if err := json.NewDecoder(f).Decode(&cookies); err != nil {
		return err
	}
	log.Infof("🍪  Loaded %d cookies from %s", len(cookies), cookiesFile)
	if context != nil {
		return context.AddCookies(cookies)
	}
	if bc.Session == nil || bc.Session.Context == nil {
		return errors.New("no browser context")
	}
	return bc.Session.Context.AddCookies(cookies)
}

// current cookies to file
func (bc *BrowserContext) SaveCookies() error {
	cookiesFile, ok := bc.Config["cookies_file"].(string)
	if bc.Session != nil && bc.Session.Context != nil && ok {
		cookies, err := bc.Session.Context.Cookies()
		if err != nil {
			log.Warnf("❌  Failed to save cookies: %s", err.Error())
			return err
		}
		log.Debugf("🍪  Saving %d cookies to %s", len(cookies), cookiesFile)
		// Check if the path is a directory and create it if necessary
		dirname := filepath.Dir(cookiesFile)
		if dirname != "" {
			os.MkdirAll(dirname, 0755)
		}

		f, err := os.Create(cookiesFile)
		if err != nil {
			log.Warnf("❌  Failed to save cookies: %s", err.Error())
			return err
		}
		defer f.Close()

		if err := json.NewEncoder(f).Encode(cookies); err != nil {
			log.Warnf("❌  Failed to save cookies: %s", err.Error())
			return err
		}
	}

	return nil
}

// TODO(MID): implement waitForStableNetwork
func (bc *BrowserContext) waitForStableNetwork() error {
	return nil
}

// Creates a new browser context with anti-detection measures and loads cookies if available.
func (bc *BrowserContext) createContext(browser playwright.Browser) (playwright.BrowserContext, error) {
	var context playwright.BrowserContext
	var err error
	if bc.Browser.Config["cdp_url"] != nil && len(browser.Contexts()) > 0 {
		context = browser.Contexts()[0]
	} else if bc.Browser.Config["browser_binary_path"] != nil && len(browser.Contexts()) > 0 {
		context = browser.Contexts()[0]
	} else {
		context, err = browser.NewContext(
			playwright.BrowserNewContextOptions{
				NoViewport:        playwright.Bool(true),
				UserAgent:         playwright.String(structx.GetDefaultValue(bc.Browser.Config, "user_agent", "")),
				JavaScriptEnabled: playwright.Bool(true),
				BypassCSP:         playwright.Bool(bc.Browser.Config["disable_security"].(bool)),
				IgnoreHttpsErrors: playwright.Bool(bc.Browser.Config["disable_security"].(bool)),
				// RecordVideo: &playwright.RecordVideo{
				// 	Dir: bc.Browser.Config["save_recording_path"].(string),
				// 	Size: &playwright.Size{
				// 		Width:  bc.Browser.Config["browser_window_size"].(map[string]interface{})["width"].(int),
				// 		Height: bc.Browser.Config["browser_window_size"].(map[string]interface{})["height"].(int),
				// 	},
				// },
				// RecordHarPath:   playwright.String(bc.Browser.Config["save_har_path"].(string)),
				Locale:          playwright.String(structx.GetDefaultValue(bc.Browser.Config, "locale", "")),
				HttpCredentials: structx.GetDefaultValue[*playwright.HttpCredentials](bc.Browser.Config, "http_credentials", nil),
				IsMobile:        playwright.Bool(structx.GetDefaultValue(bc.Browser.Config, "is_mobile", false)),
				HasTouch:        playwright.Bool(bc.Browser.Config["has_touch"].(bool)),
				// Geolocation: bc.Browser.Config["geolocation"].(*playwright.Geolocation),
				// Permissions:     bc.Browser.Config["permissions"].([]string),
				TimezoneId: playwright.String(structx.GetDefaultValue(bc.Browser.Config, "timezone_id", "")),
			},
		)
		if err != nil {
			return nil, err
		}
	}

	bc.LoadCookies(context)

	initScript := `// Webdriver property
            Object.defineProperty(navigator, 'webdriver', {
                get: () => undefined
            });

            // Languages
            Object.defineProperty(navigator, 'languages', {
                get: () => ['en-US']
            });

            // Plugins
            Object.defineProperty(navigator, 'plugins', {
                get: () => [1, 2, 3, 4, 5]
            });

            // Chrome runtime
            window.chrome = { runtime: {} };

            // Permissions
            const originalQuery = window.navigator.permissions.query;
            window.navigator.permissions.query = (parameters) => (
                parameters.name === 'notifications' ?
                    Promise.resolve({ state: Notification.permission }) :
                    originalQuery(parameters)
            );
            (function () {
                const originalAttachShadow = Element.prototype.attachShadow;
                Element.prototype.attachShadow = function attachShadow(options) {
                    return originalAttachShadow.call(this, { ...options, mode: "open" });
                };
            })();`
	context.AddInitScript(playwright.Script{Content: &initScript})
	return context, nil
}

func (bc *BrowserContext) getCurrentPage(session *BrowserSession) playwright.Page {
	pages := session.Context.Pages()
	if bc.Browser.Config["cdp_url"] != nil && bc.State.TargetId != nil {
		targets := bc.getCdpTargets()
		for _, target := range targets {
			if target["targetId"] == *bc.State.TargetId {
				for _, page := range pages {
					if page.URL() == target["url"] {
						return page
					}
				}
			}
		}
	}
	if bc.ActiveTab != nil && !bc.ActiveTab.IsClosed() && slices.Contains(session.Context.Pages(), bc.ActiveTab) {
		return bc.ActiveTab
	}

	// fall back to most recently opened non-extension page (extensions are almost always invisible background targets)
	nonExtensionPages := []playwright.Page{}
	for _, page := range pages {
		if !strings.HasPrefix(page.URL(), "chrome-extension://") && !strings.HasPrefix(page.URL(), "chrome://") {
			nonExtensionPages = append(nonExtensionPages, page)
		}
	}
	if len(nonExtensionPages) > 0 {
		return nonExtensionPages[len(nonExtensionPages)-1]
	}
	page, err := session.Context.NewPage()
	if err == nil {
		return page
	}
	session, err = bc.initializeSession()
	if err != nil {
		panic(err)
	}
	page, err = session.Context.NewPage()
	if err != nil {
		panic(err)
	}
	bc.ActiveTab = page
	return page
}

func (bc *BrowserContext) GetTabsInfo() []*TabInfo {
	// Get information about all tabs
	session := bc.GetSession()

	tabsInfo := []*TabInfo{}
	for pageId, page := range session.Context.Pages() {
		title, _ := page.Title()
		tabInfo := TabInfo{
			PageId:       pageId,
			Url:          page.URL(),
			Title:        title,
			ParentPageId: nil,
		}
		tabsInfo = append(tabsInfo, &tabInfo)
	}
	return tabsInfo
}

func (bc *BrowserContext) SwitchToTab(pageId int) error {
	// Switch to a specific tab by its PageId
	session := bc.GetSession()
	pages := session.Context.Pages()

	if pageId >= len(pages) {
		message := "No tab found with page_id: " + strconv.Itoa(pageId)
		return &BrowserError{Message: message}
	}

	for pageId < 0 {
		pageId += len(pages)
	}
	page := pages[pageId]

	if !bc.isUrlAllowed(page.URL()) {
		return NewURLNotAllowedError(page.URL())
	}

	// Update target ID if using CDP
	if bc.Browser.Config["cdp_url"] != nil {
		targets := bc.getCdpTargets()
		for _, target := range targets {
			if target["url"] == page.URL() {
				targetId, ok := target["targetId"].(string)
				if ok {
					bc.State.TargetId = &targetId
					break
				}
			}
		}
	}

	bc.ActiveTab = page
	page.BringToFront()
	page.WaitForLoadState()
	return nil
}

func (bc *BrowserContext) GoBack() error {
	page := bc.GetCurrentPage()
	_, err := page.GoBack(playwright.PageGoBackOptions{Timeout: playwright.Float(1000), WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		return err
	}
	log.Debug("⏮️  Went back to " + page.URL())
	return nil
}

func (bc *BrowserContext) CreateNewTab(url string) error {
	if len(url) > 0 && !bc.isUrlAllowed(url) {
		return &BrowserError{Message: "Cannot create new tab with non-allowed URL: " + url}
	}
	session := bc.GetSession()
	newPage, err := session.Context.NewPage()
	if err != nil {
		return err
	}

	bc.ActiveTab = newPage

	newPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{Timeout: playwright.Float(500)})

	if len(url) > 0 {
		_, err := newPage.Goto(url)
		bc.waitForPageAndFramesLoad(playwright.Float(1.0))
		if err != nil {
			return err
		}
	}

	// TODO(MID): check CDP
	// Get target ID for new page if using CDP
	// if bc.Browser.Config["cdp_url"] != nil {
	// 	targets := bc.getCdpTargets()
	// 	for _, target := range targets {
	// 		if target["url"] == newPage.URL() {
	// 			bc.State.TargetId = playwright.String(target["targetId"].(string))
	// 			break
	// 		}
	// 	}
	// }

	return nil
}

// Removes all highlight overlays and labels created by the highlightElement function.
// Handles cases where the page might be closed or inaccessible.
func (bc *BrowserContext) RemoveHighlights() {
	page := bc.GetCurrentPage()
	if page == nil {
		return
	}
	_, err := page.Evaluate(` try {
                    // Remove the highlight container and all its contents
                    const container = document.getElementById('playwright-highlight-container');
                    if (container) {
                        container.remove();
                    }

                    // Remove highlight attributes from elements
                    const highlightedElements = document.querySelectorAll('[browser-user-highlight-id^="playwright-highlight-"]');
                    highlightedElements.forEach(el => {
                        el.removeAttribute('browser-user-highlight-id');
                    });
                } catch (e) {
                    console.error('Failed to remove highlights:', e);
                }`)
	if err != nil {
		log.Debugf("⚠  Failed to remove highlights (this is usually ok): %v", err)
	}
}
