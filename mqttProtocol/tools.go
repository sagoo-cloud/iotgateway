package mqttProtocol

type GatewayBatchReqBuilder struct {
	batchReq GatewayBatchReq
}

// NewGatewayBatchReqBuilder 创建一个新的GatewayBatchReqBuilder
func NewGatewayBatchReqBuilder() *GatewayBatchReqBuilder {
	return &GatewayBatchReqBuilder{
		batchReq: GatewayBatchReq{
			Sys: SysInfo{},
			Params: PropertyInfo{
				Properties: make(map[string]interface{}),
				Events:     make(map[string]EventNode),
				SubDevices: make([]Sub, 0),
			},
		},
	}
}

func (b *GatewayBatchReqBuilder) SetId(id string) *GatewayBatchReqBuilder {
	b.batchReq.Id = id
	return b
}

func (b *GatewayBatchReqBuilder) SetVersion(version string) *GatewayBatchReqBuilder {
	b.batchReq.Version = version
	return b
}

func (b *GatewayBatchReqBuilder) SetSys(sys SysInfo) *GatewayBatchReqBuilder {
	b.batchReq.Sys = sys
	return b
}

func (b *GatewayBatchReqBuilder) AddProperty(key string, value interface{}) *GatewayBatchReqBuilder {
	b.batchReq.Params.Properties[key] = value
	return b
}

func (b *GatewayBatchReqBuilder) AddEvent(key string, event EventNode) *GatewayBatchReqBuilder {
	b.batchReq.Params.Events[key] = event
	return b
}

func (b *GatewayBatchReqBuilder) AddSubDevice(sub Sub) *GatewayBatchReqBuilder {
	b.batchReq.Params.SubDevices = append(b.batchReq.Params.SubDevices, sub)
	return b
}

func (b *GatewayBatchReqBuilder) SetMethod(method string) *GatewayBatchReqBuilder {
	b.batchReq.Method = method
	return b
}

func (b *GatewayBatchReqBuilder) Build() GatewayBatchReq {
	return b.batchReq
}
