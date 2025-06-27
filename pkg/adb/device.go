package adb

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Device 表示一个 Android 设备
type Device struct {
	Serial     string
	Wrapper    *Wrapper
	properties map[string]string
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	Serial string
	Status string
}

// NewDevice 创建一个新的设备实例
func NewDevice(serial string, wrapper *Wrapper) *Device {
	return &Device{
		Serial:     serial,
		Wrapper:    wrapper,
		properties: make(map[string]string),
	}
}

// Shell 在设备上执行 shell 命令
func (d *Device) Shell(ctx context.Context, command string) (string, error) {
	return d.Wrapper.Shell(ctx, d.Serial, command)
}

// GetProperties 获取设备所有属性
func (d *Device) GetProperties(ctx context.Context) (map[string]string, error) {
	if len(d.properties) == 0 {
		props, err := d.Wrapper.GetProperties(ctx, d.Serial)
		if err != nil {
			return nil, err
		}
		d.properties = props
	}
	return d.properties, nil
}

// GetProperty 获取设备特定属性
func (d *Device) GetProperty(ctx context.Context, name string) (string, error) {
	props, err := d.GetProperties(ctx)
	if err != nil {
		return "", err
	}
	return props[name], nil
}

// Model 获取设备型号
func (d *Device) Model(ctx context.Context) (string, error) {
	return d.GetProperty(ctx, "ro.product.model")
}

// Brand 获取设备品牌
func (d *Device) Brand(ctx context.Context) (string, error) {
	return d.GetProperty(ctx, "ro.product.brand")
}

// AndroidVersion 获取 Android 版本
func (d *Device) AndroidVersion(ctx context.Context) (string, error) {
	return d.GetProperty(ctx, "ro.build.version.release")
}

// SDKLevel 获取 SDK 级别
func (d *Device) SDKLevel(ctx context.Context) (string, error) {
	return d.GetProperty(ctx, "ro.build.version.sdk")
}

// Tap 在指定坐标点击
func (d *Device) Tap(ctx context.Context, x, y int) error {
	cmd := fmt.Sprintf("input tap %d %d", x, y)
	_, err := d.Wrapper.Shell(ctx, d.Serial, cmd)
	return err
}

// Swipe 执行滑动手势
func (d *Device) Swipe(ctx context.Context, startX, startY, endX, endY, durationMs int) error {
	cmd := fmt.Sprintf("input swipe %d %d %d %d %d", startX, startY, endX, endY, durationMs)
	_, err := d.Wrapper.Shell(ctx, d.Serial, cmd)
	return err
}

