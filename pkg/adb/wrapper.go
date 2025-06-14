package adb

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Wrapper ADB 命令包装器
type Wrapper struct {
	adbPath string
}

// NewWrapper 创建新的 ADB wrapper
func NewWrapper(adbPath string) *Wrapper {
	if adbPath == "" {
		adbPath = "adb" // 使用系统 PATH 中的 adb
	}
	return &Wrapper{
		adbPath: adbPath,
	}
}

// GetDevices 获取连接的设备列表
func (w *Wrapper) GetDevices(ctx context.Context) ([]DeviceInfo, error) {
	cmd := exec.CommandContext(ctx, w.adbPath, "devices")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行 adb devices 失败: %w", err)
	}

	var devices []DeviceInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			devices = append(devices, DeviceInfo{
				Serial: parts[0],
				Status: parts[1],
			})
		}
	}

	return devices, nil
}

// Shell 在设备上执行 shell 命令
func (w *Wrapper) Shell(ctx context.Context, serial, command string) (string, error) {
	fmt.Println("execute shell command: ", w.adbPath, " -s ", serial, " shell ", command)
	cmd := exec.CommandContext(ctx, w.adbPath, "-s", serial, "shell", command)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("执行 shell 命令失败: %w", err)
	}
	return string(output), nil
}

// GetProperties 获取设备属性
func (w *Wrapper) GetProperties(ctx context.Context, serial string) (map[string]string, error) {
	output, err := w.Shell(ctx, serial, "getprop")
	if err != nil {
		return nil, err
	}

	properties := make(map[string]string)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析 [key]: [value] 格式
		if strings.HasPrefix(line, "[") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], "[]")
				value := strings.Trim(parts[1], "[]")
				properties[key] = value
			}
		}
	}

	return properties, nil
}

// PushFile 推送文件到设备
func (w *Wrapper) PushFile(ctx context.Context, serial, localPath, remotePath string) error {
	cmd := exec.CommandContext(ctx, w.adbPath, "-s", serial, "push", localPath, remotePath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("推送文件失败: %w", err)
	}
	return nil
}

// PullFile 从设备拉取文件
func (w *Wrapper) PullFile(ctx context.Context, serial, remotePath, localPath string) error {
	cmd := exec.CommandContext(ctx, w.adbPath, "-s", serial, "pull", remotePath, localPath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("拉取文件失败: %w", err)
	}
	return nil
}

// Connect 连接到网络设备
func (w *Wrapper) Connect(ctx context.Context, host string, port int) (string, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	cmd := exec.CommandContext(ctx, w.adbPath, "connect", address)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
	}

	if strings.Contains(string(output), "connected") {
		return address, nil
	}

	return "", fmt.Errorf("连接失败: %s", string(output))
}

// Disconnect 断开设备连接
func (w *Wrapper) Disconnect(ctx context.Context, serial string) error {
	cmd := exec.CommandContext(ctx, w.adbPath, "disconnect", serial)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("断开连接失败: %w", err)
	}
	return nil
}

// RunDeviceCommand 运行设备命令
func (w *Wrapper) RunDeviceCommand(ctx context.Context, serial string, args []string) (string, string, error) {
	fullArgs := append([]string{"-s", serial}, args...)
	cmd := exec.CommandContext(ctx, w.adbPath, fullArgs...)

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return "", outputStr, fmt.Errorf("执行设备命令失败: %w", err)
	}

	return outputStr, "", nil
}
