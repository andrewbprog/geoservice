package main

import (
	"geoservice/config"
	_ "geoservice/docs"
	"geoservice/internal/app"
	"log"
	_ "net/http/pprof"
	"os"
)

// @title           GeoService API
// @version         1.0
// @description     API для работы с геоданными и адресами через DaData
// @termsOfService  http://swagger.io/terms/
// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html
// @host            localhost:8080
// @BasePath        /api
// @schemes         http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Авторизация с token в формате: Bearer {token}
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./app [rpc-server|gateway]")
	}

	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	switch os.Args[1] {
	// rpc-server и gateway - названия контейнеров внутри docker-compose
	case "rpc-server":
		app.RunRPC(cfg)
	case "gateway":
		app.RunGateway(cfg)
	default:
		log.Fatalf("Unknown mode: %s", os.Args[1])
	}
}
