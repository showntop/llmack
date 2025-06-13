package adb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/showntop/llmack/pkg/adb"
)

// Controller Android 设备交互工具
type Controller struct {
	Serial            string
	DeviceManager     *adb.Manager
	ClickableElements []UIElement
	LastScreenshot    string
	Reason            string
	Success           bool
	Finished          bool
	Memory            []string
	Screenshots       []ScreenshotInfo
}

func NewController(serial string) *Controller {

	ctrl := &Controller{
		Serial:            serial,
		DeviceManager:     adb.NewManager(""),
		ClickableElements: make([]UIElement, 0),
	}

	if registry != nil {
		err := RegisterTool(registry, "get_clickable_elements", "获取可点击的 UI 元素", ctrl.GetClickableElements)
		if err != nil {
			panic(err)
		}
		err = RegisterTool(registry, "tap_by_index", "通过元素索引点击元素", ctrl.TapByIndex)
		if err != nil {
			panic(err)
		}
		err = RegisterTool(registry, "tap_by_coordinates", "通过坐标点击元素", ctrl.TapByCoordinates)
		if err != nil {
			panic(err)
		}
		err = RegisterTool(registry, "swipe", "滑动", ctrl.Swipe)
		if err != nil {
			panic(err)
		}
		err = RegisterTool(registry, "input_text", "输入文本", ctrl.InputText)
		if err != nil {
			panic(err)
		}
		err = RegisterTool(registry, "press_key", "按键", ctrl.PressKey)
		if err != nil {
			panic(err)
		}
		err = RegisterTool(registry, "start_app", "启动应用", ctrl.StartApp)
		if err != nil {
			panic(err)
		}
		err = RegisterTool(registry, "install_app", "安装应用", ctrl.InstallApp)
		if err != nil {
			panic(err)
		}
		err = RegisterTool(registry, "take_screenshot", "截屏", ctrl.TakeScreenshot)
		if err != nil {
			panic(err)
		}
		err = RegisterTool(registry, "list_packages", "列出包", ctrl.ListPackages)
		if err != nil {
			panic(err)
		}
		// err = RegisterTool(registry, "complete", "完成任务", ctrl.Complete)
		// if err != nil {
		// 	panic(err)
		// }
		err = RegisterTool(registry, "get_phone_state", "获取手机状态", ctrl.GetPhoneState)
		if err != nil {
			panic(err)
		}
	}
	return ctrl
}

type ActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UIElement UI 元素结构
type UIElement struct {
	Index       int            `json:"index,omitempty"`
	Text        string         `json:"text,omitempty"`
	ContentDesc string         `json:"content_desc,omitempty"`
	Bounds      string         `json:"bounds,omitempty"`
	Clickable   bool           `json:"clickable,omitempty"`
	Children    []UIElement    `json:"children,omitempty"`
	Attributes  map[string]any `json:"attributes,omitempty"`
}

