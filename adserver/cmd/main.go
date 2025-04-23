package main

import (
	"adserver/generated/ad_service"
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
)

// loadEnv charge les variables d'environnement depuis le fichier .env
func loadEnv() {
	// Dans le conteneur Docker, le .env est monté à /app/.env
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

// main est le point d'entrée de l'application.
// Il configure et démarre le serveur gRPC avec MongoDB comme base de données.
func main() {
	startTime := time.Now()
	log.Printf("Starting Ad Server application...")

	// Affichage d'informations système
	log.Printf("System info: Go version=%s, OS=%s, Arch=%s, CPUs=%d",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU())

	// Chargement des variables d'environnement
	loadEnv()

	// Récupération des variables d'environnement avec valeurs par défaut
	grpcHost := getEnvOrDefault("GRPC_HOST", "0.0.0.0")
	grpcPort := getEnvOrDefault("GRPC_PORT", "50051")
	mongoURI := getEnvOrDefault("MONGODB_URI", "mongodb://mongodb:27017")
	mongoDatabase := getEnvOrDefault("MONGODB_DATABASE", "adserver")
	serviceName := getEnvOrDefault("SERVICE_NAME", "adserver")
	environment := getEnvOrDefault("ENVIRONMENT", "development")

	// Configuration du serveur gRPC
	address := fmt.Sprintf("%s:%s", grpcHost, grpcPort)
	log.Printf("Attempting to listen on %s...", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("TCP listener established successfully on %s", address)

	// Création d'une nouvelle instance du serveur gRPC
	grpcServer := grpc.NewServer()
	log.Printf("gRPC server instance created")

	// Configuration de la connexion MongoDB avec un timeout de 10 secondes
	log.Printf("Establishing connection to MongoDB at %s...", mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connexion à MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Test de la connexion MongoDB
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Printf("Connected successfully to MongoDB at %s, database: %s", mongoURI, mongoDatabase)

	// Déconnexion propre de MongoDB à la fin du programme
	defer func() {
		log.Printf("Closing MongoDB connection...")
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		if err = client.Disconnect(disconnectCtx); err != nil {
			log.Printf("Error when disconnecting from MongoDB: %v", err)
		} else {
			log.Printf("MongoDB connection closed successfully")
		}
	}()

	// Initialisation du repository MongoDB
	repo := mongodb.NewMongoRepository(client.Database(mongoDatabase))
	log.Printf("MongoDB repository initialized")

	// Création du service d'annonces avec le repository MongoDB
	adService := application.NewAdService(repo)
	log.Printf("Ad service initialized")

	// Enregistrement du handler gRPC pour le service d'annonces
	ad_service.RegisterAdServiceServer(grpcServer, handler.NewAdHandler(adService))
	log.Printf("Ad service handler registered with gRPC server")

	// Configuration de la gestion des signaux pour un arrêt propre
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Printf("Signal handling configured (SIGINT, SIGTERM)")

	// Calcul du temps de démarrage
	startupDuration := time.Since(startTime)

	// Démarrage du serveur gRPC dans une goroutine séparée
	go func() {
		log.Printf("Starting %s gRPC server on %s in %s environment (startup time: %v)",
			serviceName, address, environment, startupDuration)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Attente d'un signal d'arrêt (Ctrl+C ou SIGTERM)
	sig := <-stop
	log.Printf("Received signal: %v", sig)

	// Arrêt propre du serveur gRPC
	log.Println("Initiating graceful shutdown of the server...")
	serverStopTime := time.Now()
	grpcServer.GracefulStop()
	log.Printf("Server gracefully stopped in %v", time.Since(serverStopTime))
	log.Printf("Total uptime: %v", time.Since(startTime))
}
