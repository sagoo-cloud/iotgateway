package vars

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sagoo-cloud/iotgateway/model"
)

// 设备信息列表 - 保持原有变量名以确保向下兼容
var deviceMessageMap sync.Map

// 增强的消息缓存，包含过期时间
type enhancedUpMessage struct {
	Message    model.UpMessage `json:"message"`
	ExpireTime int64           `json:"expireTime"`
	CreateTime int64           `json:"createTime"`
}

// 复合键缓存，用于支持同一设备的多个并发消息
var compositeKeyMessageMap sync.Map

// 消息过期时间（默认30秒）
const messageExpireTimeout = 30 * time.Second

// 定期清理过期消息的定时器
var cleanupTicker *time.Ticker
var cleanupOnce sync.Once

// 启动定期清理任务
func startCleanupRoutine() {
	cleanupOnce.Do(func() {
		cleanupTicker = time.NewTicker(5 * time.Minute) // 每5分钟清理一次
		go func() {
			for {
				select {
				case <-cleanupTicker.C:
					cleanExpiredMessages()
				}
			}
		}()
	})
}

// 清理过期消息
func cleanExpiredMessages() {
	now := time.Now().Unix()
	// 清理复合键缓存中的过期消息
	compositeKeyMessageMap.Range(func(key, value interface{}) bool {
		if enhanced, ok := value.(enhancedUpMessage); ok {
			if enhanced.ExpireTime < now {
				compositeKeyMessageMap.Delete(key)
			}
		}
		return true
	})

	// 清理原有缓存中的过期消息
	deviceMessageMap.Range(func(key, value interface{}) bool {
		if enhanced, ok := value.(enhancedUpMessage); ok {
			if enhanced.ExpireTime < now {
				deviceMessageMap.Delete(key)
			}
		}
		return true
	})
}

// 生成复合键
func generateCompositeKey(deviceKey, messageId string) string {
	return fmt.Sprintf("%s_%s", deviceKey, messageId)
}

// UpdateUpMessageMap 保持原有函数签名，内部优化实现
func UpdateUpMessageMap(key string, device model.UpMessage) {
	// 启动清理任务
	startCleanupRoutine()

	now := time.Now()
	enhanced := enhancedUpMessage{
		Message:    device,
		ExpireTime: now.Add(messageExpireTimeout).Unix(),
		CreateTime: now.Unix(),
	}

	// 原有的按设备Key存储方式（保持向下兼容）
	deviceMessageMap.Store(key, enhanced)

	// 新增：按复合键存储，支持同一设备的多个并发消息
	if device.MessageID != "" {
		compositeKey := generateCompositeKey(key, device.MessageID)
		compositeKeyMessageMap.Store(compositeKey, enhanced)
	}
}

// GetUpMessageMap 保持原有函数签名，内部优化实现
func GetUpMessageMap(key string) (res model.UpMessage, err error) {
	// 先尝试从原有缓存获取
	v, ok := deviceMessageMap.Load(key)
	if !ok {
		err = errors.New("not data")
		return
	}

	// 检查是否为增强类型
	if enhanced, ok := v.(enhancedUpMessage); ok {
		// 检查是否过期
		if enhanced.ExpireTime < time.Now().Unix() {
			deviceMessageMap.Delete(key)
			err = errors.New("message expired")
			return
		}
		res = enhanced.Message
		return
	}

	// 兼容旧数据格式
	if oldMsg, ok := v.(model.UpMessage); ok {
		res = oldMsg
		return
	}

	err = errors.New("invalid data format")
	return
}

// 新增：根据设备Key和消息ID获取消息（用于精确匹配）
func GetUpMessageByCompositeKey(deviceKey, messageId string) (res model.UpMessage, err error) {
	compositeKey := generateCompositeKey(deviceKey, messageId)
	v, ok := compositeKeyMessageMap.Load(compositeKey)
	if !ok {
		// 如果复合键缓存中没有，尝试从原有缓存获取作为兼容
		return GetUpMessageMap(deviceKey)
	}

	if enhanced, ok := v.(enhancedUpMessage); ok {
		// 检查是否过期
		if enhanced.ExpireTime < time.Now().Unix() {
			compositeKeyMessageMap.Delete(compositeKey)
			err = errors.New("message expired")
			return
		}
		res = enhanced.Message
		return
	}

	err = errors.New("invalid data format")
	return
}

// DeleteFromUpMessageMap 保持原有函数签名，内部优化实现
func DeleteFromUpMessageMap(key string) {
	// 删除原有缓存
	if v, ok := deviceMessageMap.LoadAndDelete(key); ok {
		// 如果是增强类型，同时删除复合键缓存
		if enhanced, ok := v.(enhancedUpMessage); ok {
			if enhanced.Message.MessageID != "" {
				compositeKey := generateCompositeKey(key, enhanced.Message.MessageID)
				compositeKeyMessageMap.Delete(compositeKey)
			}
		}
	}
}

// 新增：根据复合键删除消息
func DeleteFromUpMessageMapByCompositeKey(deviceKey, messageId string) {
	compositeKey := generateCompositeKey(deviceKey, messageId)
	compositeKeyMessageMap.Delete(compositeKey)

	// 检查是否还有其他消息，如果没有则清理设备缓存
	hasOtherMessages := false
	compositeKeyMessageMap.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if len(keyStr) > len(deviceKey)+1 && keyStr[:len(deviceKey)+1] == deviceKey+"_" {
				hasOtherMessages = true
				return false // 停止遍历
			}
		}
		return true
	})

	// 如果没有其他消息，清理设备缓存
	if !hasOtherMessages {
		deviceMessageMap.Delete(deviceKey)
	}
}

// 新增：清理指定设备的所有消息（设备离线时调用）
func ClearDeviceMessages(deviceKey string) {
	// 清理设备缓存
	deviceMessageMap.Delete(deviceKey)

	// 清理所有相关的复合键缓存
	keysToDelete := make([]interface{}, 0)
	compositeKeyMessageMap.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if len(keyStr) > len(deviceKey)+1 && keyStr[:len(deviceKey)+1] == deviceKey+"_" {
				keysToDelete = append(keysToDelete, key)
			}
		}
		return true
	})

	// 删除收集到的键
	for _, key := range keysToDelete {
		compositeKeyMessageMap.Delete(key)
	}
}

// 新增：获取缓存统计信息（用于监控和调试）
func GetCacheStats() map[string]interface{} {
	deviceCount := 0
	compositeCount := 0
	expiredCount := 0
	now := time.Now().Unix()

	deviceMessageMap.Range(func(key, value interface{}) bool {
		deviceCount++
		if enhanced, ok := value.(enhancedUpMessage); ok {
			if enhanced.ExpireTime < now {
				expiredCount++
			}
		}
		return true
	})

	compositeKeyMessageMap.Range(func(key, value interface{}) bool {
		compositeCount++
		return true
	})

	return map[string]interface{}{
		"deviceCacheCount":    deviceCount,
		"compositeCacheCount": compositeCount,
		"expiredCount":        expiredCount,
		"lastCleanupTime":     time.Now().Format("2006-01-02 15:04:05"),
	}
}
