package network

import (
	"github.com/sagoo-cloud/iotgateway/log"
	"os"
)

// NewConnectionListener 新连接事件监听器
type NewConnectionListener struct{}

func (l *NewConnectionListener) OnEvent(event *Event) {
	log.Debug("New connection:", event.Conn.RemoteAddr().String())
}

// DataReceivedListener 数据接收事件监听器
type DataReceivedListener struct{}

func (l *DataReceivedListener) OnEvent(event *Event) {

	// 解包
	pgResData, err := event.Protocol.Decode(event.Conn, event.Data)
	if err != nil {
		//	fmt.Printf("接收到数据: %x\n", event.Data)
		//	log.Error("通过协议解析数据出错: %s", err.Error())
		return
	}

	// 发送应答响应
	echo, err := event.Protocol.Encode(pgResData)
	if len(echo) > 0 && err == nil {
		if err != nil {
			log.Error("%s unable to encode data: %v\n", os.Stderr, err)
			return
		}
		_, err := event.Conn.Write(echo)
		if err != nil {
			log.Error("Error occurred:", err.Error())
			return
		}
	}
	return
}

// ConnectionClosedListener 连接关闭事件监听器
type ConnectionClosedListener struct{}

func (l *ConnectionClosedListener) OnEvent(event *Event) {
	log.Debug("Connection closed:", event.Conn.RemoteAddr().String())
}
