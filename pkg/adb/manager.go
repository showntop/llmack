package adb

import (
	"context"
	"fmt"
)

// Manager 设备管理器
type Manager struct {
	wrapper *Wrapper
	devices map[string]*Device
}

// NewManager 创建新的设备管理器
func NewManager(adbPath string) *Manager {
	return &Manager{
		wrapper: NewWrapper(adbPath),
		devices: make(map[string]*Device),
	}
}

// ListDevices 列出连接的设备
func (m *Manager) ListDevices(ctx context.Context) ([]*Device, error) {
	devicesInfo, err := m.wrapper.GetDevices(ctx)
	if err != nil {
		return nil, err
	}

	// 更新设备缓存
	currentSerials := make(map[string]bool)
	for _, deviceInfo := range devicesInfo {
		serial := deviceInfo.Serial
		currentSerials[serial] = true

		if _, exists := m.devices[serial]; !exists {
			m.devices[serial] = NewDevice(serial, m.wrapper)
		}
	}

	// 移除已断开连接的设备
	for serial := range m.devices {
		if !currentSerials[serial] {
			delete(m.devices, serial)
		}
	}

	// 返回设备列表
	var devices []*Device
	for _, device := range m.devices {
		devices = append(devices, device)
	}

	return devices, nil
}

// GetDevice 获取特定设备
func (m *Manager) GetDevice(ctx context.Context, serial string) (*Device, error) {
	if device, exists := m.devices[serial]; exists {
		return device, nil
	}

	// 尝试查找设备
	devices, err := m.ListDevices(ctx)
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		if device.Serial == serial {
			return device, nil
		}
	}

	return nil, fmt.Errorf("设备 %s 未找到", serial)
}

// Connect 连接到网络设备
func (m *Manager) Connect(ctx context.Context, host string, port int) (*Device, error) {
	serial, err := m.wrapper.Connect(ctx, host, port)
	if err != nil {
		return nil, err
	}

	return m.GetDevice(ctx, serial)
}

// Disconnect 断开设备连接
func (m *Manager) Disconnect(ctx context.Context, serial string) error {
	err := m.wrapper.Disconnect(ctx, serial)
	if err == nil {
		delete(m.devices, serial)
	}
	return err
}
