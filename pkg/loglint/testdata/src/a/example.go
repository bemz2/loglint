package a

import (
	"fmt"
	"log/slog"

	"go.uber.org/zap"
)

type Config struct {
	Password string
	Token    string
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sugar := logger.Sugar()

	cfg := Config{
		Password: "super-secret",
		Token:    "jwt-token",
	}

	password := "12345"
	token := "abc"
	apiKey := "key"

	// =========================
	// ✅ КОРРЕКТНЫЕ ЛОГИ
	// =========================

	slog.Info("server started")
	slog.Error("database connection failed")

	logger.Info("request completed")
	sugar.Infof("worker started")

	slog.Info("token validated") // должно быть OK по ТЗ

	sugar.Infow("request finished", "status", 200)

	// =========================
	// ❌ RULE 1 — Uppercase first letter
	// =========================

	slog.Info("Starting server")      // want "log message must start with a lowercase letter"
	logger.Error("Failed to connect") // want "log message must start with a lowercase letter"
	sugar.Warnf("Worker stopped")     // want "log message must start with a lowercase letter"

	// =========================
	// ❌ RULE 2 — Non-English text
	// =========================

	slog.Info("запуск сервера")        // want "log message must be in English only"
	logger.Error("ошибка подключения") // want "log message must be in English only"

	// =========================
	// ❌ RULE 3 — Special characters / emoji
	// =========================

	slog.Info("server started!")         // want "log message must not contain special symbols or emoji"
	slog.Info("connection failed!!!")    // want "log message must not contain special symbols or emoji"
	slog.Warn("something went wrong...") // want "log message must not contain special symbols or emoji"
	logger.Info("server ready 🚀")        // want "log message must not contain special symbols or emoji"

	// =========================
	// ❌ RULE 4 — Sensitive data in message
	// =========================

	slog.Info("user password: " + password) // want "log message may contain sensitive data"
	slog.Debug("api_key=" + apiKey)         // want "log message may contain sensitive data"
	slog.Info("token: " + token)            // want "log message may contain sensitive data"

	logger.Info("user password: " + cfg.Password) // want "log message may contain sensitive data"
	sugar.Infof("token is %s", token)             // want "log message may contain sensitive data"

	sugar.Infow("login attempt", "password", password) // want "log message may contain sensitive data"
	sugar.Infow("auth", "token", token)                // want "log message may contain sensitive data"

	// =========================
	// 🧪 Edge cases
	// =========================

	slog.Info("Server running") // want "log message must start with a lowercase letter"
	slog.Info("db-read failed")
	slog.Info(password) // want "log message may contain sensitive data"

	slog.Info(fmt.Sprintf("user password: %s", password)) // want "log message may contain sensitive data"

	slog.Info("config token: " + cfg.Token) // want "log message may contain sensitive data"

	slog.Info("http request failed")
	slog.Info("user_id resolved")

	slog.Info("服务启动") // want "log message must be in English only"
	slog.Info("!!!")  // want "log message must not contain special symbols or emoji"
	slog.Info("")
}
