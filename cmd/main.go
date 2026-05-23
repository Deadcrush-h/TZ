package main

import (
	"tz-app/internal/handlers"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"tz-app/internal/config"
	"tz-app/internal/database"
	"tz-app/internal/logger"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем логгер
	if err := logger.Init(cfg.LogLevel, cfg.LogEncoding); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting application...",
		zap.String("port", cfg.Port),
		zap.String("log_level", cfg.LogLevel),
	)

	// Подключаемся к БД
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	defer sqlDB.Close()

	// Создаем обработчики
	deptHandler := handlers.NewDepartmentHandler(db)
	empHandler := handlers.NewEmployeeHandler(db)

	// Настраиваем роутер
	mux := http.NewServeMux()

	// Регистрируем endpoints
	mux.HandleFunc("POST /api/departments", deptHandler.CreateDepartment)
	mux.HandleFunc("GET /api/departments/{id}", deptHandler.GetDepartment)
	mux.HandleFunc("PATCH /api/departments/{id}", deptHandler.UpdateDepartment)
	mux.HandleFunc("DELETE /api/departments/{id}", deptHandler.DeleteDepartment)
	mux.HandleFunc("POST /api/departments/{id}/employees", empHandler.CreateEmployee)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Запускаем сервер
	port := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Server starting", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Ждем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	if err := server.Close(); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
