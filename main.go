package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"microservice-alert-service/alert/application/commandservices"
	"microservice-alert-service/alert/application/eventhandlers"
	"microservice-alert-service/alert/application/queryservices"
	"microservice-alert-service/alert/infrastructure/configuration"
	"microservice-alert-service/alert/infrastructure/messaging/kafka"
	"microservice-alert-service/alert/infrastructure/notifications/gmail"
	"microservice-alert-service/alert/infrastructure/notifications/twilio"
	gormconfig "microservice-alert-service/alert/infrastructure/persistence/gorm/configuration"
	"microservice-alert-service/alert/infrastructure/persistence/gorm/model"
	"microservice-alert-service/alert/infrastructure/persistence/gorm/repositories"
	"microservice-alert-service/alert/infrastructure/persistence/memory"
	"microservice-alert-service/alert/interfaces/rest"
	"microservice-alert-service/alert/interfaces/rest/controllers"
)

func main() {
	logger := log.New(os.Stdout, "alert-service ", log.LstdFlags|log.LUTC)

	cfg, err := configuration.Load()
	if err != nil {
		logger.Fatalf("config error: %v", err)
	}
	logger.Printf("kafka brokers resolved: %v", cfg.KafkaBrokers)
	logger.Printf("kafka consumption topics: %v | group: %s", cfg.KafkaConsumptionTopics, cfg.KafkaConsumerGroup)

	db, err := gormconfig.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("database error: %v", err)
	}

	if cfg.AutoMigrate {
		if err := db.AutoMigrate(
			&model.AlertThresholdModel{},
			&model.InactivityRuleModel{},
			&model.AlertModel{},
			&model.NotificationPreferenceModel{},
			&model.NotificationLogModel{},
		); err != nil {
			logger.Fatalf("migration error: %v", err)
		}
	} else {
		logger.Println("auto-migrate disabled by AUTO_MIGRATE=false")
	}

	alertRepo := repositories.NewAlertRepository(db)
	thresholdRepo := repositories.NewAlertThresholdRepository(db)
	inactivityRepo := repositories.NewInactivityRuleRepository(db)
	preferenceRepo := repositories.NewNotificationPreferenceRepository(db)
	logRepo := repositories.NewNotificationLogRepository(db)
	deviceActivityRepo := memory.NewDeviceActivityRepository()

	emailSender := gmail.NewSender(cfg, logger)
	smsSender := twilio.NewSender(cfg, logger)

	defaultEmailTo := cfg.MailFrom
	if defaultEmailTo == "" {
		defaultEmailTo = cfg.MailUsername
	}
	defaultSmsTo := cfg.TwilioPhoneNumber

	notificationService := commandservices.NewNotificationService(
		preferenceRepo,
		logRepo,
		emailSender,
		smsSender,
		defaultEmailTo,
		defaultSmsTo,
		logger,
	)

	kafkaProducer := kafka.NewAlertEventProducer(cfg, logger)
	alertCommandService := commandservices.NewAlertCommandService(alertRepo, kafkaProducer, notificationService, logger)
	thresholdCommandService := commandservices.NewThresholdCommandService(thresholdRepo, logger)
	inactivityCommandService := commandservices.NewInactivityRuleCommandService(inactivityRepo, logger)
	preferenceCommandService := commandservices.NewNotificationPreferenceCommandService(preferenceRepo, logger)

	alertQueryService := queryservices.NewAlertQueryService(alertRepo)
	thresholdQueryService := queryservices.NewThresholdQueryService(thresholdRepo)
	inactivityQueryService := queryservices.NewInactivityRuleQueryService(inactivityRepo)
	preferenceQueryService := queryservices.NewNotificationPreferenceQueryService(preferenceRepo)

	eventHandler := eventhandlers.NewIntegrationEventHandler(
		thresholdRepo,
		inactivityRepo,
		deviceActivityRepo,
		alertCommandService,
		logger,
	)
	kafkaController := controllers.NewKafkaController(kafkaProducer)

	router := rest.NewRouter(
		cfg,
		alertCommandService,
		alertQueryService,
		thresholdCommandService,
		thresholdQueryService,
		inactivityCommandService,
		inactivityQueryService,
		preferenceCommandService,
		preferenceQueryService,
		kafkaController,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	consumer := kafka.NewConsumptionConsumer(cfg, eventHandler, logger)
	if consumer.Enabled() {
		go consumer.Start(ctx, eventHandler)
	} else {
		logger.Println("kafka consumer disabled: missing brokers or topics")
	}

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Printf("server shutdown error: %v", err)
		}
	}()

	logger.Printf("alert service listening on :%s", cfg.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}
