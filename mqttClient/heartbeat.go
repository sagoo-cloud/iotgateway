package mqttClient

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/sagoo-cloud/iotgateway/lib"
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"github.com/sagoo-cloud/iotgateway/vars"
	"github.com/sagoo-cloud/iotgateway/version"
	"time"
)

// heartbeat 网关服务心跳
func heartbeat(ctx context.Context, duration time.Duration) {
	if duration == 0 {
		duration = 30
	}
	ticker := time.NewTicker(time.Second * duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//设备数量
			versionInfo := version.GetVersion()
			if versionInfo == "" || versionInfo == "0.0" {
				versionInfo = "v0.0.1"
			}
			count := vars.CountDevices()
			builder := mqttProtocol.NewGatewayBatchReqBuilder().SetId(lib.RandString(16))
			builder.SetVersion("1.0")
			builder.AddProperty("Status", 0)
			builder.AddProperty("Count", count)
			builder.AddProperty("Version", versionInfo)
			builder.SetMethod("thing.event.property.pack.post")
			data := gconv.Map(builder.Build())
			outData := gjson.New(data).MustToJson()
			topic := fmt.Sprintf(propertyTopic, vars.Gateway.ProductKey, vars.Gateway.DeviceKey)
			g.Log().Debug(context.Background(), "---------网关向平台发送心跳数据-------：", outData)
			err := Publish(topic, outData)
			if err != nil {
				g.Log().Errorf(context.Background(), "publish error: %s", err.Error())
			}
		case <-ctx.Done():
			fmt.Println("Heartbeat stopped")
			return
		}
	}
}
