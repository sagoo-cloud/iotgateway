package network

import (
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"net"
	"sync"
)

// EventType 监听器类型
type EventType int

const (
	EventNewConnection    EventType = iota // 新连接事件
	EventDataReceived                      // 数据接收事件
	EventConnectionClosed                  // 连接关闭事件
)

// Event 事件结构体
type Event struct {
	EventType EventType // 事件类型
	Conn      net.Conn  // 连接
	Data      []byte    // 数据
	Protocol  mqttProtocol.Protocol
}

// EventListener 事件监听器接口
type EventListener interface {
	OnEvent(event *Event) // 事件处理方法
}

// EventListeners 事件监听器列表
type EventListeners []EventListener

// EventDispatcher 定义事件分发器
type EventDispatcher struct {
	listeners map[EventType]EventListeners // 事件监听器列表
	mu        sync.RWMutex                 // 互斥锁
}

// AddEventListener 添加事件监听器
func (d *EventDispatcher) AddEventListener(eventType EventType, listener EventListener) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.listeners[eventType]; !ok {
		d.listeners[eventType] = make(EventListeners, 0)
	}
	d.listeners[eventType] = append(d.listeners[eventType], listener)
}

// DispatchEvent 分发事件
func (d *EventDispatcher) DispatchEvent(event *Event) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if listeners, ok := d.listeners[event.EventType]; ok {
		for _, listener := range listeners {
			listener.OnEvent(event)
		}
	}
}
