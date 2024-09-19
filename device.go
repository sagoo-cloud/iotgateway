package iotgateway

import (
	"github.com/sagoo-cloud/iotgateway/model"
	"github.com/sagoo-cloud/iotgateway/vars"
)

// GetDevice 获取设备
func GetDevice(deviceKey string) *model.Device {
	if deviceKey != "" {
		device, err := vars.GetDevice(deviceKey)
		if err != nil {
			return nil
		}
		return device
	}
	return nil
}

// SaveDevice 保存设置
func SaveDevice(deviceKey string, device *model.Device) {
	vars.UpdateDeviceMap(deviceKey, device)
}

// GetDeviceCount 获取设备统计
func GetDeviceCount() int {
	return vars.CountDevices()
}
