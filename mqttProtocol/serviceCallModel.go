package mqttProtocol

// 服务调用结构体
type (
	ServiceCallRequest struct {
		Id      string                 `json:"id"`
		Version string                 `json:"version"`
		Params  map[string]interface{} `json:"params"`
		Method  string                 `json:"method"`
	}

	ServiceCallOutputRes struct {
		Code    int                    `json:"code"`
		Data    map[string]interface{} `json:"data"`
		Id      string                 `json:"id"`
		Message string                 `json:"message"`
		Version string                 `json:"version"`
	}
)
