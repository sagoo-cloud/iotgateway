# 协议开发指南

## 概述

本指南详细介绍如何为SagooIOT网关SDK开发自定义协议处理器，包括协议接口实现、数据解析、编码处理等核心内容。

## 协议接口定义

所有协议处理器都必须实现 `ProtocolHandler` 接口：

```go
type ProtocolHandler interface {
    Init(device *model.Device, data []byte) error
    Encode(device *model.Device, data interface{}, param ...string) ([]byte, error)
    Decode(device *model.Device, data []byte) ([]byte, error)
}
```

## 方法详解

### 1. Init 方法

**作用：** 初始化设备连接，设置设备标识，订阅平台事件

**调用时机：** 设备首次连接或发送数据时

**参数说明：**
- `device`: 设备对象，可能为nil（需要判断）
- `data`: 初始数据包

**实现要点：**

```go
func (p *MyProtocol) Init(device *model.Device, data []byte) error {
    // 1. 数据有效性检查
    if len(data) < 4 {
        return errors.New("初始数据包长度不足")
    }
    
    // 2. 解析设备标识
    deviceKey, err := p.parseDeviceKey(data)
    if err != nil {
        return fmt.Errorf("解析设备标识失败: %v", err)
    }
    
    // 3. 设置设备信息
    if device != nil {
        device.DeviceKey = deviceKey
        device.OnlineStatus = true
        device.LastActive = time.Now()
        
        // 设置设备元数据
        device.Metadata = map[string]interface{}{
            "protocol": "MyProtocol",
            "version":  "1.0",
        }
    }
    
    // 4. 订阅平台下发事件
    if deviceKey != "" {
        iotgateway.ServerGateway.SubscribeServiceEvent(deviceKey)
        iotgateway.ServerGateway.SubscribeSetEvent(deviceKey)
    }
    
    // 5. 记录设备上线日志
    glog.Infof(context.Background(), "设备上线: %s", deviceKey)
    
    return nil
}
```

### 2. Decode 方法

**作用：** 解析设备上行数据，触发相应的数据上报事件

**调用时机：** 每次收到设备数据时

**参数说明：**
- `device`: 设备对象
- `data`: 接收到的原始数据

**返回值：**
- `[]byte`: 需要回复给设备的数据（可以为nil）
- `error`: 解析错误

**实现模板：**

```go
func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    // 1. 异常恢复
    defer func() {
        if r := recover(); r != nil {
            glog.Errorf(context.Background(), "协议解析异常: %v", r)
        }
    }()
    
    // 2. 参数验证
    if device == nil {
        return nil, errors.New("设备对象为空")
    }
    
    if len(data) == 0 {
        return nil, errors.New("数据为空")
    }
    
    // 3. 更新设备活跃时间
    device.LastActive = time.Now()
    
    // 4. 解析数据包
    packet, err := p.parsePacket(data)
    if err != nil {
        glog.Errorf(context.Background(), "解析数据包失败: deviceKey=%s, error=%v", 
            device.DeviceKey, err)
        return nil, err
    }
    
    // 5. 根据数据包类型处理
    switch packet.Type {
    case PacketTypeHeartbeat:
        return p.handleHeartbeat(device, packet)
    case PacketTypeData:
        return p.handleData(device, packet)
    case PacketTypeAlarm:
        return p.handleAlarm(device, packet)
    case PacketTypeResponse:
        return p.handleResponse(device, packet)
    default:
        return nil, fmt.Errorf("未知数据包类型: %d", packet.Type)
    }
}
```

### 3. Encode 方法

**作用：** 编码下行数据，将平台指令转换为设备可识别的格式

**调用时机：** 需要向设备发送数据时

**参数说明：**
- `device`: 目标设备对象
- `data`: 要发送的数据
- `param`: 可选参数

**实现模板：**

```go
func (p *MyProtocol) Encode(device *model.Device, data interface{}, param ...string) ([]byte, error) {
    if device == nil {
        return nil, errors.New("设备对象为空")
    }
    
    // 根据数据类型进行编码
    switch v := data.(type) {
    case map[string]interface{}:
        return p.encodeCommand(device, v)
    case string:
        return p.encodeString(device, v)
    case []byte:
        return v, nil
    default:
        // 默认JSON编码
        return json.Marshal(data)
    }
}
```

