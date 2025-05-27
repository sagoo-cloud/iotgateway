# SagooIOT 网关 SDK 开发手册

## 目录

1. [概述](#概述)
2. [快速开始](#快速开始)
3. [架构设计](#架构设计)
4. [核心组件](#核心组件)
5. [协议开发](#协议开发)
6. [配置管理](#配置管理)
7. [事件系统](#事件系统)
8. [网络层开发](#网络层开发)
9. [MQTT集成](#mqtt集成)
10. [最佳实践](#最佳实践)
11. [故障排查](#故障排查)
12. [API参考](#api参考)

---

## 概述

SagooIOT 网关 SDK 是一个用于快速开发物联网网关的 Go 语言框架。它提供了完整的网关基础设施，包括网络通信、协议解析、MQTT集成、事件处理等核心功能。

### 主要特性

- 🚀 **多协议支持** - 支持TCP、UDP、MQTT等多种网络协议
- 🔧 **可扩展架构** - 插件化的协议处理器设计
- 📡 **MQTT集成** - 与SagooIOT平台无缝集成
- 🎯 **事件驱动** - 基于事件的异步处理机制
- 🛡️ **高可靠性** - 内置连接管理、错误恢复、内存优化
- 📊 **监控友好** - 完善的日志和统计信息

### 系统要求

- Go 1.23.0 或更高版本
- 支持的操作系统：Linux、Windows、macOS
- 内存：建议 512MB 以上
- 网络：支持TCP/UDP/MQTT协议

---

## 快速开始

### 1. 安装SDK

```bash
go get -u github.com/sagoo-cloud/iotgateway
```

### 2. 创建基础项目

创建一个新的Go项目并初始化：

```bash
mkdir my-gateway
cd my-gateway
go mod init my-gateway
```

### 3. 编写主程序

创建 `main.go` 文件：

```go
package main

import (
    "context"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/gctx"
    "github.com/gogf/gf/v2/os/glog"
    "github.com/sagoo-cloud/iotgateway"
    "github.com/sagoo-cloud/iotgateway/version"
)

// 编译时版本信息
var (
    BuildVersion = "1.0.0"
    BuildTime    = ""
    CommitID     = ""
)

func main() {
    // 初始化日志
    glog.SetDefaultLogger(g.Log())
    
    // 显示版本信息
    version.ShowLogo(BuildVersion, BuildTime, CommitID)
    
    ctx := gctx.GetInitCtx()

    // 创建协议处理器（可选，如果不需要自定义协议可以传nil）
    protocol := &MyProtocol{}

    // 创建网关实例
    gateway, err := iotgateway.NewGateway(ctx, protocol)
    if err != nil {
        panic(err)
    }

    // 初始化自定义事件处理
    initEvents()

    // 启动网关
    gateway.Start()
}

// 初始化事件处理
func initEvents() {
    // 在这里注册自定义事件处理器
}
```

### 4. 实现协议处理器

创建 `protocol.go` 文件：

```go
package main

import (
    "github.com/sagoo-cloud/iotgateway/model"
)

type MyProtocol struct{}

// Init 初始化协议处理器
func (p *MyProtocol) Init(device *model.Device, data []byte) error {
    // 设备初始化逻辑
    if device != nil {
        device.DeviceKey = "device_001" // 设置设备标识
    }
    return nil
}

// Decode 解码设备上行数据
func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    // 解析设备数据
    // 这里实现你的协议解析逻辑
    
    // 示例：解析后触发数据上报
    // pushDeviceData(device.DeviceKey, parsedData)
    
    return nil, nil // 如果不需要回复数据，返回nil
}

// Encode 编码下行数据
func (p *MyProtocol) Encode(device *model.Device, data interface{}, param ...string) ([]byte, error) {
    // 编码要发送给设备的数据
    // 这里实现你的数据编码逻辑
    
    return []byte("encoded data"), nil
}
```

### 5. 配置文件

创建 `config/config.yaml` 文件：

```yaml
server:
  name: "我的IoT网关"
  addr: ":8080"
  netType: "tcp"
  duration: 60s
  productKey: "your_product_key"
  deviceKey: "your_device_key"
  deviceName: "Gateway Device"
  description: "IoT Gateway"
  deviceType: "gateway"
  manufacturer: "YourCompany"
  packetConfig:
    type: 3  # Delimiter类型
    delimiter: "\r\n"

mqtt:
  address: "tcp://localhost:1883"
  username: "your_username"
  password: "your_password"
  clientId: "gateway_client"
  keepAliveDuration: 30s
  duration: 60s
```

### 6. 运行网关

```bash
go run .
```

---

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    SagooIOT 网关 SDK                        │
├─────────────────────────────────────────────────────────────┤
│  应用层 (Application Layer)                                 │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │   事件处理器     │  │   业务逻辑       │                  │
│  └─────────────────┘  └─────────────────┘                  │
├─────────────────────────────────────────────────────────────┤
│  网关核心层 (Gateway Core Layer)                            │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │   协议处理器     │  │   设备管理       │                  │
│  └─────────────────┘  └─────────────────┘                  │
├─────────────────────────────────────────────────────────────┤
│  通信层 (Communication Layer)                               │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │   网络服务器     │  │   MQTT客户端     │                  │
│  │  (TCP/UDP)      │  │                 │                  │
│  └─────────────────┘  └─────────────────┘                  │
├─────────────────────────────────────────────────────────────┤
│  基础设施层 (Infrastructure Layer)                          │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │   配置管理       │  │   日志系统       │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

### 数据流向

```
设备 ──TCP/UDP──> 网络服务器 ──> 协议处理器 ──> 事件系统 ──> MQTT ──> SagooIOT平台
 ↑                                                                      │
 └──────────────── 协议处理器 <── 事件系统 <── MQTT <──────────────────────┘
```

---

## 核心组件

### 1. Gateway 网关核心

`Gateway` 是整个SDK的核心组件，负责协调各个子系统的工作。

```go
type Gateway struct {
    Address    string                    // 网关地址
    Version    string                    // 版本信息
    Status     string                    // 运行状态
    MQTTClient mqtt.Client              // MQTT客户端
    Server     network.NetworkServer    // 网络服务器
    Protocol   network.ProtocolHandler // 协议处理器
}
```

**主要方法：**

- `NewGateway(ctx, protocol)` - 创建网关实例
- `Start()` - 启动网关服务
- `SubscribeServiceEvent(deviceKey)` - 订阅服务下发事件
- `SubscribeSetEvent(deviceKey)` - 订阅属性设置事件

### 2. 设备模型

```go
type Device struct {
    DeviceKey    string                 // 设备唯一标识
    ClientID     string                 // 客户端ID
    OnlineStatus bool                   // 在线状态
    Conn         net.Conn               // 网络连接
    Metadata     map[string]interface{} // 元数据
    Info         map[string]interface{} // 设备信息
    AlarmInfo    map[string]interface{} // 报警信息
    LastActive   time.Time              // 最后活跃时间
}
```

### 3. 消息模型

```go
// 上行消息
type UpMessage struct {
    MessageID   string `json:"messageId"`   // 消息ID
    SendTime    int64  `json:"sendTime"`    // 发送时间
    RequestCode string `json:"requestCode"` // 请求代码
    MethodName  string `json:"methodName"`  // 方法名称
    Topic       string `json:"topic"`       // MQTT主题
}
```

---

## 协议开发

### 协议接口定义

所有自定义协议都需要实现 `ProtocolHandler` 接口：

```go
type ProtocolHandler interface {
    Init(device *model.Device, data []byte) error
    Encode(device *model.Device, data interface{}, param ...string) ([]byte, error)
    Decode(device *model.Device, data []byte) ([]byte, error)
}
```

### 协议开发步骤

#### 1. 实现Init方法

```go
func (p *MyProtocol) Init(device *model.Device, data []byte) error {
    // 1. 解析设备标识信息
    deviceKey := parseDeviceKey(data)
    if device != nil {
        device.DeviceKey = deviceKey
        device.OnlineStatus = true
    }
    
    // 2. 订阅平台下发事件
    if deviceKey != "" {
        iotgateway.ServerGateway.SubscribeServiceEvent(deviceKey)
        iotgateway.ServerGateway.SubscribeSetEvent(deviceKey)
    }
    
    return nil
}
```

#### 2. 实现Decode方法

```go
func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    // 1. 解析数据包
    packet, err := parsePacket(data)
    if err != nil {
        return nil, err
    }
    
    // 2. 根据数据类型处理
    switch packet.Type {
    case "heartbeat":
        return p.handleHeartbeat(device, packet)
    case "data":
        return p.handleData(device, packet)
    case "alarm":
        return p.handleAlarm(device, packet)
    }
    
    return nil, nil
}

func (p *MyProtocol) handleData(device *model.Device, packet *Packet) ([]byte, error) {
    // 准备属性数据
    properties := map[string]interface{}{
        "temperature": packet.Temperature,
        "humidity":    packet.Humidity,
        "timestamp":   time.Now().Unix(),
    }
    
    // 触发属性上报事件
    eventData := g.Map{
        "DeviceKey":         device.DeviceKey,
        "PropertieDataList": properties,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
    
    // 返回确认消息
    return []byte("ACK"), nil
}
```

#### 3. 实现Encode方法

```go
func (p *MyProtocol) Encode(device *model.Device, data interface{}, param ...string) ([]byte, error) {
    // 根据数据类型编码
    switch v := data.(type) {
    case map[string]interface{}:
        return p.encodeCommand(v)
    case string:
        return []byte(v), nil
    default:
        return json.Marshal(data)
    }
}

func (p *MyProtocol) encodeCommand(cmd map[string]interface{}) ([]byte, error) {
    // 构造命令包
    packet := CommandPacket{
        Header:  0xAA,
        Command: cmd["command"].(string),
        Data:    cmd["data"],
        CRC:     0, // 计算CRC
    }
    
    return packet.Serialize(), nil
}
```

### 协议开发最佳实践

1. **错误处理**：始终检查数据包的完整性和有效性
2. **设备标识**：确保能正确解析设备唯一标识
3. **状态管理**：及时更新设备在线状态
4. **数据验证**：验证数据格式和范围
5. **性能优化**：避免在协议处理中进行耗时操作

---

## 配置管理

### 配置结构

```go
type GatewayConfig struct {
    GatewayServerConfig GatewayServerConfig `json:"server"`
    MqttConfig          MqttConfig          `json:"mqtt"`
}
```

### 服务器配置

```go
type GatewayServerConfig struct {
    Name         string        `json:"name"`         // 网关服务名称
    Addr         string        `json:"addr"`         // 监听地址
    NetType      string        `json:"netType"`      // 网络类型: tcp/udp/mqtt
    SerUpTopic   string        `json:"serUpTopic"`   // 上行Topic
    SerDownTopic string        `json:"serDownTopic"` // 下行Topic
    Duration     time.Duration `json:"duration"`     // 心跳间隔
    ProductKey   string        `json:"productKey"`   // 产品标识
    DeviceKey    string        `json:"deviceKey"`    // 设备标识
    PacketConfig PacketConfig  `json:"packetConfig"` // 粘包处理配置
}
```

### 粘包处理配置

```go
type PacketConfig struct {
    Type         PacketHandlingType // 处理类型
    FixedLength  int               // 固定长度
    HeaderLength int               // 头部长度
    Delimiter    string            // 分隔符
}

// 处理类型常量
const (
    NoHandling         = 0 // 不处理
    FixedLength        = 1 // 固定长度
    HeaderBodySeparate = 2 // 头部+体
    Delimiter          = 3 // 分隔符
)
```

### MQTT配置

```go
type MqttConfig struct {
    Address               string        `json:"address"`               // MQTT服务器地址
    Username              string        `json:"username"`              // 用户名
    Password              string        `json:"password"`              // 密码
    ClientId              string        `json:"clientId"`              // 客户端ID
    ClientCertificateKey  string        `json:"clientCertificateKey"`  // 客户端证书密钥
    ClientCertificateCert string        `json:"clientCertificateCert"` // 客户端证书
    KeepAliveDuration     time.Duration `json:"keepAliveDuration"`     // 保持连接时长
    Duration              time.Duration `json:"duration"`              // 心跳间隔
}
```

### 配置文件示例

```yaml
# config/config.yaml
server:
  name: "智能网关"
  addr: ":8080"
  netType: "tcp"
  duration: 60s
  productKey: "smart_gateway"
  deviceKey: "gateway_001"
  deviceName: "智能网关设备"
  description: "用于工业设备接入的智能网关"
  deviceType: "gateway"
  manufacturer: "SagooIOT"
  packetConfig:
    type: 3  # 分隔符类型
    delimiter: "\r\n"

mqtt:
  address: "tcp://mqtt.sagoo.cn:1883"
  username: "gateway_user"
  password: "gateway_pass"
  clientId: "gateway_001"
  keepAliveDuration: 30s
  duration: 60s
```

---

## 事件系统

### 事件类型

SDK提供了以下预定义事件：

```go
const (
    PushAttributeDataToMQTT  = "PushAttributeDataToMQTT"  // 属性上报
    PushServiceResDataToMQTT = "PushServiceResDataToMQTT" // 服务调用结果上报
    PushSetResDataToMQTT     = "PushSetResDataToMQTT"     // 属性设置结果上报
)
```

### 属性数据上报

```go
func pushAttributeData(deviceKey string, properties map[string]interface{}) {
    eventData := g.Map{
        "DeviceKey":         deviceKey,
        "PropertieDataList": properties,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
}

// 使用示例
properties := map[string]interface{}{
    "temperature": 25.6,
    "humidity":    60.5,
    "pressure":    1013.25,
}
pushAttributeData("device_001", properties)
```

### 事件数据上报

```go
func pushEventData(deviceKey string, events map[string]interface{}) {
    eventData := g.Map{
        "DeviceKey":     deviceKey,
        "EventDataList": events,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
}

// 使用示例
events := map[string]interface{}{
    "alarm": map[string]interface{}{
        "level":   "high",
        "message": "温度过高",
        "time":    time.Now().Unix(),
    },
}
pushEventData("device_001", events)
```

### 服务调用响应

```go
func handleServiceCall(e event.Event) error {
    deviceKey := gconv.String(e.Data()["DeviceKey"])
    messageId := gconv.String(e.Data()["MessageID"])
    params := e.Data()
    
    // 处理服务调用逻辑
    result := processService(deviceKey, params)
    
    // 发送响应
    replyData := g.Map{
        "DeviceKey": deviceKey,
        "MessageID": messageId, // 重要：传递消息ID确保精确匹配
        "ReplyData": result,
    }
    event.Async(consts.PushServiceResDataToMQTT, replyData)
    
    return nil
}

// 注册服务调用处理器
func initEvents() {
    event.On("restart", event.ListenerFunc(handleRestartService), event.Normal)
    event.On("getStatus", event.ListenerFunc(handleGetStatus), event.Normal)
}
```

### 自定义事件处理

```go
// 定义自定义事件
const CustomDataEvent = "CustomDataEvent"

// 注册事件处理器
func initCustomEvents() {
    event.On(CustomDataEvent, event.ListenerFunc(handleCustomData), event.Normal)
}

func handleCustomData(e event.Event) error {
    data := e.Data()
    // 处理自定义数据
    log.Printf("收到自定义数据: %+v", data)
    return nil
}

// 触发自定义事件
func triggerCustomEvent(data map[string]interface{}) {
    event.MustFire(CustomDataEvent, data)
}
```

---

## 网络层开发

### 支持的网络类型

1. **TCP服务器** - 适用于长连接设备
2. **UDP服务器** - 适用于短连接或广播设备  
3. **MQTT客户端** - 适用于MQTT协议设备

### TCP服务器配置

```yaml
server:
  netType: "tcp"
  addr: ":8080"
  packetConfig:
    type: 3  # 分隔符处理
    delimiter: "\r\n"
```

### UDP服务器配置

```yaml
server:
  netType: "udp"
  addr: ":8080"
```

### MQTT客户端配置

```yaml
server:
  netType: "mqtt"
  serUpTopic: "device/+/data"      # 设备上行数据Topic
  serDownTopic: "device/+/command" # 设备下行命令Topic
```

### 粘包处理

SDK提供了多种粘包处理方式：

#### 1. 固定长度

```yaml
packetConfig:
  type: 1  # FixedLength
  fixedLength: 64
```

#### 2. 头部+体分离

```yaml
packetConfig:
  type: 2  # HeaderBodySeparate
  headerLength: 4  # 头部4字节表示体长度
```

#### 3. 分隔符

```yaml
packetConfig:
  type: 3  # Delimiter
  delimiter: "\r\n"
```

### 自定义网络处理

```go
// 实现自定义网络选项
func WithCustomTimeout(timeout time.Duration) network.Option {
    return func(s *network.BaseServer) {
        s.SetTimeout(timeout)
    }
}

// 创建自定义网络服务器
server := network.NewTCPServer(
    network.WithTimeout(2*time.Minute),
    network.WithProtocolHandler(protocol),
    network.WithCleanupInterval(5*time.Minute),
    WithCustomTimeout(30*time.Second),
)
```

---

## MQTT集成

### 连接管理

SDK自动管理MQTT连接，包括：

- 自动重连
- 心跳保持
- 连接状态监控
- 证书认证支持

### Topic规范

#### 上行数据Topic

```
/sys/{productKey}/{deviceKey}/thing/event/property/pack/post
```

#### 服务调用Topic

```
/sys/{productKey}/{deviceKey}/thing/service/{serviceId}
```

#### 属性设置Topic

```
/sys/{productKey}/{deviceKey}/thing/service/property/set
```

### 数据格式

#### 属性上报格式

```json
{
  "id": "message_id",
  "version": "1.0",
  "method": "thing.event.property.pack.post",
  "params": {
    "properties": {
      "temperature": {
        "value": 25.6,
        "time": 1640995200000
      }
    },
    "subDevices": [
      {
        "identity": {
          "productKey": "product_001",
          "deviceKey": "device_001"
        },
        "properties": {
          "humidity": {
            "value": 60.5,
            "time": 1640995200000
          }
        }
      }
    ]
  }
}
```

#### 服务调用格式

```json
{
  "id": "service_call_id",
  "version": "1.0",
  "method": "thing.service.restart",
  "params": {
    "delay": 5
  }
}
```

### MQTT客户端使用

```go
// 发布数据到指定Topic
func publishToMQTT(topic string, data interface{}) error {
    client := iotgateway.ServerGateway.MQTTClient
    if client == nil || !client.IsConnected() {
        return errors.New("MQTT客户端未连接")
    }
    
    jsonData, err := json.Marshal(data)
    if err != nil {
        return err
    }
    
    token := client.Publish(topic, 1, false, jsonData)
    return token.Error()
}
```

---

## 最佳实践

### 1. 错误处理

```go
func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    defer func() {
        if r := recover(); r != nil {
            log.Errorf("协议解析异常: %v", r)
        }
    }()
    
    // 数据长度检查
    if len(data) < 4 {
        return nil, errors.New("数据包长度不足")
    }
    
    // 数据格式验证
    if !isValidPacket(data) {
        return nil, errors.New("无效的数据包格式")
    }
    
    // 解析数据
    packet, err := parsePacket(data)
    if err != nil {
        return nil, fmt.Errorf("解析数据包失败: %v", err)
    }
    
    return p.processPacket(device, packet)
}
```

### 2. 性能优化

```go
// 使用对象池减少内存分配
var packetPool = sync.Pool{
    New: func() interface{} {
        return &Packet{}
    },
}

func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    // 从池中获取对象
    packet := packetPool.Get().(*Packet)
    defer packetPool.Put(packet)
    
    // 重置对象状态
    packet.Reset()
    
    // 解析数据
    err := packet.Parse(data)
    if err != nil {
        return nil, err
    }
    
    return p.processPacket(device, packet)
}
```

### 3. 并发安全

```go
type SafeProtocol struct {
    mu       sync.RWMutex
    devices  map[string]*model.Device
    counters map[string]int64
}

func (p *SafeProtocol) updateDeviceCounter(deviceKey string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    p.counters[deviceKey]++
}

func (p *SafeProtocol) getDeviceCounter(deviceKey string) int64 {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    return p.counters[deviceKey]
}
```

### 4. 资源管理

```go
func (p *MyProtocol) Init(device *model.Device, data []byte) error {
    // 设置设备清理回调
    if device != nil {
        device.Metadata = map[string]interface{}{
            "cleanup": func() {
                // 清理设备相关资源
                p.cleanupDevice(device.DeviceKey)
            },
        }
    }
    
    return nil
}

func (p *MyProtocol) cleanupDevice(deviceKey string) {
    // 清理缓存
    vars.ClearDeviceMessages(deviceKey)
    
    // 清理其他资源
    p.removeDeviceFromCache(deviceKey)
}
```

### 5. 日志记录

```go
import (
    "github.com/gogf/gf/v2/os/glog"
    "context"
)

func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    ctx := context.Background()
    
    glog.Debugf(ctx, "收到设备数据: deviceKey=%s, dataLen=%d", 
        device.DeviceKey, len(data))
    
    packet, err := parsePacket(data)
    if err != nil {
        glog.Errorf(ctx, "解析数据包失败: deviceKey=%s, error=%v", 
            device.DeviceKey, err)
        return nil, err
    }
    
    glog.Infof(ctx, "成功解析数据包: deviceKey=%s, type=%s", 
        device.DeviceKey, packet.Type)
    
    return p.processPacket(device, packet)
}
```

---

## 故障排查

### 常见问题

#### 1. 设备连接失败

**现象：** 设备无法连接到网关

**排查步骤：**
1. 检查网关监听地址和端口
2. 检查防火墙设置
3. 检查网络连通性
4. 查看网关日志

```bash
# 检查端口监听
netstat -tlnp | grep 8080

# 测试连接
telnet gateway_ip 8080
```

#### 2. 数据解析错误

**现象：** 收到数据但解析失败

**排查步骤：**
1. 检查数据格式是否正确
2. 检查协议实现是否有误
3. 添加调试日志查看原始数据

```go
func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    // 添加调试日志
    glog.Debugf(context.Background(), "原始数据: %x", data)
    glog.Debugf(context.Background(), "数据字符串: %s", string(data))
    
    // 解析逻辑...
}
```

#### 3. MQTT连接问题

**现象：** 无法连接到MQTT服务器

**排查步骤：**
1. 检查MQTT服务器地址和端口
2. 检查用户名密码
3. 检查证书配置
4. 查看MQTT连接日志

```go
// 添加MQTT连接状态监控
func monitorMQTTConnection() {
    client := iotgateway.ServerGateway.MQTTClient
    if client != nil {
        isConnected := client.IsConnected()
        glog.Infof(context.Background(), "MQTT连接状态: %v", isConnected)
    }
}
```

#### 4. 内存泄漏

**现象：** 网关运行时间长后内存持续增长

**排查步骤：**
1. 检查缓存清理是否正常
2. 检查设备离线清理
3. 使用内存分析工具

```go
// 监控缓存状态
func monitorCacheStats() {
    stats := vars.GetCacheStats()
    glog.Infof(context.Background(), "缓存统计: %+v", stats)
    
    if expiredCount := stats["expiredCount"].(int); expiredCount > 100 {
        glog.Warnf(context.Background(), "发现大量过期消息: %d", expiredCount)
    }
}
```

### 调试工具

#### 1. 启用调试日志

```yaml
# config/config.yaml
logger:
  level: "debug"
  stdout: true
```

#### 2. 性能监控

```go
import _ "net/http/pprof"
import "net/http"

func init() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}
```

访问 `http://localhost:6060/debug/pprof/` 查看性能数据。

#### 3. 健康检查

```go
func healthCheck() map[string]interface{} {
    return map[string]interface{}{
        "gateway_status":    "running",
        "mqtt_connected":    iotgateway.ServerGateway.MQTTClient.IsConnected(),
        "device_count":      vars.CountDevices(),
        "cache_stats":       vars.GetCacheStats(),
        "uptime":           time.Since(startTime).String(),
    }
}
```

---

## API参考

### 核心API

#### Gateway

```go
// 创建网关实例
func NewGateway(ctx context.Context, protocol network.ProtocolHandler) (*Gateway, error)

// 启动网关
func (gw *Gateway) Start()

// 订阅服务下发事件
func (gw *Gateway) SubscribeServiceEvent(deviceKey string)

// 订阅属性设置事件
func (gw *Gateway) SubscribeSetEvent(deviceKey string)

// 向设备下发数据（MQTT模式）
func (gw *Gateway) DeviceDownData(data interface{})
```

#### 设备管理

```go
// 更新设备信息
func UpdateDeviceMap(key string, device *model.Device)

// 获取设备信息
func GetDeviceMap(key string) (*model.Device, error)

// 删除设备信息
func DeleteFromDeviceMap(key string)

// 获取设备数量
func CountDevices() int
```

#### 消息缓存

```go
// 存储消息
func UpdateUpMessageMap(key string, device model.UpMessage)

// 获取消息
func GetUpMessageMap(key string) (model.UpMessage, error)

// 根据复合键获取消息
func GetUpMessageByCompositeKey(deviceKey, messageId string) (model.UpMessage, error)

// 删除消息
func DeleteFromUpMessageMap(key string)

// 根据复合键删除消息
func DeleteFromUpMessageMapByCompositeKey(deviceKey, messageId string)

// 清理设备所有消息
func ClearDeviceMessages(deviceKey string)

// 获取缓存统计
func GetCacheStats() map[string]interface{}
```

#### 事件系统

```go
// 触发事件
func event.MustFire(eventName string, data interface{})

// 异步触发事件
func event.Async(eventName string, data interface{})

// 注册事件监听器
func event.On(eventName string, listener event.Listener, priority event.Priority)
```

### 常量定义

```go
// 事件类型
const (
    PushAttributeDataToMQTT  = "PushAttributeDataToMQTT"
    PushServiceResDataToMQTT = "PushServiceResDataToMQTT"
    PushSetResDataToMQTT     = "PushSetResDataToMQTT"
)

// 网络类型
const (
    NetTypeTcpServer  = "tcp"
    NetTypeUDPServer  = "udp"
    NetTypeMqttServer = "mqtt"
)

// 粘包处理类型
const (
    NoHandling         = 0
    FixedLength        = 1
    HeaderBodySeparate = 2
    Delimiter          = 3
)
```

---

## 版本历史

### v1.0.0 (当前版本)
- ✅ 基础网关功能
- ✅ TCP/UDP/MQTT支持
- ✅ 协议处理器接口
- ✅ 事件驱动架构
- ✅ MQTT集成
- ✅ 消息缓存优化
- ✅ 内存泄漏防护

### 路线图
- 🔄 WebSocket支持
- 🔄 插件系统
- 🔄 图形化配置界面
- 🔄 集群部署支持
- 🔄 更多协议模板

---

## 技术支持

- **文档**: [https://docs.sagoo.cn](https://docs.sagoo.cn)
- **示例项目**: [iotgateway-example](https://github.com/sagoo-cloud/iotgateway-example)
- **问题反馈**: [GitHub Issues](https://github.com/sagoo-cloud/iotgateway/issues)
- **技术交流**: QQ群 123456789

---

*本文档持续更新中，如有问题请及时反馈。* 