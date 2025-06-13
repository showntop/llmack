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

	"git.code.oa.com/trpc-go/trpc-naming-polaris/registry"
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
	registry := registry.NewRegistry()

	ctrl := &Controller{
		Serial:            serial,
		DeviceManager:     adb.NewManager(""),
		ClickableElements: make([]UIElement, 0),
	}

	RegisterTool(registry, "get_clickables", "获取可点击的 UI 元素", ctrl.GetClickables)
	return ctrl
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
type GetClickablesElementsParams struct {
	Serial string `json:"serial"`
}

func (t *Controller) GetClickables(ctx context.Context, params GetClickablesElementsParams) ([]UIElement, error) {
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
func (t *Controller) TapByIndex(ctx context.Context, index int, serial string) error {
	if len(t.ClickableElements) == 0 {
		return fmt.Errorf("没有可用的可点击元素")
	}

	element := t.findElementByIndex(t.ClickableElements, index)
	if element == nil {
		return fmt.Errorf("索引 %d 处没有找到元素", index)
	}

	// 解析边界并计算中心点
	x, y, err := t.parseBounds(element.Bounds)
	if err != nil {
		return fmt.Errorf("解析元素边界失败: %w", err)
	}

	return t.TapByCoordinates(ctx, x, y)
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
func (t *Controller) TapByCoordinates(ctx context.Context, x, y int) error {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return err
	}

	return device.Tap(ctx, x, y)
}

// Tap 点击（通过索引）
func (t *Controller) Tap(ctx context.Context, index int) error {
	return t.TapByIndex(ctx, index, t.Serial)
}

// Swipe 滑动手势
func (t *Controller) Swipe(ctx context.Context, startX, startY, endX, endY, durationMs int) error {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return err
	}

	return device.Swipe(ctx, startX, startY, endX, endY, durationMs)
}

// InputText 输入文本
func (t *Controller) InputText(ctx context.Context, text string, serial string) error {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return err
	}

	return device.InputText(ctx, text)
}

// PressKey 按键
func (t *Controller) PressKey(ctx context.Context, keycode int) error {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return err
	}

	return device.PressKey(ctx, keycode)
}

// StartApp 启动应用
func (t *Controller) StartApp(ctx context.Context, pkg, activity string) error {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return err
	}

	return device.StartApp(ctx, pkg, activity)
}

// InstallApp 安装应用
func (t *Controller) InstallApp(ctx context.Context, apkPath string, reinstall, grantPermissions bool) error {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return err
	}

	return device.InstallApp(ctx, apkPath, reinstall, grantPermissions)
}

// TakeScreenshot 截屏
func (t *Controller) TakeScreenshot(ctx context.Context) (bool, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return false, err
	}

	localPath, _, err := device.TakeScreenshot(ctx, 75)
	if err != nil {
		return false, err
	}

	t.LastScreenshot = localPath
	t.Screenshots = append(t.Screenshots, ScreenshotInfo{
		Path:      localPath,
		Timestamp: time.Now(),
	})

	return true, nil
}

// ListPackages 列出包
func (t *Controller) ListPackages(ctx context.Context, includeSystemApps bool) ([]string, error) {
	device, err := t.GetDevice(ctx)
	if err != nil {
		return nil, err
	}

	packages, err := device.ListPackages(ctx, includeSystemApps)
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
func (t *Controller) Complete(success bool, reason string) {
	t.Success = success
	t.Reason = reason
	t.Finished = true
}

// GetPhoneState 获取手机状态
func (t *Controller) GetPhoneState(ctx context.Context, serial string) (*PhoneState, error) {
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