## 协议开发实例

### 示例1：简单文本协议

假设设备使用简单的文本协议，格式为：`DEVICE_ID|DATA_TYPE|DATA|CRC`

```go
package protocol

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
    "time"
    
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/glog"
    "github.com/gogf/gf/v2/util/gconv"
    "github.com/gookit/event"
    "github.com/sagoo-cloud/iotgateway"
    "github.com/sagoo-cloud/iotgateway/consts"
    "github.com/sagoo-cloud/iotgateway/model"
)

type TextProtocol struct{}

// 数据包类型
const (
    DataTypeHeartbeat = "HB"
    DataTypeData      = "DATA"
    DataTypeAlarm     = "ALARM"
    DataTypeResponse  = "RESP"
)

// 数据包结构
type TextPacket struct {
    DeviceID string
    DataType string
    Data     string
    CRC      string
}

func (p *TextProtocol) Init(device *model.Device, data []byte) error {
    // 解析初始数据包获取设备ID
    packet, err := p.parseTextPacket(data)
    if err != nil {
        return err
    }
    
    if device != nil {
        device.DeviceKey = packet.DeviceID
        device.OnlineStatus = true
        device.LastActive = time.Now()
    }
    
    // 订阅平台事件
    if packet.DeviceID != "" {
        iotgateway.ServerGateway.SubscribeServiceEvent(packet.DeviceID)
        iotgateway.ServerGateway.SubscribeSetEvent(packet.DeviceID)
    }
    
    glog.Infof(context.Background(), "文本协议设备上线: %s", packet.DeviceID)
    return nil
}

func (p *TextProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    defer func() {
        if r := recover(); r != nil {
            glog.Errorf(context.Background(), "文本协议解析异常: %v", r)
        }
    }()
    
    if device == nil || len(data) == 0 {
        return nil, errors.New("无效参数")
    }
    
    // 解析数据包
    packet, err := p.parseTextPacket(data)
    if err != nil {
        return nil, err
    }
    
    // 验证CRC
    if !p.verifyCRC(packet) {
        return nil, errors.New("CRC校验失败")
    }
    
    // 更新设备活跃时间
    device.LastActive = time.Now()
    
    // 处理不同类型的数据
    switch packet.DataType {
    case DataTypeHeartbeat:
        return p.handleHeartbeat(device, packet)
    case DataTypeData:
        return p.handleData(device, packet)
    case DataTypeAlarm:
        return p.handleAlarm(device, packet)
    default:
        return nil, fmt.Errorf("未知数据类型: %s", packet.DataType)
    }
}

func (p *TextProtocol) Encode(device *model.Device, data interface{}, param ...string) ([]byte, error) {
    if device == nil {
        return nil, errors.New("设备对象为空")
    }
    
    switch v := data.(type) {
    case map[string]interface{}:
        return p.encodeCommand(device, v)
    case string:
        return p.encodeString(device, v)
    default:
        return nil, errors.New("不支持的数据类型")
    }
}

// 解析文本数据包
func (p *TextProtocol) parseTextPacket(data []byte) (*TextPacket, error) {
    text := strings.TrimSpace(string(data))
    parts := strings.Split(text, "|")
    
    if len(parts) != 4 {
        return nil, errors.New("数据包格式错误")
    }
    
    return &TextPacket{
        DeviceID: parts[0],
        DataType: parts[1],
        Data:     parts[2],
        CRC:      parts[3],
    }, nil
}

// 验证CRC
func (p *TextProtocol) verifyCRC(packet *TextPacket) bool {
    // 简单的CRC校验实现
    expected := p.calculateCRC(packet.DeviceID + "|" + packet.DataType + "|" + packet.Data)
    return expected == packet.CRC
}

// 计算CRC
func (p *TextProtocol) calculateCRC(data string) string {
    // 简单的校验和实现
    sum := 0
    for _, b := range []byte(data) {
        sum += int(b)
    }
    return fmt.Sprintf("%04X", sum&0xFFFF)
}

// 处理心跳数据
func (p *TextProtocol) handleHeartbeat(device *model.Device, packet *TextPacket) ([]byte, error) {
    glog.Debugf(context.Background(), "收到心跳: deviceKey=%s", device.DeviceKey)
    
    // 回复心跳确认
    response := fmt.Sprintf("%s|HB_ACK|OK|%s", 
        device.DeviceKey, 
        p.calculateCRC(device.DeviceKey+"|HB_ACK|OK"))
    
    return []byte(response), nil
}

// 处理数据上报
func (p *TextProtocol) handleData(device *model.Device, packet *TextPacket) ([]byte, error) {
    // 解析数据字段
    dataFields := strings.Split(packet.Data, ",")
    properties := make(map[string]interface{})
    
    for _, field := range dataFields {
        kv := strings.Split(field, "=")
        if len(kv) == 2 {
            // 尝试转换为数值
            if val, err := strconv.ParseFloat(kv[1], 64); err == nil {
                properties[kv[0]] = val
            } else {
                properties[kv[0]] = kv[1]
            }
        }
    }
    
    // 添加时间戳
    properties["timestamp"] = time.Now().Unix()
    
    // 触发属性上报事件
    eventData := g.Map{
        "DeviceKey":         device.DeviceKey,
        "PropertieDataList": properties,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
    
    glog.Infof(context.Background(), "设备数据上报: deviceKey=%s, data=%+v", 
        device.DeviceKey, properties)
    
    // 回复确认
    response := fmt.Sprintf("%s|DATA_ACK|OK|%s", 
        device.DeviceKey, 
        p.calculateCRC(device.DeviceKey+"|DATA_ACK|OK"))
    
    return []byte(response), nil
}

// 处理报警数据
func (p *TextProtocol) handleAlarm(device *model.Device, packet *TextPacket) ([]byte, error) {
    // 解析报警数据
    alarmData := map[string]interface{}{
        "level":   "high",
        "message": packet.Data,
        "time":    time.Now().Unix(),
    }
    
    events := map[string]interface{}{
        "alarm": alarmData,
    }
    
    // 触发事件上报
    eventData := g.Map{
        "DeviceKey":     device.DeviceKey,
        "EventDataList": events,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
    
    glog.Warnf(context.Background(), "设备报警: deviceKey=%s, alarm=%s", 
        device.DeviceKey, packet.Data)
    
    // 回复确认
    response := fmt.Sprintf("%s|ALARM_ACK|OK|%s", 
        device.DeviceKey, 
        p.calculateCRC(device.DeviceKey+"|ALARM_ACK|OK"))
    
    return []byte(response), nil
}

// 编码命令
func (p *TextProtocol) encodeCommand(device *model.Device, cmd map[string]interface{}) ([]byte, error) {
    command := gconv.String(cmd["command"])
    params := gconv.String(cmd["params"])
    
    data := fmt.Sprintf("%s,%s", command, params)
    packet := fmt.Sprintf("%s|CMD|%s", device.DeviceKey, data)
    crc := p.calculateCRC(packet)
    
    result := fmt.Sprintf("%s|%s", packet, crc)
    return []byte(result), nil
}

// 编码字符串
func (p *TextProtocol) encodeString(device *model.Device, str string) ([]byte, error) {
    packet := fmt.Sprintf("%s|MSG|%s", device.DeviceKey, str)
    crc := p.calculateCRC(packet)
    
    result := fmt.Sprintf("%s|%s", packet, crc)
    return []byte(result), nil
}
```

