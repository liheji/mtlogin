package ntfy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Ntfy Ntfy结构体
type Ntfy struct {
	URL   string
	Topic string
	Auth  string
}

// TextMessage 文本消息
type TextMessage struct {
	Topic    string    `json:"topic"`
	Title    string    `json:"title"`
	Message  string    `json:"message"`
	Priority int       `json:"priority"`
	Tags     []string  `json:"tags,omitempty"`
	Actions  []Actions `json:"actions,omitempty"`
	Attach   string    `json:"attach,omitempty"`
	Filename string    `json:"filename,omitempty"`
	Click    string    `json:"click,omitempty"`
}

type Actions struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	URL    string `json:"url"`
}

// NewNtfy 创建Ntfy实例
func NewNtfy(url, topic, auth string) *Ntfy {
	return &Ntfy{URL: url, Topic: topic, Auth: auth}
}

// SendText 发送文本消息
func (n *Ntfy) SendText(text string) error {
	message := TextMessage{
		Topic:    n.Topic,
		Title:    "M-Team 登录保活通知",
		Message:  text,
		Priority: 1,
		Tags:     []string{"bread"},
		Actions: []Actions{
			{
				Action: "view",
				Label:  "手动登录",
				URL:    "https://kp.m-team.cc/",
			},
		},
		Click: "https://kp.m-team.cc/",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	return n.sendRequest(jsonData)
}

// sendRequest 发送HTTP请求到 Ntfy
func (n *Ntfy) sendRequest(data []byte) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", n.URL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 如果设置了密钥，认证
	if n.Auth != "" {
		req.Header.Set("Authorization", n.Auth)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %v", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API返回错误: %s", string(body))
	}

	fmt.Println("消息发送成功:", string(body))
	return nil
}