// ScreenshotInfo 截图信息
type ScreenshotInfo struct {
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

// PhoneState 手机状态
type PhoneState struct {
	BatteryLevel    int    `json:"battery_level"`
	WifiConnected   bool   `json:"wifi_connected"`
	NetworkType     string `json:"network_type"`
	Brightness      int    `json:"brightness"`
	VolumeLevel     int    `json:"volume_level"`
	IsScreenLocked  bool   `json:"is_screen_locked"`
	CurrentActivity string `json:"current_activity"`
}

// GetDevice 获取设备实例
func (t *Controller) GetDevice(ctx context.Context) (*adb.Device, error) {
	if t.Serial == "" {
		return nil, fmt.Errorf("未指定设备序列号")
	}

	device, err := t.DeviceManager.GetDevice(ctx, t.Serial)
	if err != nil {
		return nil, fmt.Errorf("设备 %s 未找到", t.Serial)
	}

	return device, nil
}

// GetClickables 获取可点击的 UI 元素
type GetClickableElementsParams struct {
	Serial string `json:"serial"`
}

func (t *Controller) GetClickableElements(ctx context.Context, params GetClickableElementsParams) ([]UIElement, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return nil, err
	}

	// 创建临时文件
	tempFile, err := ioutil.TempFile("", "ui_elements_*.json")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tempFile.Name())
	localPath := tempFile.Name()

	// 重试逻辑
	maxRetries := 30
	retryInterval := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		// 清理 logcat
		_, _ = device.Shell(ctx, "logcat -c")

		// 触发自定义服务获取交互元素
		_, err = device.Shell(ctx, "am broadcast -a com.droidrun.portal.GET_ELEMENTS")
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		// 轮询 JSON 文件路径
		devicePath, err := t.pollForJSONPath(ctx, device, 10*time.Second)
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		// 从设备拉取 JSON 文件
		err = device.Wrapper.PullFile(ctx, device.Serial, devicePath, localPath)
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		// 读取并解析 JSON 文件
		data, err := ioutil.ReadFile(localPath)
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		var elements []UIElement
		err = json.Unmarshal(data, &elements)
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		if len(elements) > 0 {
			// 过滤掉 type 属性
			filteredElements := t.filterUIElements(elements)
			t.ClickableElements = filteredElements

			// 小延迟确保 UI 完全加载
			time.Sleep(500 * time.Millisecond)

			return filteredElements, nil
		}

		time.Sleep(retryInterval)
	}

	return nil, fmt.Errorf("在 %d 秒重试后未能获取 UI 元素", maxRetries)
}

// pollForJSONPath 轮询 JSON 文件路径
func (t *Controller) pollForJSONPath(ctx context.Context, device *adb.Device, timeout time.Duration) (string, error) {
	startTime := time.Now()
	pollInterval := 200 * time.Millisecond

	for time.Since(startTime) < timeout {
		logcatOutput, err := device.Shell(ctx, "logcat -d | grep \"DROIDRUN_FILE\" | grep \"JSON data written to\" | tail -1")
		if err == nil {
			re := regexp.MustCompile(`JSON data written to: (.*)`)
			matches := re.FindStringSubmatch(logcatOutput)
			if len(matches) > 1 {
				return strings.TrimSpace(matches[1]), nil
			}
		}

		time.Sleep(pollInterval)
	}

	return "", fmt.Errorf("轮询超时")
}

// filterUIElements 过滤 UI 元素，移除 type 属性
func (t *Controller) filterUIElements(elements []UIElement) []UIElement {
	var filtered []UIElement
	for _, element := range elements {
		// 移除 type 属性
		if element.Attributes != nil {
			delete(element.Attributes, "type")
		}

		// 递归过滤子元素
		if len(element.Children) > 0 {
			element.Children = t.filterUIElements(element.Children)
		}

		filtered = append(filtered, element)
	}
	return filtered
}

// TapByIndex 通过索引点击元素
type TapByIndexParams struct {
	Index int `json:"index"`
}

func (t *Controller) TapByIndex(ctx context.Context, params TapByIndexParams) (*ActionResult, error) {
	if len(t.ClickableElements) == 0 {
		return &ActionResult{Success: false, Message: "没有可用的可点击元素"}, fmt.Errorf("没有可用的可点击元素")
	}

	element := t.findElementByIndex(t.ClickableElements, params.Index)
	if element == nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("索引 %d 处没有找到元素", params.Index)}, fmt.Errorf("索引 %d 处没有找到元素", params.Index)
	}

	// 解析边界并计算中心点
	x, y, err := t.parseBounds(element.Bounds)
	if err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("解析元素边界失败: %w", err)}, fmt.Errorf("解析元素边界失败: %w", err)
	}

	return t.TapByCoordinates(ctx, TapByCoordinatesParams{X: x, Y: y})
}

