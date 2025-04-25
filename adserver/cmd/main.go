package main

import (
	"adserver/generated/ad_service"
	"adserver/generated/impression_service"
	"adserver/internal/adapters/grpc/handler"
	"adserver/internal/adapters/mongodb"
	"adserver/internal/application"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// loadEnv charge les variables d'environnement depuis le fichier .env
func loadEnv() {
	if err := godotenv.Load("/app/.env"); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	} else {
		log.Printf("Environment variables loaded successfully from .env file")
	}
}

// getEnvOrDefault récupère une variable d'environnement ou retourne une valeur par défaut
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		log.Printf("Config: %s=%s", key, value)
		return value
	}
	log.Printf("Config: %s=%s (default)", key, defaultValue)
	return defaultValue
}

func main() {
	startTime := time.Now()
	log.Printf("Starting Ad Server application...")

	log.Printf("System info: Go version=%s, OS=%s, Arch=%s, CPUs=%d",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU())

	loadEnv()

	grpcHost := getEnvOrDefault("GRPC_HOST", "0.0.0.0")
	grpcPort := getEnvOrDefault("GRPC_PORT", "50051")
	mongoURI := getEnvOrDefault("MONGODB_URI", "mongodb://mongodb:27017")
	mongoDatabase := getEnvOrDefault("MONGODB_DATABASE", "adserver")
	serviceName := getEnvOrDefault("SERVICE_NAME", "adserver")
	environment := getEnvOrDefault("ENVIRONMENT", "development")

	address := fmt.Sprintf("%s:%s", grpcHost, grpcPort)
	log.Printf("Listening on %s...", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	// Activer la réflexion pour grpcurl
	reflection.Register(grpcServer)

	// Connexion gRPC au microservice impression-tracker
	imprAddr := getEnvOrDefault("IMPRESSION_GRPC_ADDR", "impression-tracker:50052")
	impressionConn, err := grpc.NewClient(imprAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to impression-tracker %s: %v", imprAddr, err)
	}
	defer impressionConn.Close()

	impressionClient := impression_service.NewImpressionServiceClient(impressionConn)
	log.Printf("Connected to ImpressionService at %s", imprAddr)

	// Connexion MongoDB
	log.Printf("Connecting to MongoDB at %s...", mongoURI)
	mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer mongoCancel()
	client, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		if err := client.Disconnect(disconnectCtx); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()

	repo := mongodb.NewMongoRepository(client.Database(mongoDatabase))
	adService := application.NewAdService(repo)

	ad_service.RegisterAdServiceServer(grpcServer, handler.NewAdHandler(adService, impressionClient))
	log.Printf("AdService handler registered")

	// Nettoyage des publicités expirées
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			count, err := adService.DeleteExpired(context.Background())
			if err != nil {
				log.Printf("[CleanupExpired] error: %v", err)
			} else {
				log.Printf("[CleanupExpired] deleted %d ads", count)
			}
		}
	}()

	// Démarrer le serveur gRPC
	go func() {
		log.Printf("%s gRPC server running on %s (%s)",
			serviceName, address, environment)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Gestion du shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Printf("Server stopped. Total uptime: %v", time.Since(startTime))
}
