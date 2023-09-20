package vars

import (
	"errors"
	"github.com/sagoo-cloud/iotgateway/model"
	"sync"
)

// 设备列表
var deviceListAllMap sync.Map

func UpdateDeviceMap(key string, device *model.Device) {
	deviceListAllMap.Store(key, device)
}

func GetDevice(key string) (res *model.Device, err error) {
	v, ok := deviceListAllMap.Load(key)
	if !ok {
		err = errors.New("not data")
		return
	}
	res = v.(*model.Device)
	return
}

// CountDevices 统计设备数量
func CountDevices() int {
	count := 0
	deviceListAllMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
