package main

import (
	"context"
	"impression-tracker/generated/impression_service"
	"impression-tracker/internal/adapters/dragonfly"
	"impression-tracker/internal/adapters/grpc/handler"
	"impression-tracker/internal/adapters/mongodb"
	"impression-tracker/internal/application"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func loadEnvFile() {
	if err := godotenv.Load("/app/.env"); err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	} else {
		log.Println("Loaded environment variables from .env")
	}
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		log.Printf("Config: %s=%s", key, val)
		return val
	}
	log.Printf("Config: %s=%s (default)", key, fallback)
	return fallback
}

func main() {
	start := time.Now()
	log.Println("Starting Impression Tracker Service...")

	log.Printf("Go version: %s | OS: %s | Arch: %s | CPUs: %d",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU())

	loadEnvFile()

	// Environment variables
	grpcAddr := getEnvOrDefault("GRPC_ADDR", ":50052")
	mongoURI := getEnvOrDefault("MONGO_URI", "mongodb://mongodb:27017")
	mongoDB := getEnvOrDefault("MONGO_DB", "impression_tracker")
	mongoColl := getEnvOrDefault("MONGO_COLLECTION", "impressions")
	dragonflyAddr := getEnvOrDefault("DRAGONFLY_ADDR", "localhost:6379")
	syncIntervalStr := getEnvOrDefault("SYNC_INTERVAL", "1m")

	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		log.Printf("Invalid SYNC_INTERVAL format: %v. Using 1m default.", err)
		syncInterval = time.Minute
	}

	// Dragonfly cache repo
	log.Printf("Connecting to Dragonfly: %s", dragonflyAddr)
	cacheRepo, err := dragonfly.NewDragonflyRepository(dragonflyAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Dragonfly: %v", err)
	}
	defer cacheRepo.Close()

	// MongoDB repository
	log.Printf("Connecting to MongoDB: %s", mongoURI)
	storeRepo, err := mongodb.NewMongoDBRepository(mongoURI, mongoDB, mongoColl)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer storeRepo.Close()

	// Application service
	service := application.NewService(cacheRepo, storeRepo, syncInterval)
	service.Start()
	defer service.Stop()

	// gRPC server setup
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}
	grpcServer := grpc.NewServer()
	impression_service.RegisterImpressionServiceServer(grpcServer, handler.NewServer(service))

	log.Printf("gRPC server listening on %s (Startup: %v)", grpcAddr, time.Since(start))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Signal handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sig := <-stop
	log.Printf("Received signal: %v", sig)

	// Ajout de l'attente avec timeout avant l'arrêt
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down...")
	shutdownStart := time.Now()
	grpcServer.GracefulStop() // Arrêt propre du serveur
	<-ctx.Done()              // Attendre que le contexte expire
	log.Printf("Server shut down in %v | Total uptime: %v", time.Since(shutdownStart), time.Since(start))
}
