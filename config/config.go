package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Channels []ChannelConfig `yaml:"channels"`
	Server   ServerConfig    `yaml:"server"`
	Polling  PollingConfig   `yaml:"polling"`
	Storage  StorageConfig   `yaml:"storage"`
	Podcast  PodcastConfig   `yaml:"podcast"`
}

type ChannelConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}

type PollingConfig struct {
	Interval string `yaml:"interval"`
}

type StorageConfig struct {
	AudioDir string `yaml:"audio_dir"`
	FeedDir  string `yaml:"feed_dir"`
}

type PodcastConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Email       string `yaml:"email"`
	Language    string `yaml:"language"`
	Category    string `yaml:"category"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Polling.Interval == "" {
		cfg.Polling.Interval = "30m"
	}
	if cfg.Storage.AudioDir == "" {
		cfg.Storage.AudioDir = "./audio"
	}
	if cfg.Storage.FeedDir == "" {
		cfg.Storage.FeedDir = "./feeds"
	}

	return &cfg, nil
}

func (p *PollingConfig) Duration() (time.Duration, error) {
	return time.ParseDuration(p.Interval)
}
