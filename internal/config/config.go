package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                     string
	Env                      string
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPassword               string
	DBName                   string
	DBSSLMode                string
	DatabaseURL              string
	DBMaxOpenConns           int
	DBMaxIdleConns           int
	DBConnMaxLifetimeMinutes int
	MQTTBrokerURL            string
	MQTTClientID             string
	MQTTKeepAliveSeconds     int
	MQTTCommandTopicPrefix   string
	CommandChannelBufferSize int
}

func LoadConfig() *Config {
	// Attempt to load .env file; continue without error if absent (e.g., in Docker container)
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found or unable to read; reading from system environment variables")
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "greenhouse")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	// Construct DSN from discrete variables, or fallback to DATABASE_URL if explicitly provided
	defaultDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)
	databaseURL := getEnv("DATABASE_URL", defaultDSN)

	return &Config{
		Port:                     getEnv("PORT", "8080"),
		Env:                      getEnv("ENV", "development"),
		DBHost:                   dbHost,
		DBPort:                   dbPort,
		DBUser:                   dbUser,
		DBPassword:               dbPassword,
		DBName:                   dbName,
		DBSSLMode:                dbSSLMode,
		DatabaseURL:              databaseURL,
		DBMaxOpenConns:           getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:           getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetimeMinutes: getEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 5),
		MQTTBrokerURL:            getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
		MQTTClientID:             getEnv("MQTT_CLIENT_ID", "greenhouse-backend"),
		MQTTKeepAliveSeconds:     getEnvAsInt("MQTT_KEEP_ALIVE_SECONDS", 60),
		MQTTCommandTopicPrefix:   getEnv("MQTT_COMMAND_TOPIC_PREFIX", "greenhouse/control"),
		CommandChannelBufferSize: getEnvAsInt("COMMAND_CHANNEL_BUFFER_SIZE", 100),
	}
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return fallback
}