### 示例2：二进制协议

假设设备使用二进制协议，格式为：`[Header(2)][Length(2)][DeviceID(4)][Type(1)][Data(N)][CRC(2)]`

```go
package protocol

import (
    "bytes"
    "encoding/binary"
    "errors"
    "fmt"
    "time"
    
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/glog"
    "github.com/gookit/event"
    "github.com/sagoo-cloud/iotgateway"
    "github.com/sagoo-cloud/iotgateway/consts"
    "github.com/sagoo-cloud/iotgateway/model"
)

type BinaryProtocol struct{}

// 协议常量
const (
    ProtocolHeader    = 0xAABB
    MinPacketLength   = 11 // 最小包长度
    
    // 数据类型
    TypeHeartbeat = 0x01
    TypeData      = 0x02
    TypeAlarm     = 0x03
    TypeCommand   = 0x04
)

// 二进制数据包结构
type BinaryPacket struct {
    Header   uint16
    Length   uint16
    DeviceID uint32
    Type     uint8
    Data     []byte
    CRC      uint16
}

func (p *BinaryProtocol) Init(device *model.Device, data []byte) error {
    packet, err := p.parseBinaryPacket(data)
    if err != nil {
        return err
    }
    
    if device != nil {
        device.DeviceKey = fmt.Sprintf("device_%08X", packet.DeviceID)
        device.OnlineStatus = true
        device.LastActive = time.Now()
    }
    
    deviceKey := fmt.Sprintf("device_%08X", packet.DeviceID)
    if deviceKey != "" {
        iotgateway.ServerGateway.SubscribeServiceEvent(deviceKey)
        iotgateway.ServerGateway.SubscribeSetEvent(deviceKey)
    }
    
    glog.Infof(context.Background(), "二进制协议设备上线: %s", deviceKey)
    return nil
}

func (p *BinaryProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    defer func() {
        if r := recover(); r != nil {
            glog.Errorf(context.Background(), "二进制协议解析异常: %v", r)
        }
    }()
    
    if device == nil || len(data) < MinPacketLength {
        return nil, errors.New("无效参数")
    }
    
    packet, err := p.parseBinaryPacket(data)
    if err != nil {
        return nil, err
    }
    
    // 验证CRC
    if !p.verifyCRC(packet) {
        return nil, errors.New("CRC校验失败")
    }
    
    device.LastActive = time.Now()
    
    switch packet.Type {
    case TypeHeartbeat:
        return p.handleHeartbeat(device, packet)
    case TypeData:
        return p.handleData(device, packet)
    case TypeAlarm:
        return p.handleAlarm(device, packet)
    default:
        return nil, fmt.Errorf("未知数据类型: %d", packet.Type)
    }
}

func (p *BinaryProtocol) Encode(device *model.Device, data interface{}, param ...string) ([]byte, error) {
    if device == nil {
        return nil, errors.New("设备对象为空")
    }
    
    // 解析设备ID
    var deviceID uint32
    fmt.Sscanf(device.DeviceKey, "device_%08X", &deviceID)
    
    switch v := data.(type) {
    case map[string]interface{}:
        return p.encodeCommand(deviceID, v)
    case []byte:
        return p.encodeData(deviceID, TypeCommand, v)
    default:
        return nil, errors.New("不支持的数据类型")
    }
}

// 解析二进制数据包
func (p *BinaryProtocol) parseBinaryPacket(data []byte) (*BinaryPacket, error) {
    if len(data) < MinPacketLength {
        return nil, errors.New("数据包长度不足")
    }
    
    reader := bytes.NewReader(data)
    packet := &BinaryPacket{}
    
    // 读取头部
    if err := binary.Read(reader, binary.BigEndian, &packet.Header); err != nil {
        return nil, err
    }
    
    if packet.Header != ProtocolHeader {
        return nil, errors.New("无效的协议头")
    }
    
    // 读取长度
    if err := binary.Read(reader, binary.BigEndian, &packet.Length); err != nil {
        return nil, err
    }
    
    if int(packet.Length) != len(data) {
        return nil, errors.New("数据包长度不匹配")
    }
    
    // 读取设备ID
    if err := binary.Read(reader, binary.BigEndian, &packet.DeviceID); err != nil {
        return nil, err
    }
    
    // 读取类型
    if err := binary.Read(reader, binary.BigEndian, &packet.Type); err != nil {
        return nil, err
    }
    
    // 读取数据
    dataLen := int(packet.Length) - MinPacketLength
    if dataLen > 0 {
        packet.Data = make([]byte, dataLen)
        if _, err := reader.Read(packet.Data); err != nil {
            return nil, err
        }
    }
    
    // 读取CRC
    if err := binary.Read(reader, binary.BigEndian, &packet.CRC); err != nil {
        return nil, err
    }
    
    return packet, nil
}

// 验证CRC
func (p *BinaryProtocol) verifyCRC(packet *BinaryPacket) bool {
    expected := p.calculateCRC(packet)
    return expected == packet.CRC
}

// 计算CRC
func (p *BinaryProtocol) calculateCRC(packet *BinaryPacket) uint16 {
    // 简单的CRC16实现
    var crc uint16 = 0xFFFF
    
    // 计算除CRC外的所有字段
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.BigEndian, packet.Header)
    binary.Write(buf, binary.BigEndian, packet.Length)
    binary.Write(buf, binary.BigEndian, packet.DeviceID)
    binary.Write(buf, binary.BigEndian, packet.Type)
    buf.Write(packet.Data)
    
    data := buf.Bytes()
    for _, b := range data {
        crc ^= uint16(b)
        for i := 0; i < 8; i++ {
            if crc&1 != 0 {
                crc = (crc >> 1) ^ 0xA001
            } else {
                crc >>= 1
            }
        }
    }
    
    return crc
}

// 处理心跳
func (p *BinaryProtocol) handleHeartbeat(device *model.Device, packet *BinaryPacket) ([]byte, error) {
    glog.Debugf(context.Background(), "收到心跳: deviceKey=%s", device.DeviceKey)
    
    // 构造心跳回复
    return p.encodeData(packet.DeviceID, TypeHeartbeat, []byte{0x01}), nil
}

// 处理数据
func (p *BinaryProtocol) handleData(device *model.Device, packet *BinaryPacket) ([]byte, error) {
    if len(packet.Data) < 8 {
        return nil, errors.New("数据长度不足")
    }
    
    // 解析数据（假设为温度和湿度）
    reader := bytes.NewReader(packet.Data)
    var temperature, humidity float32
    
    binary.Read(reader, binary.BigEndian, &temperature)
    binary.Read(reader, binary.BigEndian, &humidity)
    
    properties := map[string]interface{}{
        "temperature": temperature,
        "humidity":    humidity,
        "timestamp":   time.Now().Unix(),
    }
    
    // 触发属性上报
    eventData := g.Map{
        "DeviceKey":         device.DeviceKey,
        "PropertieDataList": properties,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
    
    glog.Infof(context.Background(), "设备数据上报: deviceKey=%s, temp=%.2f, hum=%.2f", 
        device.DeviceKey, temperature, humidity)
    
    // 回复确认
    return p.encodeData(packet.DeviceID, TypeData, []byte{0x01}), nil
}

// 处理报警
func (p *BinaryProtocol) handleAlarm(device *model.Device, packet *BinaryPacket) ([]byte, error) {
    if len(packet.Data) < 1 {
        return nil, errors.New("报警数据为空")
    }
    
    alarmCode := packet.Data[0]
    alarmData := map[string]interface{}{
        "code":    alarmCode,
        "level":   p.getAlarmLevel(alarmCode),
        "message": p.getAlarmMessage(alarmCode),
        "time":    time.Now().Unix(),
    }
    
    events := map[string]interface{}{
        "alarm": alarmData,
    }
    
    eventData := g.Map{
        "DeviceKey":     device.DeviceKey,
        "EventDataList": events,
    }
    event.MustFire(consts.PushAttributeDataToMQTT, eventData)
    
    glog.Warnf(context.Background(), "设备报警: deviceKey=%s, code=%d", 
        device.DeviceKey, alarmCode)
    
    return p.encodeData(packet.DeviceID, TypeAlarm, []byte{0x01}), nil
}

// 编码数据
func (p *BinaryProtocol) encodeData(deviceID uint32, dataType uint8, data []byte) []byte {
    length := uint16(MinPacketLength + len(data))
    
    buf := new(bytes.Buffer)
    
    // 写入头部信息
    binary.Write(buf, binary.BigEndian, ProtocolHeader)
    binary.Write(buf, binary.BigEndian, length)
    binary.Write(buf, binary.BigEndian, deviceID)
    binary.Write(buf, binary.BigEndian, dataType)
    buf.Write(data)
    
    // 计算CRC
    packet := &BinaryPacket{
        Header:   ProtocolHeader,
        Length:   length,
        DeviceID: deviceID,
        Type:     dataType,
        Data:     data,
    }
    crc := p.calculateCRC(packet)
    binary.Write(buf, binary.BigEndian, crc)
    
    return buf.Bytes()
}

// 编码命令
func (p *BinaryProtocol) encodeCommand(deviceID uint32, cmd map[string]interface{}) ([]byte, error) {
    // 根据命令类型编码
    cmdType := cmd["type"].(string)
    
    switch cmdType {
    case "restart":
        return p.encodeData(deviceID, TypeCommand, []byte{0x01}), nil
    case "config":
        // 编码配置命令
        return p.encodeConfigCommand(deviceID, cmd)
    default:
        return nil, fmt.Errorf("不支持的命令类型: %s", cmdType)
    }
}

// 编码配置命令
func (p *BinaryProtocol) encodeConfigCommand(deviceID uint32, cmd map[string]interface{}) ([]byte, error) {
    buf := new(bytes.Buffer)
    
    // 命令码
    buf.WriteByte(0x02)
    
    // 配置参数
    if interval, ok := cmd["interval"]; ok {
        binary.Write(buf, binary.BigEndian, uint16(interval.(int)))
    }
    
    if threshold, ok := cmd["threshold"]; ok {
        binary.Write(buf, binary.BigEndian, float32(threshold.(float64)))
    }
    
    return p.encodeData(deviceID, TypeCommand, buf.Bytes()), nil
}

// 获取报警级别
func (p *BinaryProtocol) getAlarmLevel(code uint8) string {
    switch code {
    case 1:
        return "low"
    case 2:
        return "medium"
    case 3:
        return "high"
    default:
        return "unknown"
    }
}

// 获取报警消息
func (p *BinaryProtocol) getAlarmMessage(code uint8) string {
    switch code {
    case 1:
        return "温度异常"
    case 2:
        return "湿度异常"
    case 3:
        return "设备故障"
    default:
        return "未知报警"
    }
}
```

