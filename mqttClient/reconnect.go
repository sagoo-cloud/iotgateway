package mqttClient

import (
	"context"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/os/glog"
)

// ReconnectManager MQTT重连管理器
type ReconnectManager struct {
	initialBackoff    time.Duration
	maxBackoff        time.Duration
	maxRetries        int
	currentRetry      int
	lastReconnectTime time.Time
	mu                sync.Mutex
	client            mqtt.Client
	ctx               context.Context
	cancel            context.CancelFunc
	isReconnecting    bool
}

// NewReconnectManager 创建新的重连管理器
func NewReconnectManager(client mqtt.Client) *ReconnectManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ReconnectManager{
		initialBackoff: time.Second,
		maxBackoff:     time.Minute * 5,
		maxRetries:     10,
		client:         client,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Reconnect 执行重连操作
func (m *ReconnectManager) Reconnect() error {
	m.mu.Lock()
	if m.isReconnecting {
		m.mu.Unlock()
		return fmt.Errorf("已经在重连中")
	}
	m.isReconnecting = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.isReconnecting = false
		m.mu.Unlock()
	}()

	// 计算当前应该等待的时间
	backoff := m.initialBackoff
	for i := 0; i < m.currentRetry; i++ {
		backoff *= 2
		if backoff > m.maxBackoff {
			backoff = m.maxBackoff
			break
		}
	}

	// 等待计算出的时间
	select {
	case <-m.ctx.Done():
		return fmt.Errorf("重连被取消")
	case <-time.After(backoff):
	}

	// 尝试重连
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		m.currentRetry++
		if m.currentRetry >= m.maxRetries {
			return fmt.Errorf("达到最大重试次数: %d", m.maxRetries)
		}
		return token.Error()
	}

	// 重连成功，重置计数器
	m.currentRetry = 0
	m.lastReconnectTime = time.Now()
	return nil
}

// StartReconnectLoop 启动重连循环
func (m *ReconnectManager) StartReconnectLoop() {
	m.mu.Lock()
	if m.isReconnecting {
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()

	go func() {
		for {
			select {
			case <-m.ctx.Done():
				return
			default:
				if !m.client.IsConnected() {
					if err := m.Reconnect(); err != nil {
						glog.Error(m.ctx, "【IotGateway】MQTT重连失败: %v", err)
					} else {
						glog.Info(m.ctx, "【IotGateway】MQTT重连成功")
						// 重连成功后等待一段时间再检查
						time.Sleep(time.Second * 5)
					}
				} else {
					// 已连接，等待较长时间再检查
					time.Sleep(time.Second * 30)
				}
			}
		}
	}()
}

// Stop 停止重连管理器
func (m *ReconnectManager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

// GetReconnectStatus 获取重连状态
func (m *ReconnectManager) GetReconnectStatus() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return map[string]interface{}{
		"currentRetry":      m.currentRetry,
		"maxRetries":        m.maxRetries,
		"lastReconnectTime": m.lastReconnectTime,
		"isConnected":       m.client.IsConnected(),
		"isReconnecting":    m.isReconnecting,
	}
}
