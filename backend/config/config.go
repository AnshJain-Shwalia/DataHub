package config

import (
	"log"
	"sync"

	"github.com/caarlos0/env/v6"
)

type envConfig struct {
	Port               int    `env:"PORT,required"`
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	DatabaseUrl        string `env:"DATABASE_URL,required"`
	IsProduction       bool   `env:"IS_PRODUCTION" envDefault:"false"`
}

var (
	instance *envConfig
	once     sync.Once
	initErr  error
)

func LoadConfig() *envConfig {
	once.Do(func() {
		tmpConfig := &envConfig{}
		if err := env.Parse(tmpConfig); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		log.Println("config loaded successfully")
		instance = tmpConfig
	})
	return instance
}