// InputText 输入文本
func (d *Device) InputText(ctx context.Context, text string) error {
	// 对文本进行转义处理
	// escapedText := strings.ReplaceAll(text, " ", "%s")
	// cmd := fmt.Sprintf("input text %s", escapedText)
	// _, err := d.Wrapper.Shell(ctx, d.Serial, cmd)
	// return err

	// Save the current keyboard
	originalKeyboard, err := d.Wrapper.Shell(ctx, d.Serial, "settings get secure default_input_method")
	if err != nil {
		return err
	}
	originalKeyboard = strings.TrimSpace(originalKeyboard)
	fmt.Println("current keyboard: ", originalKeyboard)

	// Enable the Droidrun keyboard
	d.Wrapper.Shell(ctx, d.Serial, "ime enable com.droidrun.portal/.DroidrunKeyboardIME")
	// Set the Droidrun keyboard as the default
	d.Wrapper.Shell(ctx, d.Serial, "ime set com.droidrun.portal/.DroidrunKeyboardIME")

	// Wait for keyboard to change
	time.Sleep(200 * time.Millisecond)

	// Encode the text to Base64
	encodedText := base64.StdEncoding.EncodeToString([]byte(text))
	cmd := fmt.Sprintf("am broadcast -a com.droidrun.portal.DROIDRUN_INPUT_B64 --es msg %s -p com.droidrun.portal", encodedText)
	_, err = d.Wrapper.Shell(ctx, d.Serial, cmd)
	if err != nil {
		return err
	}

	// Wait for text input to complete
	time.Sleep(500 * time.Millisecond)

	// Restore the original keyboard
	if !strings.Contains(originalKeyboard, "com.droidrun.portal") {
		cmd = fmt.Sprintf("ime set %s", originalKeyboard)
		_, err = d.Wrapper.Shell(ctx, d.Serial, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

// PressKey 按键
func (d *Device) PressKey(ctx context.Context, keycode int) error {
	cmd := fmt.Sprintf("input keyevent %d", keycode)
	_, err := d.Wrapper.Shell(ctx, d.Serial, cmd)
	return err
}

// StartActivity 启动应用活动
func (d *Device) StartActivity(ctx context.Context, pkg, activity string, extras map[string]string) error {
	cmd := fmt.Sprintf("am start -n %s/%s", pkg, activity)
	if extras != nil {
		for key, value := range extras {
			cmd += fmt.Sprintf(" -e %s %s", key, value)
		}
	}
	_, err := d.Wrapper.Shell(ctx, d.Serial, cmd)
	return err
}

// StartApp 启动应用
func (d *Device) StartApp(ctx context.Context, pkg, activity string) error {
	if activity != "" {
		if !strings.HasPrefix(activity, ".") && !strings.Contains(activity, ".") {
			activity = "." + activity
		}

		if !strings.HasPrefix(activity, ".") && strings.Contains(activity, ".") && !strings.HasPrefix(activity, pkg) {
			// 完全限定的活动名称
			parts := strings.Split(activity, "/")
			if len(parts) > 1 {
				return d.StartActivity(ctx, parts[0], parts[1], nil)
			}
			return d.StartActivity(ctx, activity, "", nil)
		}

		// 相对活动名称
		return d.StartActivity(ctx, pkg, activity, nil)
	}

	// 使用 monkey 启动主活动
	cmd := fmt.Sprintf("monkey -p %s -c android.intent.category.LAUNCHER 1", pkg)
	_, err := d.Wrapper.Shell(ctx, d.Serial, cmd)
	return err
}

// InstallApp 安装 APK 应用
func (d *Device) InstallApp(ctx context.Context, apkPath string, reinstall, grantPermissions bool) error {
	if _, err := os.Stat(apkPath); os.IsNotExist(err) {
		return fmt.Errorf("APK 文件不存在: %s", apkPath)
	}

	args := []string{"install"}
	if reinstall {
		args = append(args, "-r")
	}
	if grantPermissions {
		args = append(args, "-g")
	}
	args = append(args, apkPath)

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	stdout, stderr, err := d.Wrapper.RunDeviceCommand(ctx, d.Serial, args)
	if err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}

	if strings.Contains(strings.ToLower(stdout), "success") {
		return nil
	}

	return fmt.Errorf("安装失败: %s %s", stdout, stderr)
}

// UninstallApp 卸载应用
func (d *Device) UninstallApp(ctx context.Context, pkg string, keepData bool) error {
	args := []string{"uninstall"}
	if keepData {
		args = append(args, "-k")
	}
	args = append(args, pkg)

	stdout, stderr, err := d.Wrapper.RunDeviceCommand(ctx, d.Serial, args)
	if err != nil {
		return fmt.Errorf("卸载失败: %w", err)
	}

	if strings.Contains(strings.ToLower(stdout), "success") {
		return nil
	}

	return fmt.Errorf("卸载失败: %s %s", stdout, stderr)
}

// TakeScreenshot 截屏
func (d *Device) TakeScreenshot(ctx context.Context, quality int) (string, []byte, error) {
	// 在设备上创建截图
	remotePath := "/sdcard/screenshot.png"
	_, err := d.Wrapper.Shell(ctx, d.Serial, fmt.Sprintf("screencap -p %s", remotePath))
	if err != nil {
		return "", nil, fmt.Errorf("创建截图失败: %w", err)
	}

	// 创建临时文件
	tempDir := os.TempDir()
	timestamp := time.Now().Unix()
	localPath := filepath.Join(tempDir, fmt.Sprintf("screenshot_%d.png", timestamp))

	// 从设备拉取截图
	err = d.Wrapper.PullFile(ctx, d.Serial, remotePath, localPath)
	if err != nil {
		return "", nil, fmt.Errorf("拉取截图失败: %w", err)
	}

	// 读取文件内容
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", nil, fmt.Errorf("读取截图文件失败: %w", err)
	}

	// 清理设备上的临时文件
	d.Wrapper.Shell(ctx, d.Serial, fmt.Sprintf("rm %s", remotePath))

	return localPath, data, nil
}

// ListPackages 列出已安装的包
func (d *Device) ListPackages(ctx context.Context, includeSystemApps bool) ([]PackageInfo, error) {
	cmd := "pm list packages -f"
	if !includeSystemApps {
		cmd += " -3" // 只显示第三方应用
	}

	output, err := d.Wrapper.Shell(ctx, d.Serial, cmd)
	if err != nil {
		return nil, err
	}

	var packages []PackageInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "package:") {
			pathAndPkg := strings.TrimPrefix(line, "package:")
			if idx := strings.LastIndex(pathAndPkg, "="); idx != -1 {
				path := pathAndPkg[:idx]
				pkg := pathAndPkg[idx+1:]
				packages = append(packages, PackageInfo{
					Package: strings.TrimSpace(pkg),
					Path:    strings.TrimSpace(path),
				})
			}
		}
	}

	return packages, nil
}

// PackageInfo 包信息
type PackageInfo struct {
	Package string `json:"package"`
	Path    string `json:"path"`
}
