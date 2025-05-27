# SagooIOT 网关 SDK 快速参考

## 快速开始

### 1. 安装依赖

```bash
go get -u github.com/sagoo-cloud/iotgateway
```

### 2. 最小化示例

```go
package main

import (
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/gctx"
    "github.com/sagoo-cloud/iotgateway"
)

func main() {
    ctx := gctx.GetInitCtx()
    gateway, _ := iotgateway.NewGateway(ctx, nil)
    gateway.Start()
}
```

## 协议处理器模板

```go
type MyProtocol struct{}

func (p *MyProtocol) Init(device *model.Device, data []byte) error {
    if device != nil {
        device.DeviceKey = "device_001"
        // 订阅平台下发事件
        iotgateway.ServerGateway.SubscribeServiceEvent(device.DeviceKey)
        iotgateway.ServerGateway.SubscribeSetEvent(device.DeviceKey)
    }
    return nil
}

func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    // 解析设备数据
    // 触发数据上报
    properties := map[string]interface{}{
        "temperature": 25.6,
        "humidity": 60.5,
    }
    
    eventData := g.Map{
        "DeviceKey": device.DeviceKey,
        "PropertieDataList": properties,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
    
    return []byte("OK"), nil
}

func (p *MyProtocol) Encode(device *model.Device, data interface{}, param ...string) ([]byte, error) {
    return json.Marshal(data)
}
```

## 配置文件模板

```yaml
# config/config.yaml
server:
  name: "IoT网关"
  addr: ":8080"
  netType: "tcp"  # tcp/udp/mqtt
  duration: 60s
  productKey: "your_product_key"
  deviceKey: "your_device_key"
  packetConfig:
    type: 3  # 0:无处理 1:固定长度 2:头部+体 3:分隔符
    delimiter: "\r\n"

mqtt:
  address: "tcp://localhost:1883"
  username: "username"
  password: "password"
  clientId: "gateway_client"
  keepAliveDuration: 30s
```

## 常用API

### 数据上报

```go
// 属性数据上报
func pushProperties(deviceKey string, data map[string]interface{}) {
    eventData := g.Map{
        "DeviceKey": deviceKey,
        "PropertieDataList": data,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
}

// 事件数据上报
func pushEvents(deviceKey string, events map[string]interface{}) {
    eventData := g.Map{
        "DeviceKey": deviceKey,
        "EventDataList": events,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
}
```

### 服务调用处理

```go
// 注册服务处理器
func initServiceHandlers() {
    event.On("restart", event.ListenerFunc(handleRestart), event.Normal)
    event.On("getStatus", event.ListenerFunc(handleGetStatus), event.Normal)
}

func handleRestart(e event.Event) error {
    deviceKey := gconv.String(e.Data()["DeviceKey"])
    messageId := gconv.String(e.Data()["MessageID"])
    
    // 执行重启逻辑
    result := map[string]interface{}{
        "status": "success",
        "message": "设备重启成功",
    }
    
    // 回复平台
    replyData := g.Map{
        "DeviceKey": deviceKey,
        "MessageID": messageId,
        "ReplyData": result,
    }
    event.Async(consts.PushServiceResDataToMQTT, replyData)
    
    return nil
}
```

### 设备管理

```go
// 设备上线
func deviceOnline(deviceKey string, device *model.Device) {
    vars.UpdateDeviceMap(deviceKey, device)
    iotgateway.ServerGateway.SubscribeServiceEvent(deviceKey)
    iotgateway.ServerGateway.SubscribeSetEvent(deviceKey)
}

// 设备离线
func deviceOffline(deviceKey string) {
    vars.DeleteFromDeviceMap(deviceKey)
    vars.ClearDeviceMessages(deviceKey)
}

// 获取设备状态
func getDeviceStatus(deviceKey string) (*model.Device, error) {
    return vars.GetDeviceMap(deviceKey)
}
```

## 常用常量

