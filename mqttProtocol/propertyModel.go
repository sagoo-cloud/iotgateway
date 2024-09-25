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

// 属性上报结构体
type (
	ReportPropertyReq struct {
		Id      string                 `json:"id"`
		Version string                 `json:"version"`
		Sys     SysInfo                `json:"sys"`
		Params  map[string]interface{} `json:"params"`
		Method  string                 `json:"method"`
	}
	ReportPropertyReply struct {
		Code int `json:"code"`
		Data struct {
		} `json:"data"`
		Id      string `json:"id"`
		Message string `json:"message"`
		Method  string `json:"method"`
		Version string `json:"version"`
	}
)

// 属性设置结构体
type (
	PropertySetRequest struct {
		Id      string                 `json:"id"`
		Version string                 `json:"version"`
		Params  map[string]interface{} `json:"params"`
		Method  string                 `json:"method"`
	}
	PropertySetRes struct {
		Code    int                    `json:"code"`
		Data    map[string]interface{} `json:"data"`
		Id      string                 `json:"id"`
		Message string                 `json:"message"`
		Version string                 `json:"version"`
	}
)
