package rpc_handle

import (
	"geoservice/internal/entity"
	"geoservice/internal/handler"
	http2 "geoservice/internal/handler/http-handle"
	middleware2 "geoservice/internal/infrastructure/middleware"
	"geoservice/pkg/metrics"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/jwtauth/v5"
	"log"
	"net/http"
	"net/http/pprof"
	"net/rpc"
	"time"
)

type Gateway struct {
	addr   string
	engine *gin.Engine
	client *rpc.Client
}

func NewGateway(addr, rpcAddr string, tokenAuth *jwtauth.JWTAuth) (*Gateway, error) {
	// Подключаемся к RPC серверу
	client, err := rpc.Dial("tcp", rpcAddr)
	if err != nil {
		return nil, err
	}

	r := gin.New()

	// ========== Middleware ==========
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(requestid.New())
	r.Use(metrics.PrometheusMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Authorization"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))

	// ======================= Public =======================

	// ========== Swagger ==========
	r.GET("/docs/swagger.json", gin.WrapF(middleware2.SwaggerSpecHandler))
	r.GET("/docs/swagger.yaml", gin.WrapF(middleware2.SwaggerSpecHandler))
	r.GET("/swagger", gin.WrapH(middleware2.SwaggerUIHandler()))
	r.GET("/docs", gin.WrapH(middleware2.RedocHandler()))

	// ========== Метрики ==========
	r.GET("/metrics", gin.WrapH(metrics.Handler()))

	public := r.Group("/api")
	{
		public.GET("/creation-token", handler.CreationTokenHandler(tokenAuth))
		public.GET("/health", http2.HealthCheck)

		// ========== Users ==========
		users := public.Group("/users")
		{
			// Создание пользователя
			users.POST("", func(c *gin.Context) {
				var req entity.CreateUserRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
					return
				}
				var reply CreateUserResponse
				if err := client.Call("User.CreateUser", req, &reply); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, reply)
			})

			// Получить пользователя по id
			users.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				var req GetUserRequest
				req.ID = id

				var reply GetUserResponse
				if err := client.Call("User.GetUser", req, &reply); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, reply)
			})

			// Обновить пользователя
			users.PUT("/:id", func(c *gin.Context) {
				id := c.Param("id")
				var req entity.UpdateUserRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
					return
				}
				// оборачиваем id + тело в одну структуру
				rpcReq := UpdateUserRequest{
					ID:   id,
					Data: req,
				}

				var reply UpdateUserResponse
				if err := client.Call("User.UpdateUser", rpcReq, &reply); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, reply)
			})

			// Удалить пользователя
			users.DELETE("/:id/delete", func(c *gin.Context) {
				id := c.Param("id")
				var req DeleteUserRequest
				req.ID = id

				var reply DeleteUserResponse
				if err := client.Call("User.DeleteUser", req, &reply); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, reply)
			})

			// Список пользователей
			users.GET("", func(c *gin.Context) {
				var req entity.Conditions
				if err := c.ShouldBindQuery(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query params"})
					return
				}

				var reply ListUsersResponse
				if err := client.Call("User.ListUsers", req, &reply); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, reply)
			})
		}
	}

	// ======================= Protected =======================

	// ========== Geo ==========
	protected := r.Group("/api/address")
	protected.Use(middleware2.JWTVerifier(tokenAuth))
	{
		protected.POST("/search", func(c *gin.Context) {
			var req SearchAddressRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
				return
			}
			var reply SearchAddressResponse
			if err := client.Call("Geo.SearchAddress", req, &reply); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, reply)
		})

		protected.POST("/geocode", func(c *gin.Context) {
			var req ReverseGeocodeRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
				return
			}
			var reply ReverseGeocodeResponse
			if err := client.Call("Geo.ReverseGeocode", req, &reply); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, reply)
		})
	}

	// ========== Профилирование ==========
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

	return &Gateway{
		addr:   addr,
		engine: r,
		client: client,
	}, nil
}

// Serve - запуск шлюза
func (g *Gateway) Serve() error {
	log.Printf("HTTP Gateway запущен на порту%s", g.addr)
	return g.engine.Run(g.addr)
}
