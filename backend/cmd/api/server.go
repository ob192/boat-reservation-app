package main

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log/slog"
	"net/http"
	"os"
	_ "strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/harbour-wave/harbour-wave-backend/internal/config"
	"github.com/harbour-wave/harbour-wave-backend/internal/handler"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"

	_ "github.com/harbour-wave/harbour-wave-backend/docs"
)

// Server acts as our dependency container, holding everything main() needs to start up.
type Server struct {
	Config     *config.Config
	Log        *slog.Logger
	DB         *gorm.DB
	Router     *gin.Engine
	SystemRepo repository.SystemRepository
	BookingSvc service.BookingService
}

// ProvideLogger maps the config's GinMode to a debug boolean for the logger.
func ProvideLogger(cfg *config.Config) *slog.Logger {
	return platform.NewLogger(cfg.GinMode == gin.DebugMode)
}

// ProvideDB maps the config to the platform.NewDB signature.
func ProvideDB(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	return platform.NewDB(cfg.DatabaseURL, cfg.GinMode == gin.DebugMode)
}

// Drop this into cmd/api/server.go, replacing the existing ProvidePaymentGateway func.

func ProvidePaymentGateway(cfg *config.Config, log *slog.Logger) provider.PaymentGateway {
	switch cfg.PaymentGateway {
	case "liqpay":
		if cfg.LiqpayPublicKey == "" || cfg.LiqpayPrivateKey == "" {
			log.Error("liqpay selected but LIQPAY_PUBLIC_KEY / LIQPAY_PRIVATE_KEY are not set")
			os.Exit(1)
		}
		log.Info("payment gateway: liqpay")
		return provider.NewLiqPayGateway(cfg.LiqpayPublicKey, cfg.LiqpayPrivateKey, cfg.BackendURL)
	case "stub", "":
		log.Info("payment gateway: stub (development)")
		return provider.NewStubGateway(cfg.BackendURL)
	default:
		log.Warn("unknown PAYMENT_GATEWAY value, falling back to stub", "value", cfg.PaymentGateway)
		return provider.NewStubGateway(cfg.BackendURL)
	}
}

// ProvideRouter isolates the Gin wiring from main().
func ProvideRouter(
	cfg *config.Config,
	log *slog.Logger,
	authSvc service.AuthService,
	availabilityH *handler.AvailabilityHandler,
	bookingH *handler.BookingHandler,
	checkoutH *handler.CheckoutHandler,
	webhookH *handler.WebhookHandler,
	adminH *handler.AdminHandler,
) *gin.Engine {
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger(log))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.POST("/payment/webhook", webhookH.Handle)

	protected := r.Group("/", handler.AuthMiddleware(authSvc))
	{
		protected.GET("/availability/:month", availabilityH.GetAvailability)
		protected.GET("/slots/:date", availabilityH.GetSlots)
		protected.POST("/bookings", bookingH.CreateBooking)
		protected.POST("/checkout", checkoutH.CreateCheckout)
		protected.GET("/bookings/by-session/:sessionId", bookingH.GetBySession)
	}

	admin := r.Group("/admin", handler.AuthMiddleware(authSvc), handler.AdminMiddleware())
	{
		admin.GET("/slots/:date/:time/bookings", adminH.GetSlotBookings)
		admin.PUT("/slots/:date/:time", adminH.UpsertSlot)
		admin.PATCH("/bookings/:bookingId/price", adminH.OverridePrice)
		admin.PUT("/slots/:date/:time/block", adminH.BlockSlot)
		admin.DELETE("/slots/:date/:time/block", adminH.UnblockSlot)
		admin.PUT("/dates/:date/block", adminH.BlockDate)
		admin.DELETE("/dates/:date/block", adminH.UnblockDate)
		admin.PUT("/system/bookings-enabled", adminH.SetBookingsEnabled)
		admin.POST("/bookings/:bookingId/cancel", adminH.CancelBooking)
	}

	return r
}

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("http",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
