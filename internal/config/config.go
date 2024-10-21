package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
    S3Bucket      string
    ChatGPTAPIKey string
}

func LoadConfig() (Config, error) {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    return Config{
        S3Bucket:      os.Getenv("S3_BUCKET"),
        ChatGPTAPIKey: os.Getenv("CHATGPT_API_KEY"),
    }, nil
}
