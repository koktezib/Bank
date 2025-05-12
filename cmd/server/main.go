package main

import (
	"Bank/internal/config"
	"Bank/internal/handler"
	"Bank/internal/middleware"
	"Bank/internal/repository"
	"Bank/internal/service"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
)

func runMigrations(dsn string, migrationsPath string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	log.Println("✅ Миграции применены")
	return nil
}

func startScheduler(interval time.Duration, svc *service.CreditService) {
	log.Println("Шедулер: первичный запуск обработки платежей…")
	if err := svc.ProcessDuePayments(); err != nil {
		log.Printf("Ошибка при первичной обработке платежей: %v", err)
	} else {
		log.Println("Первичная обработка завершена")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Шедулер: очередная проверка просроченных платежей…")
		if err := svc.ProcessDuePayments(); err != nil {
			log.Printf("Ошибка при обработке платежей: %v", err)
		} else {
			log.Println("Обработка завершена")
		}
	}
}

func main() {
	cfg := config.Load()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	runMigrations(dsn, "file://./migrations")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	authH := handler.NewAuthHandler(authSvc)

	r := mux.NewRouter()
	r.HandleFunc("/register", authH.Register).Methods("POST")
	r.HandleFunc("/login", authH.Login).Methods("POST")

	authRouter := r.PathPrefix("/").Subrouter()
	authRouter.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	mailCfg := service.MailConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUser,
		Password: cfg.SMTPPass,
		From:     cfg.SMTPUser,
	}
	mailSvc := service.NewMailService(mailCfg)
	accRepo := repository.NewAccountRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	accSvc := service.NewAccountService(db, userRepo, accRepo, txRepo, mailSvc)
	accH := handler.NewAccountHandler(accSvc)

	authRouter.HandleFunc("/accounts", accH.CreateAccount).Methods("POST")
	authRouter.HandleFunc("/accounts", accH.ListAccounts).Methods("GET")
	authRouter.HandleFunc("/accounts/deposit", accH.Deposit).Methods("POST")
	authRouter.HandleFunc("/accounts/withdraw", accH.Withdraw).Methods("POST")
	authRouter.HandleFunc("/transfer", accH.Transfer).Methods("POST")

	cardRepo := repository.NewCardRepository(db)
	cardSvc := service.NewCardService(
		cfg.PGPPublicKey,
		cfg.PGPPrivateKey,
		cfg.PGPPrivateKeyPassphrase,
		cfg.HMACSecret,
		cardRepo,
		accRepo,
	)
	cardH := handler.NewCardHandler(cardSvc)

	authRouter.HandleFunc("/cards", cardH.Create).Methods("POST")
	authRouter.HandleFunc("/cards", cardH.List).Methods("GET")

	cbrSvc := service.NewCBRService()
	creditRepo := repository.NewCreditRepository(db)
	scheduleRepo := repository.NewPaymentScheduleRepository(db)
	creditSvc := service.NewCreditService(db, creditRepo, scheduleRepo, accRepo, cbrSvc)
	creditH := handler.NewCreditHandler(creditSvc)

	authRouter.HandleFunc("/credits", creditH.Create).Methods("POST")
	authRouter.HandleFunc("/credits/{creditId}/schedule", creditH.GetSchedule).Methods("GET")

	analyticsSvc := service.NewAnalyticsService(txRepo, accRepo, scheduleRepo)
	analyticsH := handler.NewAnalyticsHandler(analyticsSvc)

	authRouter.HandleFunc("/analytics", analyticsH.GetStats).Methods("GET")
	authRouter.HandleFunc("/accounts/{accountId}/predict", analyticsH.Predict).Methods("GET")

	go startScheduler(5*time.Hour, creditSvc)
	log.Println("Server is running on :8080")

	log.Fatal(http.ListenAndServe(":8080", r))
}
