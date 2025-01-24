package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort            string
	DbHost             string
	DbPort             string
	DbUser             string
	DbPass             string
	DbName             string
	DbSSL              string
	S3Bucket           string
	S3Endpoint         string
	AwsRegion          string
	AwsAccessKeyId     string
	AwsSecretAccessKey string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		DbHost:  getEnv("DB_HOST", "localhost"),
		DbPort:  getEnv("DB_PORT", "5432"),
		DbUser:  getEnv("DB_USER", "postgres"),
		DbPass:  getEnv("DB_PASSWORD", "password"),
		DbName:  getEnv("DB_NAME", "mydb"),
		DbSSL:   getEnv("DB_SSL", "disable"),

		S3Bucket:           getEnv("AWS_S3_BUCKET_NAME", ""),
		S3Endpoint:         getEnv("AWS_S3_ENDPOINT", ""),
		AwsRegion:          getEnv("AWS_REGION", ""),
		AwsAccessKeyId:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AwsSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
