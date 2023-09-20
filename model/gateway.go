package model

// GatewayInfo 网关数据
type GatewayInfo struct {
	ProductKey   string `json:"productKey"`
	DeviceKey    string `json:"deviceKey"`
	DeviceName   string `json:"deviceName"`
	Description  string `json:"description"`
	DeviceType   string `json:"deviceType"`
	Version      string `json:"version"`
	Manufacturer string `json:"manufacturer"`
}

// UpMessage 上行消息
type UpMessage struct {
	MessageID   string `json:"messageId"`
	SendTime    int64  `json:"sendTime"`
	RequestCode string `json:"requestCode"`
	MethodName  string `json:"methodName"`
	Topic       string `json:"topic"`
}

// DownMessage 下行消息
type DownMessage struct {
	FuncCode      string `json:"funcCode"`
	ChannelNumber string `json:"channelNumber"`
	ErrorCode     string `json:"errorCode"`
}
