package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SendDingTalk 发送钉钉消息
// webhook: 钉钉机器人的地址
// content: 要发送的文字内容
func SendDingTalk(webhook string, content string) error {
	// 1. 如果没有配置 webhook，直接跳过，不算错
	if webhook == "" {
		return nil
	}

	// 2. 准备数据包 (这是钉钉规定的格式，必须这么写)
	// 我们要把数据定义成一个 map (字典)，然后转成 JSON
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": "🚨 [巡检告警] \n" + content,
		},
	}

	// 把 map 转成 JSON 字节 (序列化)
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 3. 发送 HTTP POST 请求
	// 设置 3 秒超时，别因为发警报卡住主程序
	client := &http.Client{Timeout: 3 * time.Second}

	// 创建请求
	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// 执行发送
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 4. 检查对方是不是收到了 (状态码 200)
	if resp.StatusCode != 200 {
		return fmt.Errorf("钉钉返回异常状态码: %d", resp.StatusCode)
	}

	return nil
}