## 协议开发最佳实践

### 1. 错误处理

```go
func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    // 1. 使用defer恢复panic
    defer func() {
        if r := recover(); r != nil {
            glog.Errorf(context.Background(), "协议解析panic: %v", r)
        }
    }()
    
    // 2. 参数验证
    if device == nil {
        return nil, errors.New("设备对象为空")
    }
    
    if len(data) == 0 {
        return nil, errors.New("数据为空")
    }
    
    // 3. 数据长度检查
    if len(data) < p.getMinPacketLength() {
        return nil, fmt.Errorf("数据包长度不足，期望至少%d字节，实际%d字节", 
            p.getMinPacketLength(), len(data))
    }
    
    // 4. 协议头验证
    if !p.isValidHeader(data) {
        return nil, errors.New("无效的协议头")
    }
    
    // 5. 详细的错误信息
    packet, err := p.parsePacket(data)
    if err != nil {
        return nil, fmt.Errorf("解析数据包失败: %v, 原始数据: %x", err, data)
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
    defer func() {
        packet.Reset() // 重置对象状态
        packetPool.Put(packet)
    }()
    
    // 解析数据
    err := packet.Parse(data)
    if err != nil {
        return nil, err
    }
    
    return p.processPacket(device, packet)
}

// 使用字节缓冲池
var bufferPool = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 1024))
    },
}

func (p *MyProtocol) Encode(device *model.Device, data interface{}, param ...string) ([]byte, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    // 编码数据到缓冲区
    err := p.encodeToBuffer(buf, data)
    if err != nil {
        return nil, err
    }
    
    // 复制数据（因为缓冲区会被重用）
    result := make([]byte, buf.Len())
    copy(result, buf.Bytes())
    
    return result, nil
}
```

