package config

import (
	"log"
	"sync"

	"github.com/caarlos0/env/v6"
)

type envConfig struct {
	// Google configs
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	GoogleCallbackURL  string `env:"GOOGLE_CALLBACK_URL" envDefault:"http://localhost:9753/auth/google/callback"`
	// GitHub configs
	GitHubClientID     string `env:"GITHUB_CLIENT_ID,required"`
	GitHubClientSecret string `env:"GITHUB_CLIENT_SECRET,required"`
	GithubCallbackURL  string `env:"GITHUB_CALLBACK_URL" envDefault:"http://localhost:9753/auth/github/callback"`
	// Database configs
	DatabaseUrl string `env:"DATABASE_URL,required"`
	// Server configs
	Port         int  `env:"PORT,required"`
	IsProduction bool `env:"IS_PRODUCTION" envDefault:"false"`
	// JWTSecret
	JWTSecret string `env:"JWT_SECRET,required"`
	// S3 configs
	AWSAccessKeyID     string `env:"AWS_ACCESS_KEY_ID,required"`
	AWSSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY,required"`
	AWSRegion          string `env:"AWS_REGION,required"`
	S3BucketName       string `env:"S3_BUCKET_NAME,required"`
	S3MaxUploadSizeMB  int    `env:"S3_MAX_UPLOAD_SIZE_MB" envDefault:"5"`
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
