package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/caarlos0/env/v6"
)

type envConfig struct {
	Port int `env:"PORT,required"`
}

var (
	instance *envConfig
	once     sync.Once
	initErr  error
)

func LoadConfig() (*envConfig, error) {
	once.Do(func() {
		tmpConfig := &envConfig{}
		if err := env.Parse(tmpConfig); err != nil {
			initErr = fmt.Errorf("failed to load config: %w", err)
			return
		}
		log.Println("config loaded successfully")
		instance = tmpConfig
	})
	
	return instance, initErr
}
