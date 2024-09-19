package vars

import "sync"

var ClientList sync.Map // 存储客户端列表

func SetClient(deviceKey, value string) {
	ClientList.Store(deviceKey, value)
}

func GetClient(deviceKey string) (string, bool) {
	value, ok := ClientList.Load(deviceKey)
	return value.(string), ok
}
