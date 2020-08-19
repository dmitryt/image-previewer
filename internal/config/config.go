package config

import (
	"context"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/file"
)

type Config struct {
	Host        string `yaml:"host" config:"required"`
	Port        int    `yaml:"port" config:"required"`
	CacheDir    string `yaml:"cacheDir" config:"required"`
	CacheSize   int    `yaml:"cacheSize" config:"required"`
	LogLevel    string `yaml:"logLevel"`
	MaxFileSize int64  `yaml:"maxFileSize" config:"required"`
}

func GetDefaultConfig() *Config {
	return &Config{
		Host:        "0.0.0.0",
		Port:        8082,
		LogLevel:    "debug",
		CacheDir:    ".cache",
		CacheSize:   10,
		MaxFileSize: 5 * 1024 * 1024,
	}
}

func Read(fpath string) (config *Config, err error) {
	config = GetDefaultConfig()
	var loader *confita.Loader
	if fpath == "" {
		loader = confita.NewLoader(
			env.NewBackend(),
		)
	} else {
		loader = confita.NewLoader(
			file.NewBackend(fpath),
			env.NewBackend(),
		)
	}
	err = loader.Load(context.Background(), config)

	return
}
