package main

import (
	"database/sql"
	"fmt"
	"log"
	"notification-service/internal/config"
	internalGrpc "notification-service/internal/grpc"
	"notification-service/internal/handler"
	"notification-service/internal/repository"
	"notification-service/internal/service"
	pb "notification-service/proto"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"net"
)

func main() {
	cfg := config.LoadConfig()

	dbAddr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	
	var db *sql.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dbAddr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Waiting for database... attempt %d/5", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS notifications (
		id UUID PRIMARY KEY,
		user_id UUID NOT NULL,
		title VARCHAR(255),
		message TEXT,
		type VARCHAR(50),
		is_read BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT NOW()
	);`)

	repo := repository.NewPostgresRepository(db)
	svc := service.NewNotificationService(repo)
	h := handler.NewNotificationHandler(svc)

	router := gin.Default()

	router.GET("/health", h.HealthCheck)
	router.GET("/notifications", h.GetNotifications)
	router.POST("/notifications", h.CreateTestNotification)
	router.PUT("/notifications/read/:id", h.MarkRead)

	router.GET("/notifications/:id", func(c *gin.Context) { c.JSON(200, gin.H{"message": "GET single placeholder"}) })
	router.DELETE("/notifications/:id", func(c *gin.Context) { c.JSON(200, gin.H{"message": "DELETE placeholder"}) })
	router.POST("/notifications/email", func(c *gin.Context) { c.JSON(200, gin.H{"message": "POST email placeholder"}) })
	router.POST("/notifications/push", func(c *gin.Context) { c.JSON(200, gin.H{"message": "POST push placeholder"}) })
	router.POST("/notifications/sms", func(c *gin.Context) { c.JSON(200, gin.H{"message": "POST sms placeholder"}) })
	router.GET("/templates", func(c *gin.Context) { c.JSON(200, gin.H{"message": "GET templates placeholder"}) })
	router.POST("/templates", func(c *gin.Context) { c.JSON(200, gin.H{"message": "POST templates placeholder"}) })
	router.GET("/preferences", func(c *gin.Context) { c.JSON(200, gin.H{"message": "GET preferences placeholder"}) })
	router.PUT("/preferences", func(c *gin.Context) { c.JSON(200, gin.H{"message": "PUT preferences placeholder"}) })

	go func() {
		lis, err := net.Listen("tcp", ":" + cfg.GRPCPort)
		if err != nil {
			log.Printf("Failed to listen for gRPC: %v", err)
			return
		}
		s := grpc.NewServer()
		grpcServer := internalGrpc.NewNotificationServer(svc)
		pb.RegisterNotificationServiceServer(s, grpcServer)
		
		log.Printf("gRPC Notification Server starting on port %s...", cfg.GRPCPort)
		if err := s.Serve(lis); err != nil {
			log.Printf("Failed to serve gRPC: %v", err)
		}
	}()

	log.Printf("Notification Service starting on port %s...", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
