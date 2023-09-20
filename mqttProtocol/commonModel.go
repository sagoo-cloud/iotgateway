package mqttProtocol

type (
	SysInfo struct {
		Ack int `json:"ack"`
	}

	PropertyNode struct {
		Value      interface{} `json:"value"`
		CreateTime int64       `json:"time"`
	}
)
