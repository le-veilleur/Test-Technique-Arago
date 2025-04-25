package mongodb

import (
	"context"
	"fmt"
	"time"

	"impression-tracker/internal/ports/out"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBRepository implémente l'interface MetricsRepository pour stocker les deltas d'impressions
// de manière persistante dans MongoDB.
type MongoDBRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// impressionDelta représente un document MongoDB stockant les informations sur un delta d'impressions.
type impressionDelta struct {
	AdID     string    `bson:"ad_id"`     // Identifiant de la publicité
	Delta    int64     `bson:"delta"`     // Nombre d'impressions à synchroniser
	DateTime time.Time `bson:"date_time"` // Date et heure de la synchronisation
}

// NewMongoDBRepository crée une nouvelle instance de MongoDBRepository.
// Elle établit une connexion avec le serveur MongoDB et vérifie que la connexion fonctionne.
func NewMongoDBRepository(uri, database, collection string) (*MongoDBRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &MongoDBRepository{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

// PersistDelta enregistre un delta d'impressions dans MongoDB.
// Le document contient l'ID de la publicité, le nombre d'impressions et la date/heure.
func (r *MongoDBRepository) PersistDelta(ctx context.Context, adID string, delta int64) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	doc := impressionDelta{
		AdID:     adID,
		Delta:    delta,
		DateTime: time.Now(),
	}

	_, err := collection.InsertOne(ctx, doc)
	return err
}

// Close ferme la connexion avec le serveur MongoDB.
func (r *MongoDBRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Disconnect(ctx)
}

// Ensure MongoDBRepository implements the MetricsRepository interface
var _ out.MetricsRepository = (*MongoDBRepository)(nil)