### 3. 并发安全

```go
type SafeProtocol struct {
    mu       sync.RWMutex
    devices  map[string]*DeviceInfo
    counters map[string]int64
}

type DeviceInfo struct {
    LastSeen    time.Time
    PacketCount int64
    ErrorCount  int64
}

func (p *SafeProtocol) updateDeviceInfo(deviceKey string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if info, exists := p.devices[deviceKey]; exists {
        info.LastSeen = time.Now()
        info.PacketCount++
    } else {
        p.devices[deviceKey] = &DeviceInfo{
            LastSeen:    time.Now(),
            PacketCount: 1,
            ErrorCount:  0,
        }
    }
}

func (p *SafeProtocol) getDeviceInfo(deviceKey string) *DeviceInfo {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    if info, exists := p.devices[deviceKey]; exists {
        // 返回副本避免并发修改
        return &DeviceInfo{
            LastSeen:    info.LastSeen,
            PacketCount: info.PacketCount,
            ErrorCount:  info.ErrorCount,
        }
    }
    return nil
}
```

### 4. 日志记录

```go
func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    ctx := context.Background()
    
    // 记录接收数据
    glog.Debugf(ctx, "收到设备数据: deviceKey=%s, dataLen=%d, data=%x", 
        device.DeviceKey, len(data), data)
    
    packet, err := p.parsePacket(data)
    if err != nil {
        // 记录解析错误
        glog.Errorf(ctx, "解析数据包失败: deviceKey=%s, error=%v, data=%x", 
            device.DeviceKey, err, data)
        return nil, err
    }
    
    // 记录解析成功
    glog.Infof(ctx, "成功解析数据包: deviceKey=%s, type=%s, dataLen=%d", 
        device.DeviceKey, packet.GetTypeName(), len(packet.Data))
    
    result, err := p.processPacket(device, packet)
    if err != nil {
        glog.Errorf(ctx, "处理数据包失败: deviceKey=%s, error=%v", 
            device.DeviceKey, err)
        return nil, err
    }
    
    if result != nil {
        glog.Debugf(ctx, "回复设备数据: deviceKey=%s, replyLen=%d, reply=%x", 
            device.DeviceKey, len(result), result)
    }
    
    return result, nil
}
```

