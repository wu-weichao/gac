package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Client struct {
	APIKey   string
	Endpoint string
	Model    string
	Client   *http.Client
}

func NewClient(apiKey, endpoint, model string) *Client {
	return &Client{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Model:    model,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GenerateCommitMessage(prompt string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("API 密钥是必需的")
	}

	request := OpenAIRequest{
		Model: c.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   200,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API 请求失败，状态码 %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("反序列化响应失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("响应中没有选项")
	}

	return response.Choices[0].Message.Content, nil
}
