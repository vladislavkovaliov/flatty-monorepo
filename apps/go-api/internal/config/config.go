package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseUrl   string
	KafkaBrockers string
	RABBITMQ_URL  string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseUrl:   getEnv("DATABASE_URL", "postgres://postgres:password@localhost:55000/flatty?sslmode=disable"),
		KafkaBrockers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		RABBITMQ_URL:  getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
