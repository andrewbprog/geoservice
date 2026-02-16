package httpserver

import (
	"context"
	"errors"
	http2 "geoservice/internal/handler/http-handle"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// APIServer структура сервера
type APIServer struct {
	server *http.Server
	addr   string
	engine *gin.Engine
	db     *gorm.DB
	redis  *redis.Client
}

// NewAPIServer создает новый экземпляр сервера
func NewAPIServer(
	addr string,
	geoHandler *http2.GeoHandler,
	userHandler *http2.UserHandler,
	tokenAuth *jwtauth.JWTAuth,
	db *gorm.DB,
	redis *redis.Client,
) *APIServer {
	r := gin.New()

	// Настройка middleware
	http2.SetupMiddleware(r)

	// Настройка маршрутов
	http2.SetupRoutes(r, geoHandler, userHandler, tokenAuth)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &APIServer{
		server: server,
		addr:   addr,
		engine: r,
		db:     db,
		redis:  redis,
	}
}

// Serve запускает сервер с graceful shutdown
func (s *APIServer) Serve() error {
	// Канал для получения сигналов остановки
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Geoservice стартовал на порту%s", s.addr)
		log.Printf("Swagger UI доступен по адресу: http://localhost%s/swagger", s.addr)
		log.Printf("Документация по адресу: http://localhost%s/docs", s.addr)

		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка при старте сервера: %v", err)
		}
	}()

	// Ожидание сигнала остановки
	<-stop
	log.Println("Получен сигнал остановки...")

	// Graceful shutdown с контекстом и таймаутом 5 секунд
	return s.gracefulShutdown()
}

// gracefulShutdown корректно завершает работу сервера с таймаутом 5 секунд
func (s *APIServer) gracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при graceful shutdown: %v", err)
		return err
	}

	// Закрываем БД
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Ошибка при закрытии БД: %v", err)
			} else {
				log.Println("База данных закрыта")
			}
		}
	}

	// Закрываем Redis
	if s.redis != nil {
		if err := s.redis.Close(); err != nil {
			log.Printf("Ошибка при закрытии Redis: %v", err)
		} else {
			log.Println("Redis закрыт")
		}
	}

	log.Println("HTTP-сервер остановлен gracefully")
	return nil
}