```go
// 事件类型
consts.PushAttributeDataToMQTT  // 属性上报
consts.PushServiceResDataToMQTT // 服务调用响应
consts.PushSetResDataToMQTT     // 属性设置响应

// 网络类型
consts.NetTypeTcpServer   // TCP服务器
consts.NetTypeUDPServer   // UDP服务器
consts.NetTypeMqttServer  // MQTT客户端

// 粘包处理类型
network.NoHandling         // 不处理
network.FixedLength        // 固定长度
network.HeaderBodySeparate // 头部+体
network.Delimiter          // 分隔符
```

## 错误处理模板

```go
func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    defer func() {
        if r := recover(); r != nil {
            glog.Errorf(context.Background(), "协议解析异常: %v", r)
        }
    }()
    
    if len(data) == 0 {
        return nil, errors.New("数据为空")
    }
    
    if device == nil {
        return nil, errors.New("设备信息为空")
    }
    
    // 解析逻辑...
    
    return nil, nil
}
```

## 调试技巧

### 启用调试日志

```go
import "github.com/gogf/gf/v2/os/glog"

func init() {
    glog.SetLevel(glog.LEVEL_ALL)
    glog.SetStdoutPrint(true)
}
```

### 监控缓存状态

```go
func monitorCache() {
    stats := vars.GetCacheStats()
    glog.Infof(context.Background(), "缓存统计: %+v", stats)
}
```

### 健康检查

```go
func healthCheck() map[string]interface{} {
    return map[string]interface{}{
        "status": "running",
        "mqtt_connected": iotgateway.ServerGateway.MQTTClient.IsConnected(),
        "device_count": vars.CountDevices(),
        "cache_stats": vars.GetCacheStats(),
    }
}
```

## 性能优化

### 对象池使用

```go
var packetPool = sync.Pool{
    New: func() interface{} {
        return &Packet{}
    },
}

func parsePacket(data []byte) *Packet {
    packet := packetPool.Get().(*Packet)
    defer packetPool.Put(packet)
    
    packet.Reset()
    packet.Parse(data)
    return packet
}
```

### 批量数据处理

```go
func batchProcessData(devices []string, data []map[string]interface{}) {
    for i, deviceKey := range devices {
        go func(key string, d map[string]interface{}) {
            pushProperties(key, d)
        }(deviceKey, data[i])
    }
}
```

## 常见问题解决

### 1. 设备连接不上

```bash
# 检查端口
netstat -tlnp | grep 8080

# 测试连接
telnet gateway_ip 8080
```

### 2. MQTT连接失败

```go
// 检查MQTT连接状态
if !iotgateway.ServerGateway.MQTTClient.IsConnected() {
    glog.Error(context.Background(), "MQTT连接断开")
}
```

### 3. 内存泄漏

```go
// 定期清理过期缓存
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        stats := vars.GetCacheStats()
        if expiredCount := stats["expiredCount"].(int); expiredCount > 100 {
            glog.Warn(context.Background(), "发现大量过期消息")
        }
    }
}()
```

## 部署脚本

### Dockerfile

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o gateway .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gateway .
COPY --from=builder /app/config ./config
CMD ["./gateway"]
```

### docker-compose.yml

```yaml
version: '3.8'
services:
  gateway:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config:/root/config
      - ./logs:/root/logs
    environment:
      - ENV=production
    restart: unless-stopped
```

### 启动脚本

```bash
#!/bin/bash
# start.sh

# 设置环境变量
export GATEWAY_ENV=production

# 创建日志目录
mkdir -p logs

# 启动网关
./gateway > logs/gateway.log 2>&1 &

echo "网关已启动，PID: $!"
```

## 监控脚本

```bash
#!/bin/bash
# monitor.sh

while true; do
    # 检查进程是否存在
    if ! pgrep -f "gateway" > /dev/null; then
        echo "$(date): 网关进程不存在，正在重启..."
        ./start.sh
    fi
    
    # 检查内存使用
    mem_usage=$(ps -o pid,ppid,cmd,%mem,%cpu --sort=-%mem -C gateway | tail -n +2)
    echo "$(date): 内存使用情况: $mem_usage"
    
    sleep 60
done
``` 