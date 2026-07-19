package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"

	"retailpulse/apps/api/internal/config"
	"retailpulse/apps/api/internal/controller"
	"retailpulse/apps/api/internal/middleware"
	"retailpulse/apps/api/internal/platform/database"
	"retailpulse/apps/api/internal/platform/security"
	"retailpulse/apps/api/internal/repository"
	"retailpulse/apps/api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration failed", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, "migrations"); err != nil {
		logger.Error("database migrations failed", "error", err)
		os.Exit(1)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()

	passwords := security.NewPasswordHasher()
	tokens := security.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
	encryptor, err := security.NewEncryptor(cfg.TokenEncryptionKey)
	if err != nil {
		logger.Error("token encryption configuration failed", "error", err)
		os.Exit(1)
	}
	authRepo := repository.NewPostgresAuthRepository(pool)
	auditRepo := repository.NewPostgresAuditRepository(pool)
	amazonRepo := repository.NewPostgresAmazonRepository(pool)
	analyticsRepo := repository.NewAnalyticsRepository(pool)
	commerceRepo := repository.NewCommerceRepository(pool)
	authService := service.NewAuthService(authRepo, auditRepo, passwords, tokens, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	amazonService := service.NewAmazonService(amazonRepo, auditRepo, encryptor, service.AmazonConfig{
		ClientID:         cfg.AmazonLWAClientID,
		ClientSecret:     cfg.AmazonLWASecret,
		ApplicationID:    cfg.AmazonAppID,
		AuthVersion:      cfg.AmazonAuthVersion,
		RedirectURL:      cfg.AmazonRedirectURL,
		SellerCentralURL: cfg.AmazonSellerCentralURL,
		SPAPIEndpoint:    cfg.AmazonSPAPIEndpoint,
		AWSAccessKey:     cfg.AmazonAWSAccessKey,
		AWSSecretKey:     cfg.AmazonAWSSecretKey,
		AWSSessionToken:  cfg.AmazonAWSSessionToken,
		AWSRegion:        cfg.AmazonAWSRegion,
		SandboxClientID:  cfg.AmazonSandboxClientID,
		SandboxSecret:    cfg.AmazonSandboxSecret,
		SandboxToken:     cfg.AmazonSandboxToken,
		SandboxEndpoint:  cfg.AmazonSandboxEndpoint,
		AdsClientID:      cfg.AmazonAdsClientID,
		AdsClientSecret:  cfg.AmazonAdsClientSecret,
		AdsRefreshToken:  cfg.AmazonAdsRefreshToken,
		AdsProfileID:     cfg.AmazonAdsProfileID,
		AdsEndpoint:      cfg.AmazonAdsEndpoint,
	})

	app := fiber.New(fiber.Config{
		AppName:      "RetailPulse AI API",
		ErrorHandler: middleware.ErrorHandler(logger),
	})
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Organization-Id",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        120,
		Expiration: time.Minute,
	}))

	controller.RegisterHealthRoutes(app)
	controller.RegisterAuthRoutes(app.Group("/v1/auth"), authService)
	controller.RegisterAmazonRoutes(app.Group("/v1/amazon", middleware.Authenticate(tokens)), amazonService, cfg.WebAppURL)
	controller.RegisterAmazonOAuthCallback(app.Group("/v1/amazon/oauth"), amazonService, cfg.WebAppURL)
	controller.RegisterAnalyticsRoutes(app.Group("/v1/analytics", middleware.Authenticate(tokens)), analyticsRepo)
	controller.RegisterCommerceRoutes(app.Group("/v1/commerce", middleware.Authenticate(tokens)), commerceRepo)
	controller.RegisterStorefrontRoutes(app.Group("/v1/storefront"), commerceRepo)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			logger.Error("api stopped", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
