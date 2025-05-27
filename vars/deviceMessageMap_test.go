package vars

import (
	"testing"

	"github.com/sagoo-cloud/iotgateway/model"
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

	// 测试复合键功能
	gotResComposite, err := GetUpMessageByCompositeKey("test123", "112233")
	if err != nil {
		t.Errorf("GetUpMessageByCompositeKey() error = %v", err)
		return
	}
	t.Log("通过复合键获取存入的数据：", gotResComposite)

	// 测试缓存统计
	stats := GetCacheStats()
	t.Log("缓存统计信息：", stats)

	// 清理测试数据
	DeleteFromUpMessageMap("test123")
}
