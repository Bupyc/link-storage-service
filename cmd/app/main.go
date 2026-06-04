package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Bupyc/link-storage-service/internal/cache"
	"github.com/Bupyc/link-storage-service/internal/config"
	"github.com/Bupyc/link-storage-service/internal/handler"
	"github.com/Bupyc/link-storage-service/internal/repository"
	"github.com/Bupyc/link-storage-service/internal/service"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to redis:", err)
	}

	repo := repository.NewRedisLinkRepository(redisClient)
	memoryCache := cache.NewMemoryCache()
	linkService := service.NewLinkService(repo, memoryCache)
	LinkHandler := handler.NewLinkHandler(linkService)

	mux := http.NewServeMux()
	LinkHandler.RegisterRoutes(mux)
	fmt.Println("app starting on port:", cfg.AppPort)

	if err := http.ListenAndServe(":"+cfg.AppPort, mux); err != nil {
		log.Fatal(err)
	}
}