// findElementByIndex 通过索引查找元素
func (t *Controller) findElementByIndex(elements []UIElement, targetIndex int) *UIElement {
	for _, element := range elements {
		if element.Index == targetIndex {
			return &element
		}
		if len(element.Children) > 0 {
			if found := t.findElementByIndex(element.Children, targetIndex); found != nil {
				return found
			}
		}
	}
	return nil
}

// parseBounds 解析边界字符串并返回中心坐标
func (t *Controller) parseBounds(bounds string) (int, int, error) {
	// 解析格式: "[x1,y1][x2,y2]"
	re := regexp.MustCompile(`\[(\d+),(\d+)\]\[(\d+),(\d+)\]`)
	matches := re.FindStringSubmatch(bounds)
	if len(matches) != 5 {
		return 0, 0, fmt.Errorf("无效的边界格式: %s", bounds)
	}

	x1, _ := parseInt(matches[1])
	y1, _ := parseInt(matches[2])
	x2, _ := parseInt(matches[3])
	y2, _ := parseInt(matches[4])

	centerX := (x1 + x2) / 2
	centerY := (y1 + y2) / 2

	return centerX, centerY, nil
}

// TapByCoordinates 通过坐标点击
type TapByCoordinatesParams struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (t *Controller) TapByCoordinates(ctx context.Context, params TapByCoordinatesParams) (*ActionResult, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("获取设备失败: %w", err)}, fmt.Errorf("获取设备失败: %w", err)
	}

	if err := device.Tap(ctx, params.X, params.Y); err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("点击失败: %w", err)}, fmt.Errorf("点击失败: %w", err)
	}

	return &ActionResult{Success: true, Message: "点击成功"}, nil
}

// Swipe 滑动手势
type SwipeParams struct {
	StartX     int `json:"start_x"`
	StartY     int `json:"start_y"`
	EndX       int `json:"end_x"`
	EndY       int `json:"end_y"`
	DurationMs int `json:"duration_ms"`
}

func (t *Controller) Swipe(ctx context.Context, params SwipeParams) (*ActionResult, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("获取设备失败: %w", err)}, fmt.Errorf("获取设备失败: %w", err)
	}

	if err := device.Swipe(ctx, params.StartX, params.StartY, params.EndX, params.EndY, params.DurationMs); err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("滑动失败: %w", err)}, fmt.Errorf("滑动失败: %w", err)
	}

	return &ActionResult{Success: true, Message: "滑动成功"}, nil
}

// InputText 输入文本
type InputTextParams struct {
	Text string `json:"text"`
}

func (t *Controller) InputText(ctx context.Context, params InputTextParams) (*ActionResult, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("获取设备失败: %w", err)}, fmt.Errorf("获取设备失败: %w", err)
	}

	if err := device.InputText(ctx, params.Text); err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("输入失败: %w", err)}, fmt.Errorf("输入失败: %w", err)
	}

	return &ActionResult{Success: true, Message: "输入成功"}, nil
}

// PressKey 按键
type PressKeyParams struct {
	Keycode int `json:"keycode"`
}

func (t *Controller) PressKey(ctx context.Context, params PressKeyParams) (*ActionResult, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("获取设备失败: %w", err)}, fmt.Errorf("获取设备失败: %w", err)
	}

	if err := device.PressKey(ctx, params.Keycode); err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("按键失败: %w", err)}, fmt.Errorf("按键失败: %w", err)
	}

	return &ActionResult{Success: true, Message: "按键成功"}, nil
}

// StartApp 启动应用
type StartAppParams struct {
	Pkg      string `json:"pkg"`
	Activity string `json:"activity"`
}

func (t *Controller) StartApp(ctx context.Context, params StartAppParams) (*ActionResult, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("获取设备失败: %w", err)}, fmt.Errorf("获取设备失败: %w", err)
	}

	if err := device.StartApp(ctx, params.Pkg, params.Activity); err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("启动应用失败: %w", err)}, fmt.Errorf("启动应用失败: %w", err)
	}

	return &ActionResult{Success: true, Message: "启动应用成功"}, nil
}

