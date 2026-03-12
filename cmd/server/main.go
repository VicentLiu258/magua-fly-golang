package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"magua-fly-golang/internal/config"
	"magua-fly-golang/internal/handler"
	"magua-fly-golang/internal/mongodb"
	"magua-fly-golang/internal/postgres"
	"magua-fly-golang/internal/pkg/jwt"
	"magua-fly-golang/internal/usecase"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pgDB, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pgDB.Close()

	if err := postgres.RunMigrations(cfg.PostgresDSN, cfg.MigratePath); err != nil {
		log.Printf("migrate (non-fatal): %v", err)
	}

	mongoClient, err := mongodb.New(ctx, cfg.MongoURI, cfg.MongoDBName)
	if err != nil {
		log.Fatalf("mongodb: %v", err)
	}
	defer mongoClient.Close(ctx)

	jwtManager := jwt.NewManager(cfg.JWTSecret, 24*time.Hour)

	healthHandler := handler.NewHealthHandler(pgDB, mongoClient)
	authHandler := handler.NewAuthHandler(usecase.NewAuthUseCase(pgDB.User(), jwtManager))
	studentHandler := handler.NewStudentHandler(pgDB.Student())
	instructorHandler := handler.NewInstructorHandler(pgDB.Instructor())
	aircraftHandler := handler.NewAircraftHandler(pgDB.Aircraft())
	bookingHandler := handler.NewBookingHandler(pgDB.Booking(), pgDB.TrainingPlan())

	router := handler.Router(
		healthHandler,
		authHandler,
		studentHandler,
		instructorHandler,
		aircraftHandler,
		bookingHandler,
		jwtManager,
	)

	portStr := strconv.Itoa(cfg.ServerPort)
	if portStr == "0" {
		portStr = "8080"
	}
	srv := &http.Server{
		Addr:    ":" + portStr,
		Handler: router,
	}
	go func() {
		log.Printf("server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

