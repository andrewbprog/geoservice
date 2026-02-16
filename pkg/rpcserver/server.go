package rpcserver

import (
	"context"
	"geoservice/internal/handler/rpc-handle"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type RPCServer struct {
	addr  string
	db    *gorm.DB
	redis *redis.Client
	geo   *rpc_handle.GeoRPCHandler
	user  *rpc_handle.UserRPCHandler
}

func NewRPCServer(addr string, db *gorm.DB, redis *redis.Client, geo *rpc_handle.GeoRPCHandler, user *rpc_handle.UserRPCHandler) *RPCServer {
	return &RPCServer{addr: addr, db: db, redis: redis, geo: geo, user: user}
}

// Serve - запуск RPC-сервера
func (s *RPCServer) Serve() error {
	err := rpc.RegisterName("Geo", s.geo)
	if err != nil {
		return err
	}
	err = rpc.RegisterName("User", s.user)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("RPC-сервер запущен на порту%s", s.addr)
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	<-stop
	log.Println("Остановка RPC-сервера...")
	return s.gracefulShutdown(listener)
}

func (s *RPCServer) gracefulShutdown(listener net.Listener) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// закрываем listener
	if err := listener.Close(); err != nil {
		log.Printf("Ошибка при закрытии listener: %v", err)
	}

	// закрываем БД
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}

	// закрываем Redis
	if s.redis != nil {
		_ = s.redis.Close()
	}

	<-ctx.Done()
	log.Println("RPC-сервер остановлен gracefully")
	return nil
}
