package config

import (
	"os"
	"strconv"
)

// Config holds application configuration (env or file).
type Config struct {
	ServerPort   int
	PostgresDSN  string
	MongoURI     string
	MongoDBName  string
	JWTSecret    string
	LogLevel     string
	MigratePath  string
	SmsSecretID  string
	SmsSecretKey string
	SmsSdkAppID  string
	SmsSignName  string
	SmsRegion    string
	// TheoryVideoStorage：gridfs（默认）或 cos（未实现时在 main 中会失败）
	TheoryVideoStorage string
}

// Load reads configuration from environment.
func Load() *Config {
	port, _ := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	return &Config{
		ServerPort:         port,
		PostgresDSN:        getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/maguafly?sslmode=disable"),
		MongoURI:           getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBName:        getEnv("MONGODB_DB", "maguafly"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		MigratePath:        getEnv("MIGRATE_PATH", "file://migrations"),
		SmsSecretID:        getEnv("TENCENT_SMS_SECRET_ID", ""),
		SmsSecretKey:       getEnv("TENCENT_SMS_SECRET_KEY", ""),
		SmsSdkAppID:        getEnv("TENCENT_SMS_SDK_APP_ID", ""),
		SmsSignName:        getEnv("TENCENT_SMS_SIGN_NAME", ""),
		SmsRegion:          getEnv("TENCENT_SMS_REGION", "ap-guangzhou"),
		TheoryVideoStorage: getEnv("THEORY_VIDEO_STORAGE", "gridfs"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
