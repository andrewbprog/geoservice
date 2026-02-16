package app

import (
	"geoservice/config"
	"geoservice/internal/handler/http-handle"
	"geoservice/internal/provider/dadata"
	"geoservice/internal/repository"
	"geoservice/internal/service"
	"geoservice/pkg/database"
	httpServ "geoservice/pkg/httpserver"
	"geoservice/pkg/metrics"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-redis/redis"
	"log"
)

func Run(cfg *config.Config) {

	// Init JWT
	tokenAuth := jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)

	// Init DB
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Init Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Host + ":" + cfg.Redis.Port,
	})

	// UserRepo с кэшем
	userRepo := repository.NewUserRepository(db)
	userRepo = repository.NewProxyUserRepository(userRepo, rdb, cfg.UserCacheTTL)

	// UserService + Handler
	userService := service.NewUserService(userRepo)
	userHandler := http_handle.NewUserHandler(userService)

	// DaData с кэшем
	dadataServ := dadata.NewDaDataService(cfg.DadataAPIKey, cfg.DadataSecretKey)
	dadataServ = dadata.NewProxyDaDataService(dadataServ, rdb, cfg.DadataCacheTTL)

	// GeoService + Handler
	geoService := service.NewGeoService(dadataServ)
	geoHandler := http_handle.NewGeoHandler(geoService)

	// Метрики
	metrics.Init()

	// HTTP-сервер
	srv := httpServ.NewAPIServer(":"+cfg.AppPort, geoHandler, userHandler, tokenAuth, db, rdb)

	// Запуск HTTP-сервера
	if err := srv.Serve(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