### 5. 配置化协议

```go
type ProtocolConfig struct {
    Header       []byte        `json:"header"`
    MinLength    int           `json:"minLength"`
    MaxLength    int           `json:"maxLength"`
    Timeout      time.Duration `json:"timeout"`
    EnableCRC    bool          `json:"enableCRC"`
    ByteOrder    string        `json:"byteOrder"` // "big" or "little"
}

type ConfigurableProtocol struct {
    config *ProtocolConfig
}

func NewConfigurableProtocol(config *ProtocolConfig) *ConfigurableProtocol {
    return &ConfigurableProtocol{
        config: config,
    }
}

func (p *ConfigurableProtocol) isValidPacket(data []byte) bool {
    if len(data) < p.config.MinLength {
        return false
    }
    
    if len(data) > p.config.MaxLength {
        return false
    }
    
    if len(p.config.Header) > 0 {
        if len(data) < len(p.config.Header) {
            return false
        }
        
        if !bytes.Equal(data[:len(p.config.Header)], p.config.Header) {
            return false
        }
    }
    
    return true
}
```

## 测试协议

### 单元测试

```go
package protocol

import (
    "testing"
    "github.com/sagoo-cloud/iotgateway/model"
)

func TestTextProtocol_ParsePacket(t *testing.T) {
    protocol := &TextProtocol{}
    
    tests := []struct {
        name    string
        data    []byte
        want    *TextPacket
        wantErr bool
    }{
        {
            name: "valid heartbeat packet",
            data: []byte("DEVICE001|HB|OK|1234"),
            want: &TextPacket{
                DeviceID: "DEVICE001",
                DataType: "HB",
                Data:     "OK",
                CRC:      "1234",
            },
            wantErr: false,
        },
        {
            name:    "invalid packet format",
            data:    []byte("DEVICE001|HB"),
            want:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := protocol.parseTextPacket(tt.data)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseTextPacket() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr && got != nil {
                if got.DeviceID != tt.want.DeviceID ||
                   got.DataType != tt.want.DataType ||
                   got.Data != tt.want.Data ||
                   got.CRC != tt.want.CRC {
                    t.Errorf("parseTextPacket() = %+v, want %+v", got, tt.want)
                }
            }
        })
    }
}

func TestTextProtocol_Decode(t *testing.T) {
    protocol := &TextProtocol{}
    device := &model.Device{
        DeviceKey: "DEVICE001",
    }
    
    // 测试心跳数据
    data := []byte("DEVICE001|HB|OK|1234")
    result, err := protocol.Decode(device, data)
    
    if err != nil {
        t.Errorf("Decode() error = %v", err)
    }
    
    if result == nil {
        t.Error("Decode() result should not be nil")
    }
}
```

