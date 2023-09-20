package mqttProtocol

import (
	"net"
)

type Protocol interface {
	Encode(args []byte) (res []byte, err error)
	Decode(conn net.Conn, buffer []byte) (res []byte, err error)
}
