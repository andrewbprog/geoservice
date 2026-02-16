package http_handle

import (
	"geoservice/internal/handler"
	middleware2 "geoservice/internal/infrastructure/middleware"
	"geoservice/pkg/metrics"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/jwtauth/v5"
	"net/http/pprof"
	"time"
)

// SetupMiddleware настраивает middleware
func SetupMiddleware(r *gin.Engine) {
	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// RequestID (через gin-contrib/requestid)
	r.Use(requestid.New())

	// Prometheus
	r.Use(metrics.PrometheusMiddleware())

	// CORS настройки
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Authorization"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))
}

func SetupRoutes(
	r *gin.Engine,
	geoHandle *GeoHandler,
	userHandle *UserHandler,
	tokenAuth *jwtauth.JWTAuth,
) {
	// -------------------- Public --------------------

	// Метрики Prometheus
	r.GET("/metrics", gin.WrapH(metrics.Handler()))

	// ---------- Swagger / ReDoc ----------
	r.GET("/docs/swagger.json", gin.WrapF(middleware2.SwaggerSpecHandler))
	r.GET("/docs/swagger.yaml", gin.WrapF(middleware2.SwaggerSpecHandler))
	r.GET("/swagger", gin.WrapH(middleware2.SwaggerUIHandler()))
	r.GET("/docs", gin.WrapH(middleware2.RedocHandler()))

	public := r.Group("/api")
	{
		public.GET("/creation-token", handler.CreationTokenHandler(tokenAuth))
		public.GET("/health", HealthCheck)
		users := public.Group("/users")
		{
			users.POST("", userHandle.CreateUser)
			users.GET("", userHandle.ListUsers)
			users.GET("/:id", userHandle.GetUser)
			users.PUT("/:id", userHandle.UpdateUser)
			users.DELETE("/:id/delete", userHandle.DeleteUser)
		}
	}

	// -------------------- Protected --------------------

	protected := r.Group("/api/address")
	protected.Use(middleware2.JWTVerifier(tokenAuth))
	{
		protected.POST("/search", geoHandle.SearchAddress)
		protected.POST("/geocode", geoHandle.ReverseGeocode)
	}
	pprofGroup := r.Group("/api/pprof")
	pprofGroup.Use(middleware2.JWTVerifier(tokenAuth))
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}

}
