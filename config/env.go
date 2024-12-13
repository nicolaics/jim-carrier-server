package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PublicHost                       string
	Port                             string
	DBUser                           string
	DBPassword                       string
	DBAddress                        string
	DBName                           string
	JWTAccessExpInSeconds            int64
	JWTAccessSecret                  string
	JWTRefreshExpInSeconds           int64
	JWTRefreshSecret                 string
	CompanyEmail                     string
	CompanyEmailPassword             string
	GoogleClientID                   string
	GoogleClientSecret               string
	GoogleApplicationCredentialsPath string
	BucketName                       string
	AWSRegion                        string
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	return Config{
		PublicHost: getEnv("PUBLIC_HOST", "http://localhost"),
		Port:       getEnv("PORT", "19230"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBAddress: fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"),
			getEnv("DB_PORT", "3306")),
		DBName:                           getEnv("DB_NAME", "jim_carrier"),
		JWTAccessExpInSeconds:            getEnvAsInt("JWT_ACCESS_EXP", (3600 * 6)), // for 6 hours
		JWTAccessSecret:                  getEnv("JWT_ACCESS_SECRET", "access-secret"),
		JWTRefreshExpInSeconds:           getEnvAsInt("JWT_REFRESH_EXP", (3600 * 24 * 7)), // for 1 week
		JWTRefreshSecret:                 getEnv("JWT_REFRESH_SECRET", "access-secret"),
		CompanyEmail:                     getEnv("COMPANY_EMAIL", "abc@gmail.com"),
		CompanyEmailPassword:             getEnv("COMPANY_EMAIL_PASSWORD", "1234"),
		GoogleClientID:                   getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:               getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleApplicationCredentialsPath: getEnv("GOOGLE_APPLICATION_CREDENTIALS_PATH", ""),
		BucketName:                       getEnv("BUCKET_NAME", "jim"),
		AWSRegion:                        getEnv("AWS_REGION", "us"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			return int64(fallback)
		}

		return i
	}

	return fallback
}
