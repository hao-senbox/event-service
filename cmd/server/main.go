package main

import (
	"context"
	"event-service/config"
	"event-service/internal/event"
	"event-service/internal/user"
	"event-service/pkg/constants"
	"event-service/pkg/consul"
	"event-service/pkg/firebase"
	"event-service/pkg/zap"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.LoadConfig()

	logger, err := zap.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	consulConn := consul.NewConsulConn(logger, cfg)
	consulClient := consulConn.Connect()
	defer consulConn.Deregister()

	mongoClient, err := connectToMongoDB(cfg.MongoURI)
	if err != nil {
		logger.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := waitPassing(consulClient, "go-main-service", 60*time.Second); err != nil {
		logger.Fatalf("Dependency not ready: %v", err)
	}

	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			logger.Fatal(err)
		}
	}()

	// Setup cron
	c := cron.New(cron.WithSeconds())
	client, _, _ := firebase.SetUpFireBase()
	userService := user.NewUserService(consulClient)
	eventCollection := mongoClient.Database(cfg.MongoDB).Collection("events")
	eventRepository := event.NewEventRepository(eventCollection)
	eventService := event.NewEventService(eventRepository, client, userService, c)
	eventHandler := event.NewEventHandler(eventService)

	router := gin.Default()
	event.RegisterRoutes(router, eventHandler)

	_, err = c.AddFunc("0 */1 * * * *", func() {
		log.Println("🔄 Cron master running...")
		ctx := context.WithValue(context.Background(), constants.TokenKey, os.Getenv("CRON_SERVICE_TOKEN"))
		if err := eventService.CronEventNotifications(ctx); err != nil {
			log.Printf("CronEventNotifications failed: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("AddFunc error: %v", err)
	}

	c.Start()
	defer c.Stop()

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Infof("Server running on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Error starting server: %v", err)
		}
	}()

	// ✅ Graceful shutdown: chờ tín hiệu kill
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Error shutting down server: %v", err)
	}
	logger.Info("Server stopped")
}

func connectToMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Println("Failed to connect to MongoDB")
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Println("Failed to ping MongoDB")
		return nil, err
	}

	log.Println("Successfully connected to MongoDB")
	return client, nil
}

func waitPassing(cli *consulapi.Client, name string, timeout time.Duration) error {
	dl := time.Now().Add(timeout)
	for time.Now().Before(dl) {
		entries, _, err := cli.Health().Service(name, "", true, nil)
		if err == nil && len(entries) > 0 {
			return nil // đã sẵn sàng
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("%s not ready in consul", name)
}
