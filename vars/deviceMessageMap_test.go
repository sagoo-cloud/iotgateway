package vars

import (
	"github.com/sagoo-cloud/iotgateway/model"
	"testing"
)

func TestUpdateUpMessageMap(t *testing.T) {

	var device = model.UpMessage{
		MessageID:  "112233",
		SendTime:   1,
		MethodName: "testMethodName",
		Topic:      "testTopic",
	}
	UpdateUpMessageMap("test123", device)
	gotRes, err := GetUpMessageMap("test123")
	if err != nil {
		t.Errorf("GetUpMessageMap() error = %v", err)
		return
	}
	t.Log("获取存入的数据：", gotRes)
}
