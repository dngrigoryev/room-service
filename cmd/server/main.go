package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"room-service/internal/config"
	"room-service/internal/repository/postgres"
	"room-service/internal/service"
	"room-service/internal/transport/http/handlers"
	"room-service/internal/transport/http/middlewares"
	"room-service/pkg/token"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	log.Println("Starting room booking service...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Running database migrations...")
	m, err := migrate.New("file://migrations", cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to instantiate migrate: %v", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Failed to run migrate up: %v", err)
	}
	log.Println("Database migrations applied successfully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := postgres.NewPool(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Connected to PostgreSQL successfully")

	tokenManager := token.NewManager(cfg.JWTSecret)

	roomRepo := postgres.NewRoomRepo(dbPool)
	scheduleRepo := postgres.NewScheduleRepo(dbPool)
	slotRepo := postgres.NewSlotRepo(dbPool)
	bookingRepo := postgres.NewBookingRepo(dbPool)
	userRepo := postgres.NewUserRepo(dbPool)

	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(scheduleRepo)
	slotService := service.NewSlotService(slotRepo, scheduleRepo, roomRepo)
	bookingService := service.NewBookingService(bookingRepo, slotRepo)
	authService := service.NewAuthService(userRepo, tokenManager)

	authHandler := handlers.NewAuthHandler(tokenManager, authService)
	roomHandler := handlers.NewRoomHandler(roomService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService)
	slotHandler := handlers.NewSlotHandler(slotService)
	bookingHandler := handlers.NewBookingHandler(bookingService)

	router := SetupRouter(tokenManager, authHandler, roomHandler, scheduleHandler, slotHandler, bookingHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	go func() {
		log.Printf("Listening on port %s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func SetupRouter(
	tokenManager *token.Manager,
	authHandler *handlers.AuthHandler,
	roomHandler *handlers.RoomHandler,
	scheduleHandler *handlers.ScheduleHandler,
	slotHandler *handlers.SlotHandler,
	bookingHandler *handlers.BookingHandler) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/_info", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.POST("/dummyLogin", authHandler.DummyLogin)
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	api := router.Group("/")
	api.Use(middlewares.AuthMiddleware(tokenManager))

	api.GET("/rooms/list", roomHandler.List)

	api.GET("/rooms/:roomId/slots/list", slotHandler.List)

	adminArea := api.Group("/")
	adminArea.Use(middlewares.RequireRole("admin"))
	{
		adminArea.POST("/rooms/create", roomHandler.Create)
		adminArea.POST("/rooms/:roomId/schedule/create", scheduleHandler.Create)

		adminArea.GET("/bookings/list", bookingHandler.List)
	}

	userArea := api.Group("/")
	userArea.Use(middlewares.RequireRole("user"))
	{

		userArea.POST("/bookings/create", bookingHandler.Create)
		userArea.GET("/bookings/my", bookingHandler.My)
		userArea.POST("/bookings/:bookingId/cancel", bookingHandler.Cancel)
	}

	return router
}
