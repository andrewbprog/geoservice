package app

import (
	"geoservice/config"
	rpchandle "geoservice/internal/handler/rpc-handle"
	"geoservice/pkg/metrics"
	"github.com/go-chi/jwtauth/v5"
	"log"
)

func RunGateway(cfg *config.Config) {
	// Init JWT
	tokenAuth := jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)

	// ---------- Metrics ----------
	metrics.Init()

	// ---------- Gateway ----------
	gw, err := rpchandle.NewGateway(":"+cfg.AppPort, "rpc:"+cfg.RpcPort, tokenAuth)
	if err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}

	if err := gw.Serve(); err != nil {
		log.Fatalf("Gateway failed: %v", err)
	}
}
