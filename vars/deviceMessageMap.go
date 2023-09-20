package vars

import (
	"errors"
	"github.com/sagoo-cloud/iotgateway/model"
	"sync"
)

// 设备信息列表
var deviceMessageMap sync.Map

func UpdateUpMessageMap(key string, device model.UpMessage) {
	deviceMessageMap.Store(key, device)
}

func GetUpMessageMap(key string) (res model.UpMessage, err error) {
	v, ok := deviceMessageMap.Load(key)
	if !ok {
		err = errors.New("not data")
		return
	}
	res = v.(model.UpMessage)
	return
}
func DeleteFromUpMessageMap(key string) {
	deviceMessageMap.Delete(key)
}
