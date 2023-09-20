package lib

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"strings"
)

// ChineseToHex 将中文转为GBK编码的十六进制
func ChineseToHex(src string) (string, error) {
	// 创建一个将文本转换为GBK编码的转换器
	t := simplifiedchinese.GBK.NewEncoder()
	reader := transform.NewReader(strings.NewReader(src), t)

	// 读取所有经过转换的内容
	res, err := io.ReadAll(reader)
	if err != nil {
		return "", errors.New("read transformation result failed: " + err.Error())
	}

	// 将GBK编码转换为十六进制
	hexStr := hex.EncodeToString(res)

	return hexStr, nil
}

// GetBytesByInt 按两个字节进行取值转为数组
func GetBytesByInt(data []byte) []uint16 {
	// 创建一个 uint16 切片来保存16位的值
	values := make([]uint16, len(data)/2)
	// 每两个字节创建一个16位的值，存入values切片
	for i := 0; i < len(data); i += 2 {
		values[i/2] = binary.BigEndian.Uint16(data[i : i+2])
	}
	return values
}

// HexToBytes 将十六进制字符串转为[]byte
func HexToBytes(hexData string) ([]byte, error) {
	// 如果有0x前缀,去掉
	if strings.HasPrefix(hexData, "0x") {
		hexData = hexData[2:]
	}
	// 使用encoding/hex包解码
	return hex.DecodeString(hexData)
}

// GetTopicInfo 通过topic获取设备KEY  3,deviceKey
func GetTopicInfo(valueName, topic string) string {
	splits := strings.Split(topic, "/")
	if len(splits) >= 3 {
		switch valueName {
		case "deviceKey":
			return splits[3]
		case "productKey":
			return splits[2]
		}
	}
	return ""
}

// RandString 生成随机字符串
func RandString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
