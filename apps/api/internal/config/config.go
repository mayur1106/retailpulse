package config

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port                   string
	DatabaseURL            string
	RedisAddr              string
	AllowedOrigins         string
	WebAppURL              string
	JWTAccessSecret        string
	JWTRefreshSecret       string
	TokenEncryptionKey     string
	AmazonLWAClientID      string
	AmazonLWASecret        string
	AmazonAppID            string
	AmazonAuthVersion      string
	AmazonRedirectURL      string
	AmazonSellerCentralURL string
	AmazonSPAPIEndpoint    string
	AmazonAWSAccessKey     string
	AmazonAWSSecretKey     string
	AmazonAWSSessionToken  string
	AmazonAWSRegion        string
	AmazonSandboxClientID  string
	AmazonSandboxSecret    string
	AmazonSandboxToken     string
	AmazonSandboxEndpoint  string
	AmazonAdsClientID      string
	AmazonAdsClientSecret  string
	AmazonAdsRefreshToken  string
	AmazonAdsProfileID     string
	AmazonAdsEndpoint      string
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port:                   getEnv("API_PORT", "4005"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		RedisAddr:              getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		AllowedOrigins:         getEnv("ALLOWED_ORIGINS", "*"),
		WebAppURL:              getEnv("WEB_APP_URL", "http://localhost:3005"),
		JWTAccessSecret:        os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret:       os.Getenv("JWT_REFRESH_SECRET"),
		TokenEncryptionKey:     os.Getenv("TOKEN_ENCRYPTION_KEY"),
		AmazonLWAClientID:      os.Getenv("AMAZON_LWA_CLIENT_ID"),
		AmazonLWASecret:        os.Getenv("AMAZON_LWA_CLIENT_SECRET"),
		AmazonAppID:            os.Getenv("AMAZON_SPAPI_APP_ID"),
		AmazonAuthVersion:      getEnv("AMAZON_SPAPI_AUTH_VERSION", ""),
		AmazonRedirectURL:      getEnv("AMAZON_LWA_REDIRECT_URL", "http://localhost:4005/v1/amazon/oauth/callback"),
		AmazonSellerCentralURL: getEnv("AMAZON_SELLER_CENTRAL_URL", "https://sellercentral.amazon.com"),
		AmazonSPAPIEndpoint:    getEnv("AMAZON_SPAPI_ENDPOINT", "https://sellingpartnerapi-na.amazon.com"),
		AmazonAWSAccessKey:     os.Getenv("AMAZON_AWS_ACCESS_KEY_ID"),
		AmazonAWSSecretKey:     os.Getenv("AMAZON_AWS_SECRET_ACCESS_KEY"),
		AmazonAWSSessionToken:  os.Getenv("AMAZON_AWS_SESSION_TOKEN"),
		AmazonAWSRegion:        getEnv("AMAZON_AWS_REGION", "us-east-1"),
		AmazonSandboxClientID:  os.Getenv("AMAZON_SANDBOX_LWA_CLIENT_ID"),
		AmazonSandboxSecret:    os.Getenv("AMAZON_SANDBOX_LWA_CLIENT_SECRET"),
		AmazonSandboxToken:     os.Getenv("AMAZON_SANDBOX_REFRESH_TOKEN"),
		AmazonSandboxEndpoint:  getEnv("AMAZON_SANDBOX_SPAPI_ENDPOINT", "https://sandbox.sellingpartnerapi-na.amazon.com"),
		AmazonAdsClientID:      os.Getenv("AMAZON_ADS_CLIENT_ID"),
		AmazonAdsClientSecret:  os.Getenv("AMAZON_ADS_CLIENT_SECRET"),
		AmazonAdsRefreshToken:  os.Getenv("AMAZON_ADS_REFRESH_TOKEN"),
		AmazonAdsProfileID:     os.Getenv("AMAZON_ADS_PROFILE_ID"),
		AmazonAdsEndpoint:      getEnv("AMAZON_ADS_ENDPOINT", "https://advertising-api.amazon.com"),
		AccessTokenTTL:         15 * time.Minute,
		RefreshTokenTTL:        30 * 24 * time.Hour,
	}

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if len(cfg.JWTAccessSecret) < 32 {
		missing = append(missing, "JWT_ACCESS_SECRET")
	}
	if len(cfg.JWTRefreshSecret) < 32 {
		missing = append(missing, "JWT_REFRESH_SECRET")
	}
	if cfg.TokenEncryptionKey != "" {
		if _, err := securityKeyLength(cfg.TokenEncryptionKey); err != nil {
			missing = append(missing, "TOKEN_ENCRYPTION_KEY")
		}
	}
	if len(missing) > 0 {
		return Config{}, errors.New("missing or invalid configuration: " + strings.Join(missing, ", "))
	}
	return cfg, nil
}

func securityKeyLength(value string) (int, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return 0, err
	}
	if len(decoded) != 32 {
		return len(decoded), errors.New("invalid key length")
	}
	return len(decoded), nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
