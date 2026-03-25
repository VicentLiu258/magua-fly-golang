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
	"magua-fly-golang/internal/pkg/sms"
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

	smsClient, err := sms.NewTencentClient(cfg)
	if err != nil {
		log.Fatalf("sms client: %v", err)
	}

	authUC := usecase.NewAuthUseCase(
		pgDB.User(),
		pgDB.Student(),
		pgDB.Instructor(),
		pgDB.SmsVerification(),
		jwtManager,
		smsClient,
		"", // TODO: set login template ID via config/env
		"", // TODO: set register template ID via config/env
	)

	healthHandler := handler.NewHealthHandler(pgDB, mongoClient)
	authHandler := handler.NewAuthHandler(authUC)
	smsHandler := handler.NewSmsHandler(authUC)
	trainingUC := usecase.NewTrainingUseCase(
		pgDB.TrainingRecord(),
		pgDB.TrainingArchive(),
		pgDB.RawFlightImport(),
		pgDB.TrainingRecordView(),
		pgDB.SyllabusSubject(),
		pgDB.SyllabusOutline(),
		pgDB.SyllabusSubjectSecondary(),
		pgDB.SyllabusTree(),
		pgDB.TrainingArchiveSubject(),
		pgDB.NextTrainingSubject(),
		pgDB.Task(),
	)
	studentHandler := handler.NewStudentHandler(
		pgDB.Student(),
		pgDB.User(),
		pgDB.Task(),
		pgDB.TrainingArchive(),
		pgDB.TrainingPlan(),
		pgDB.Instructor(),
		pgDB.Aircraft(),
		pgDB.TheoryAttempt(),
		pgDB.Contract(),
		trainingUC,
	)
	instructorHandler := handler.NewInstructorHandler(pgDB.Instructor())
	aircraftHandler := handler.NewAircraftHandler(pgDB.Aircraft(), pgDB.AircraftAvailability(), pgDB.AircraftMaintenance(), pgDB)
	trainingHandler := handler.NewTrainingHandler(trainingUC)

	theoryUC := usecase.NewTheoryUseCase(
		pgDB.TheoryDocument(),
		pgDB.TheoryAttempt(),
		mongoClient.TheoryContent(),
		mongoClient.TheoryQuestionSet(),
		mongoClient.TheoryAttemptDetail(),
		pgDB.TheoryPaper(),
		pgDB.TheoryPaperQuestion(),
	)
	theoryHandler := handler.NewTheoryHandler(theoryUC)

	subjectExamUC := usecase.NewSubjectExamUseCase(
		pgDB.SubjectLearning(),
		pgDB.SubjectAttempt(),
		mongoClient.SubjectQuestionSet(),
		mongoClient.SubjectAttemptDetail(),
	)
	subjectExamHandler := handler.NewSubjectExamHandler(subjectExamUC)

	taskUC := usecase.NewTaskUseCase(pgDB.Task(), pgDB.Student())
	taskHandler := handler.NewTaskHandler(taskUC)

	// 合同文件（Mongo GridFS + PG contracts）
	contractGridFS, err := mongoClient.NewGridFSBucket("contracts_files")
	if err != nil {
		log.Fatalf("gridfs bucket: %v", err)
	}
	contractFilesUC := usecase.NewContractFilesUseCase(
		pgDB.Contract(),
		mongoClient.ContractFilesMeta(),
		contractGridFS,
	)
	contractFilesHandler := handler.NewContractFilesHandler(contractFilesUC)

	scheduleUC := usecase.NewScheduleUseCase(pgDB.Instructor(), pgDB.InstructorSchedule(), pgDB.StaffSchedule(), pgDB.User(), pgDB.Booking(), pgDB.TrainingPlan(), pgDB.AircraftAvailability())
	bookingHandler := handler.NewBookingHandler(pgDB.Booking(), pgDB.TrainingPlan(), pgDB.Task(), scheduleUC)
	scheduleHandler := handler.NewScheduleHandler(scheduleUC)

	router := handler.Router(
		healthHandler,
		authHandler,
		smsHandler,
		studentHandler,
		instructorHandler,
		aircraftHandler,
		bookingHandler,
		scheduleHandler,
		trainingHandler,
		theoryHandler,
		subjectExamHandler,
		taskHandler,
		contractFilesHandler,
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

