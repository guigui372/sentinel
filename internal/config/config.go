package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Global struct {
	Timeout  int    `yaml:"timeout"`
	Webhook  string `yaml:"webhook"`
	Interval int    `yaml:"interval"`
}

// Server 结构体：对应 yaml 里的 servers
type Server struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	KeyPath  string `yaml:"key_path"`
	Password string `yaml:"password"`
}

// Task 结构体：对应 yaml 里的 tasks
type Task struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// Config 结构体：总入口
type Config struct {
	Global  Global   `yaml:"global"`
	Servers []Server `yaml:"servers"`
	Tasks   []Task   `yaml:"tasks"`
}

// LoadConfig 读取并解析配置文件
func LoadConfig(path string) (*Config, error) {
	// 1. 读取文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 2. 解析 YAML 到结构体
	conf := &Config{}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