### 集成测试

```go
func TestProtocolIntegration(t *testing.T) {
    // 创建测试网关
    ctx := context.Background()
    protocol := &TextProtocol{}
    gateway, err := iotgateway.NewGateway(ctx, protocol)
    if err != nil {
        t.Fatalf("创建网关失败: %v", err)
    }
    
    // 模拟设备连接
    device := &model.Device{
        DeviceKey: "TEST_DEVICE",
        ClientID:  "client_001",
    }
    
    // 测试初始化
    initData := []byte("TEST_DEVICE|HB|INIT|1234")
    err = protocol.Init(device, initData)
    if err != nil {
        t.Errorf("Init() error = %v", err)
    }
    
    // 测试数据解析
    testData := []byte("TEST_DEVICE|DATA|temp=25.6,hum=60.5|5678")
    result, err := protocol.Decode(device, testData)
    if err != nil {
        t.Errorf("Decode() error = %v", err)
    }
    
    if result == nil {
        t.Error("Decode() should return response")
    }
    
    // 测试编码
    cmd := map[string]interface{}{
        "command": "restart",
        "params":  "delay=5",
    }
    encoded, err := protocol.Encode(device, cmd)
    if err != nil {
        t.Errorf("Encode() error = %v", err)
    }
    
    if len(encoded) == 0 {
        t.Error("Encode() should return data")
    }
}
```

