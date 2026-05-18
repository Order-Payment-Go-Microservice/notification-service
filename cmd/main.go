package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/Order-Payment-Go-Microservice/notification-service/internal/config"
	"github.com/Order-Payment-Go-Microservice/notification-service/internal/database"
	internalGrpc "github.com/Order-Payment-Go-Microservice/notification-service/internal/grpc"
	"github.com/Order-Payment-Go-Microservice/notification-service/internal/handler"
	"github.com/Order-Payment-Go-Microservice/notification-service/internal/repository"
	"github.com/Order-Payment-Go-Microservice/notification-service/internal/service"
	notificationv1 "github.com/Order-Payment-Go-Microservice/proto-generation/gen/notification/v1"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	dbAddr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", dbAddr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("Redis connection failed: %v", err)
	}

	repo := repository.NewPostgresRepository(db)
	notificationSvc := service.NewNotificationService(repo, rdb)
	emailSvc := service.NewEmailService()
	notificationHandler := handler.NewNotificationHandler(notificationSvc)

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Printf("NATS connection failed: %v", err)
	} else {
		defer nc.Close()
		_, err = nc.Subscribe("notifications", func(m *nats.Msg) {
			var data map[string]string
			if err := json.Unmarshal(m.Data, &data); err != nil {
				log.Printf("[NATS] invalid payload: %v", err)
				return
			}
			log.Printf("[NATS Consumer] Received notification: %v", data)

			uID, err := uuid.Parse(data["user_id"])
			if err != nil {
				log.Printf("[NATS] invalid user_id: %v", err)
				return
			}

			nType := data["type"]
			if nType == "" {
				nType = "push"
			}

			if _, err := notificationSvc.CreateNotification(uID, data["title"], data["message"], nType); err != nil {
				log.Printf("[NATS] create notification failed: %v", err)
				return
			}

			if nType == "email" {
				to := data["email"]
				if to == "" {
					to = "user@example.com"
				}
				_ = emailSvc.SendEmail(to, data["title"], data["message"])
			}
		})
		if err != nil {
			log.Printf("NATS subscribe failed: %v", err)
		} else {
			log.Println("NATS Consumer subscribed to 'notifications'")
		}
	}

	go func() {
		grpcPort := cfg.GRPCPort
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		notificationv1.RegisterNotificationServiceServer(s, internalGrpc.NewNotificationServer(notificationSvc))
		log.Printf("gRPC Notification Server starting on port %s...", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	router := gin.Default()
	router.GET("/health", notificationHandler.HealthCheck)
	router.GET("/notifications", notificationHandler.GetHistory)
	router.POST("/notifications", notificationHandler.CreateNotification)
	router.PATCH("/notifications/:id/read", notificationHandler.MarkRead)

	log.Printf("Notification Service starting on port %s...", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
