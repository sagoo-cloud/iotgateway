package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/gookit/event"
	"github.com/sagoo-cloud/iotgateway/consts"
	"github.com/sagoo-cloud/iotgateway/lib"
	"github.com/sagoo-cloud/iotgateway/log"
	"github.com/sagoo-cloud/iotgateway/mqttClient"
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"github.com/sagoo-cloud/iotgateway/vars"
	"github.com/sagoo-cloud/iotgateway/version"
)

// LoadingPublishEvent 加载发布事件
func LoadingPublishEvent() {
	//推送属性数据到mqtt服务事件
	event.On(consts.PushAttributeDataToMQTT, event.ListenerFunc(pushAttributeDataToMQTT), event.Normal)
	//推送服务调用响应数据到mqtt服务事件
	event.On(consts.PushServiceResDataToMQTT, event.ListenerFunc(pushServiceResDataToMQTT), event.High)
	//推送设置属性响应数据到mqtt服务事件
	event.On(consts.PushSetResDataToMQTT, event.ListenerFunc(pushSetResDataToMQTT), event.High)

	// 服务下发获取网关配置信息事件
	event.On(GetGatewayVersionEvent, event.ListenerFunc(getGatewayVersionData), event.Normal)

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
			case int, int16, int32, int64,
				uint, uint16, uint32, uint64,
				float32, float64,
				bool, string:
				param.Value = v
			case mqttProtocol.PropertyNode:
				param.Value = v.(mqttProtocol.PropertyNode).Value
				param.CreateTime = v.(mqttProtocol.PropertyNode).CreateTime
			default:
				param.Value = gconv.Map(v)
			}
			if param.CreateTime == 0 {
				param.CreateTime = gtime.Timestamp()
			}

			propertieData[k] = param
		}
	}

	//eventsData 事件信息
	var eventsData = make(map[string]mqttProtocol.EventNode)
	if e.Data()["EventDataList"] != nil {
		eventDataList := gconv.Map(e.Data()["EventDataList"])
		for k, v := range eventDataList {
			var param = mqttProtocol.EventNode{}
			switch v.(type) {
			case mqttProtocol.EventNode:
				param.Value = v.(mqttProtocol.EventNode).Value
				param.CreateTime = v.(mqttProtocol.PropertyNode).CreateTime
			default:
				param.Value = gconv.Map(v)
			}
			if param.CreateTime == 0 {
				param.CreateTime = gtime.Timestamp()
			}
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
	builder.SetMethod("thing.event.property.pack.post")
	builder.Build()
	data := gconv.Map(builder.Build())
	outData := gjson.New(data).MustToJson()
	log.Debug("设备Key：%v，推送【属性数据】到MQTT服务：%s", deviceKey, outData)
	if err = mqttClient.PublishData(deviceKey, outData); err != nil {
		log.Debug("pushAttributeDataToMQTT", err.Error())
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
		log.Debug("==5555==监听回复信息====", msg)
		mqData := mqttProtocol.ServiceCallOutputRes{}
		mqData.Id = msg.MessageID
		mqData.Code = 200
		mqData.Message = "success"
		mqData.Version = "1.0"
		mqData.Data = replyDataMap

		//推送数据到mqtt
		topic := msg.Topic + "_reply"

		log.Debug("设备Key：%v，推送【服务调用应答数据】到MQTT服务：%v", deviceKey, mqData)
		outData, err := json.Marshal(mqData)
		if err != nil {
			log.Debug("服务回调响应序列化失败：", err.Error())
			return err
		}

		log.Debug("服务回调响应：", mqData)
		log.Debug("服务回调响应topic：", topic)
		err = mqttClient.Publish(topic, outData)
		if err != nil {
			log.Debug("发送服务回调响应失败：", err.Error())
		} else {
			log.Debug("发送服务回调响应成功")
		}
	}
	return
}

// pushSetResDataToMQTT 推送属性设置响应数据到mqtt服务
func pushSetResDataToMQTT(e event.Event) (err error) {
	deviceKey := gconv.String(e.Data()["DeviceKey"])
	replyData := e.Data()["ReplyData"]
	glog.Debugf(context.Background(), "【IotGateway】推送属性设置响应数据到mqtt服务：设备：%s,数据：%v", deviceKey, replyData)
	replyDataMap := make(map[string]interface{})
	if replyData != nil {
		replyDataMap = gconv.Map(replyData)
	}

	msg, err := vars.GetUpMessageMap(deviceKey)
	if msg.MessageID != "" && err == nil {
		glog.Debug(context.Background(), "【IotGateway】 监听回复信息", msg)
		mqData := mqttProtocol.ServiceCallOutputRes{}
		mqData.Id = msg.MessageID
		mqData.Code = 200
		mqData.Message = "success"
		mqData.Version = "1.0"
		mqData.Data = replyDataMap

		//推送数据到mqtt
		topic := msg.Topic + "_reply"
		outData, err := json.Marshal(mqData)
		if err != nil {
			glog.Debugf(context.Background(), "【IotGateway】属性设置响应序列化失败：%v", err.Error())
			return err
		}
		glog.Debugf(context.Background(), "【IotGateway】向平推送属性设置应答数据Topic:%s", topic)
		glog.Debugf(context.Background(), "【IotGateway】设备Key：%v，推送【属性设置应答数据】到MQTT服务：%v", deviceKey, string(outData))

		err = mqttClient.Publish(topic, outData)
		if err != nil {
			log.Debug("【IotGateway】向mqtt服务推送属性设置响应失败：", err.Error())
		} else {
			log.Debug("【IotGateway】向mqtt服务推送属性设置响应成功")
		}
		vars.DeleteFromUpMessageMap(deviceKey)
	}
	return
}

// getGatewayVersionData 获取网关版本信息事件
func getGatewayVersionData(e event.Event) (err error) {
	// 获取设备KEY
	ok, deviceKey := lib.GetMapValueForKey(e.Data(), "DeviceKey")
	if !ok {
		glog.Debug(context.Background(), "获取设备KEY失败")
		return fmt.Errorf("获取设备KEY失败: %s", e.Data())
	}
	//==== 平台端下发调用 应答====
	ra, err := vars.GetUpMessageMap(deviceKey.(string))
	if err == nil {
		if ra.MessageID != "" {

			var rd = make(map[string]interface{})
			rd["Version"] = version.GetVersion()
			rd["BuildTime"] = version.GetBuildTime()
			rd["CommitID"] = version.CommitID

			outData := g.Map{
				"DeviceKey": deviceKey,
				"ReplyData": rd,
			}
			event.Async(consts.PushServiceResDataToMQTT, outData)
		}
	}
	return
}