## 调试技巧

### 1. 数据包分析

```go
func (p *MyProtocol) debugPacket(data []byte) {
    glog.Debugf(context.Background(), "原始数据包分析:")
    glog.Debugf(context.Background(), "  长度: %d", len(data))
    glog.Debugf(context.Background(), "  十六进制: %x", data)
    glog.Debugf(context.Background(), "  字符串: %q", string(data))
    
    // 按字节分析
    for i, b := range data {
        glog.Debugf(context.Background(), "  [%02d]: 0x%02X (%d) '%c'", 
            i, b, b, printableChar(b))
    }
}

func printableChar(b byte) rune {
    if b >= 32 && b <= 126 {
        return rune(b)
    }
    return '.'
}
```

### 2. 性能监控

```go
func (p *MyProtocol) Decode(device *model.Device, data []byte) ([]byte, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > 100*time.Millisecond {
            glog.Warnf(context.Background(), "协议解析耗时过长: deviceKey=%s, duration=%v", 
                device.DeviceKey, duration)
        }
    }()
    
    // 解析逻辑...
}
```

### 3. 状态监控

```go
type ProtocolStats struct {
    TotalPackets   int64
    SuccessPackets int64
    ErrorPackets   int64
    LastError      string
    LastErrorTime  time.Time
}

func (p *MyProtocol) updateStats(success bool, err error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    p.stats.TotalPackets++
    if success {
        p.stats.SuccessPackets++
    } else {
        p.stats.ErrorPackets++
        if err != nil {
            p.stats.LastError = err.Error()
            p.stats.LastErrorTime = time.Now()
        }
    }
}

func (p *MyProtocol) GetStats() ProtocolStats {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    return p.stats
}
```

通过以上指南，您可以开发出高质量、高性能的协议处理器，满足各种设备接入需求。 