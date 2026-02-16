package app

import (
	"geoservice/config"
	rpchandle "geoservice/internal/handler/rpc-handle"
	"geoservice/internal/provider/dadata"
	"geoservice/internal/repository"
	"geoservice/internal/service"
	"geoservice/pkg/database"
	"geoservice/pkg/metrics"
	"geoservice/pkg/rpcserver"
	"github.com/go-redis/redis"
	"log"
)

func RunRPC(cfg *config.Config) {
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

	// ---------- User ----------
	userRepo := repository.NewUserRepository(db)
	userRepo = repository.NewProxyUserRepository(userRepo, rdb, cfg.UserCacheTTL)

	userService := service.NewUserService(userRepo)
	userHandler := rpchandle.NewUserRPCHandler(userService)

	// ---------- Geo ----------
	dadataServ := dadata.NewDaDataService(cfg.DadataAPIKey, cfg.DadataSecretKey)
	dadataServ = dadata.NewProxyDaDataService(dadataServ, rdb, cfg.DadataCacheTTL)

	geoService := service.NewGeoService(dadataServ)
	geoHandler := rpchandle.NewGeoRPCHandler(geoService)

	// ---------- Metrics ----------
	metrics.Init()

	// ---------- RPC ----------
	srv := rpcserver.NewRPCServer(":"+cfg.RpcPort, db, rdb, geoHandler, userHandler)
	if err := srv.Serve(); err != nil {
		log.Fatalf("RPC server failed: %v", err)
	}
}
