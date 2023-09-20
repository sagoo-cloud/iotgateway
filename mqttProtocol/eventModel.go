package mqttProtocol

// 事件上报结构体
type (
	ReportEventReq struct {
		Id      string            `json:"id"`
		Version string            `json:"version"`
		Sys     SysInfo           `json:"sys"`
		Params  ReportEventParams `json:"params"`
	}
	ReportEventParams struct {
		Value    map[string]string `json:"value"`
		CreateAt int64             `json:"time"`
	}
	ReportEventReply struct {
		Code int `json:"code"`
		Data struct {
		} `json:"data"`
		Id      string `json:"id"`
		Message string `json:"message"`
		Method  string `json:"method"`
		Version string `json:"version"`
	}
)
