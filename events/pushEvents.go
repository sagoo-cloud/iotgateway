package events

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/gookit/event"
	"github.com/sagoo-cloud/iotgateway/consts"
	"github.com/sagoo-cloud/iotgateway/mqttClient"
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"github.com/sagoo-cloud/iotgateway/vars"
)

func LoadingPublishEvent() {
	event.On(consts.PushAttributeDataToMQTT, event.ListenerFunc(pushAttributeDataToMQTT), event.Normal)
	event.On(consts.PushServiceResDataToMQTT, event.ListenerFunc(pushServiceResDataToMQTT), event.High)
}

// pushAttributeDataToMQTT 推送属性数据到mqtt服务
func pushAttributeDataToMQTT(e event.Event) (err error) {
	deviceKey := gconv.String(e.Data()["DeviceKey"])
	if deviceKey == "" {
		return errors.New("设备key为空")
	}
	// propertieData 属性信息
	var propertieData = make(map[string]interface{})
	if e.Data()["PropertieDataList"] != nil {
		propertieDataLIst := gconv.Map(e.Data()["PropertieDataList"])
		for k, v := range propertieDataLIst {
			var param = mqttProtocol.PropertyNode{}
			switch v.(type) {
			case int, int32, int64,
				uint, uint16, uint64,
				float32, float64,
				bool, string:
				param.Value = v
			default:
				param.Value = gconv.Map(v)
			}

			param.CreateTime = gtime.Timestamp()
			propertieData[k] = param
		}
	}

	//eventsData 事件信息
	var eventsData = make(map[string]mqttProtocol.EventNode)
	if e.Data()["EventDataList"] != nil {
		eventDataList := gconv.Map(e.Data()["EventDataList"])
		for k, v := range eventDataList {
			var param = mqttProtocol.EventNode{}
			param.Value = gconv.Map(v)
			param.CreateTime = gtime.Timestamp()
			eventsData[k] = param
		}
	}

	//子设备
	subDevice := mqttProtocol.Sub{
		Identity:   mqttProtocol.Identity{ProductKey: "", DeviceKey: deviceKey},
		Properties: propertieData,
		Events:     eventsData,
	}

	builder := mqttProtocol.NewGatewayBatchReqBuilder()
	builder.SetId(guid.S()).SetVersion("1.0")
	builder.AddSubDevice(subDevice)
	builder.SetMethod("hing.event.property.pack.post")
	builder.Build()
	data := gconv.Map(builder.Build())
	outData := gjson.New(data).MustToJson()
	g.Log().Debugf(context.Background(), "设备Key：%v，推送【属性数据】到MQTT服务：%s", deviceKey, outData)
	if err = mqttClient.PublishData(deviceKey, outData); err != nil {
		g.Log().Debug(context.Background(), "pushAttributeDataToMQTT", err.Error())
		return
	}
	return
}

// pushServiceResDataToMQTT 推送服务调用响应数据到mqtt服务
func pushServiceResDataToMQTT(e event.Event) (err error) {
	deviceKey := gconv.String(e.Data()["DeviceKey"])
	replyData := e.Data()["ReplyData"]
	replyDataMap := make(map[string]interface{})
	if replyData != nil {
		replyDataMap = gconv.Map(replyData)
	}

	msg, err := vars.GetUpMessageMap(deviceKey)
	if msg.MessageID != "" && err == nil {
		g.Log().Debug(context.Background(), "==5555==监听回复信息====", msg)
		mqData := mqttProtocol.ServiceCallOutputRes{}
		mqData.Id = msg.MessageID
		mqData.Code = 200
		mqData.Message = "success"
		mqData.Version = "1.0"
		mqData.Data = replyDataMap

		//推送数据到mqtt
		topic := msg.Topic + "_reply"

		g.Log().Debugf(context.Background(), "设备Key：%v，推送【服务调用应答数据】到MQTT服务：%v", deviceKey, mqData)
		outData, err := json.Marshal(mqData)
		if err != nil {
			g.Log().Debug(context.Background(), "服务回调响应序列化失败：", err.Error())
			return err
		}

		g.Log().Debug(context.Background(), "服务回调响应：", mqData)
		g.Log().Debug(context.Background(), "服务回调响应topic：", topic)
		err = mqttClient.Publish(topic, outData)
		if err != nil {
			g.Log().Debug(context.Background(), "发送服务回调响应失败：", err.Error())
		} else {
			g.Log().Debug(context.Background(), "发送服务回调响应成功")
		}
	}
	return
}
