package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	LLM LLMConfig `json:"llm"`
}

type LLMConfig struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
}

func Load() (*Config, error) {
	configPath := getConfigPath()

	// 确保全局配置目录存在
	if !filepath.IsAbs(configPath) || !strings.Contains(configPath, ".gac_config.json") {
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建配置目录失败: %w", err)
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if config.LLM.APIKey == "" {
		config.LLM.APIKey = os.Getenv("GAC_API_KEY")
	}

	return &config, nil
}

func getConfigPath() string {
	// 首先检查项目级配置
	projectConfig := filepath.Join(".", ".gac_config.json")
	if _, err := os.Stat(projectConfig); err == nil {
		return projectConfig
	}

	// 然后检查全局配置
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return projectConfig // fallback
	}

	return filepath.Join(homeDir, ".gac", "config.json")
}

func getDefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider: "openai",
			Endpoint: "https://api.openai.com/v1/chat/completions",
			Model:    "gpt-3.5-turbo",
			APIKey:   os.Getenv("GAC_API_KEY"),
		},
	}
}