// InstallApp 安装应用
type InstallAppParams struct {
	ApkPath          string `json:"apk_path"`
	Reinstall        bool   `json:"reinstall"`
	GrantPermissions bool   `json:"grant_permissions"`
}

func (t *Controller) InstallApp(ctx context.Context, params InstallAppParams) (*ActionResult, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("获取设备失败: %w", err)}, fmt.Errorf("获取设备失败: %w", err)
	}

	if err := device.InstallApp(ctx, params.ApkPath, params.Reinstall, params.GrantPermissions); err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("安装应用失败: %w", err)}, fmt.Errorf("安装应用失败: %w", err)
	}

	return &ActionResult{Success: true, Message: "安装应用成功"}, nil
}

// TakeScreenshot 截屏
type TakeScreenshotParams struct {
	Quality int `json:"quality"`
}

func (t *Controller) TakeScreenshot(ctx context.Context, params TakeScreenshotParams) (*ActionResult, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("获取设备失败: %w", err)}, fmt.Errorf("获取设备失败: %w", err)
	}

	localPath, _, err := device.TakeScreenshot(ctx, params.Quality)
	if err != nil {
		return &ActionResult{Success: false, Message: fmt.Sprintf("截屏失败: %w", err)}, fmt.Errorf("截屏失败: %w", err)
	}

	t.LastScreenshot = localPath
	t.Screenshots = append(t.Screenshots, ScreenshotInfo{
		Path:      localPath,
		Timestamp: time.Now(),
	})

	return &ActionResult{Success: true, Message: "截屏成功"}, nil
}

// ListPackages 列出包
type ListPackagesParams struct {
	IncludeSystemApps bool `json:"include_system_apps"`
}

func (t *Controller) ListPackages(ctx context.Context, params ListPackagesParams) ([]string, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return nil, err
	}

	packages, err := device.ListPackages(ctx, params.IncludeSystemApps)
	if err != nil {
		return nil, err
	}

	var packageNames []string
	for _, pkg := range packages {
		packageNames = append(packageNames, pkg.Package)
	}

	return packageNames, nil
}

// Complete 完成任务
type CompleteParams struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}

func (t *Controller) Complete(ctx context.Context, params CompleteParams) {
	t.Success = params.Success
	t.Reason = params.Reason
	t.Finished = true
}

// GetPhoneState 获取手机状态
type GetPhoneStateParams struct {
	Serial string `json:"serial"`
}

func (t *Controller) GetPhoneState(ctx context.Context, params GetPhoneStateParams) (*PhoneState, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return nil, err
	}

	state := &PhoneState{}

	// 获取电池电量
	if batteryOutput, err := device.Shell(ctx, "dumpsys battery | grep level"); err == nil {
		if matches := regexp.MustCompile(`level: (\d+)`).FindStringSubmatch(batteryOutput); len(matches) > 1 {
			state.BatteryLevel, _ = parseInt(matches[1])
		}
	}

	// 获取 WiFi 状态
	if wifiOutput, err := device.Shell(ctx, "dumpsys wifi | grep 'Wi-Fi is'"); err == nil {
		state.WifiConnected = strings.Contains(wifiOutput, "enabled")
	}

	// 获取亮度
	if brightnessOutput, err := device.Shell(ctx, "settings get system screen_brightness"); err == nil {
		state.Brightness, _ = parseInt(strings.TrimSpace(brightnessOutput))
	}

	return state, nil
}

// Remember 记住信息
func (t *Controller) Remember(information string) string {
	t.Memory = append(t.Memory, information)
	return fmt.Sprintf("已记住: %s", information)
}

// GetMemory 获取记忆
func (t *Controller) GetMemory() []string {
	return t.Memory
}

// 辅助函数
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
