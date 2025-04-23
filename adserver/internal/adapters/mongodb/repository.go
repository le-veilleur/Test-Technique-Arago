package mongodb

import (
	"context"
	"log"
	"time"

	"adserver/internal/domain"
	"adserver/internal/ports/out"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoRepository implémente l'interface AdRepository en utilisant MongoDB comme backend
type mongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository crée une nouvelle instance du repository MongoDB
// pour la collection "ads" dans la base de données spécifiée
func NewMongoRepository(db *mongo.Database) out.AdRepository {
	return &mongoRepository{collection: db.Collection("ads")}
}

// Create insère une nouvelle annonce dans la collection MongoDB et retourne son ID
func (r *mongoRepository) Create(ctx context.Context, ad *domain.Pub) (string, error) {
	start := time.Now()
	log.Printf("[MongoRepository.Create] start id=%s title=%q", ad.ID, ad.Title)
	_, err := r.collection.InsertOne(ctx, ad)
	if err != nil {
		log.Printf("[MongoRepository.Create] error: %v", err)
		return "", err
	}
	log.Printf("[MongoRepository.Create] completed in %v id=%s", time.Since(start), ad.ID)
	return ad.ID.String(), nil
}

// GetByID récupère une annonce par son ID UUID
func (r *mongoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pub, error) {
	start := time.Now()
	log.Printf("[MongoRepository.GetByID] start id=%s", id)
	var ad domain.Pub
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&ad)
	if err == mongo.ErrNoDocuments {
		log.Printf("[MongoRepository.GetByID] not found id=%s", id)
		return nil, nil
	}
	if err != nil {
		log.Printf("[MongoRepository.GetByID] error: %v", err)
		return nil, err
	}
	log.Printf("[MongoRepository.GetByID] completed in %v id=%s", time.Since(start), id)
	return &ad, nil
}

// Exists vérifie si une annonce existe dans la collection par son ID UUID
func (r *mongoRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	log.Printf("Vérification de l'existence de l'annonce avec ID: %s", id.String())

	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id})
	if err != nil {
		log.Printf("Erreur lors de la vérification de l'existence de l'annonce %s: %v", id.String(), err)
		return false, err
	}

	exists := count > 0
	log.Printf("Résultat de la vérification pour l'annonce %s: existe = %t", id.String(), exists)

	return exists, nil
}

// IncrementImpressions incrémente le compteur d'impressions et retourne le nouveau total
func (r *mongoRepository) IncrementImpressions(ctx context.Context, id uuid.UUID) (int64, error) {
	start := time.Now()
	log.Printf("[MongoRepository.IncrementImpressions] start id=%s", id)
	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"impressions": 1}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		log.Printf("[MongoRepository.IncrementImpressions] error: %v", result.Err())
		return 0, result.Err()
	}
	var ad domain.Pub
	if err := result.Decode(&ad); err != nil {
		log.Printf("[MongoRepository.IncrementImpressions] decode error: %v", err)
		return 0, err
	}
	log.Printf("[MongoRepository.IncrementImpressions] completed in %v id=%s impressions=%d", time.Since(start), id, ad.Impressions)
	return ad.Impressions, nil
}

// ResetImpressions réinitialise le compteur d'impressions et retourne l'ancien total
func (r *mongoRepository) ResetImpressions(ctx context.Context, id uuid.UUID) (int64, error) {
	start := time.Now()
	log.Printf("[MongoRepository.ResetImpressions] start id=%s", id)
	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"impressions": 0}},
		options.FindOneAndUpdate().SetReturnDocument(options.Before),
	)
	if result.Err() != nil {
		log.Printf("[MongoRepository.ResetImpressions] error: %v", result.Err())
		return 0, result.Err()
	}
	var ad domain.Pub
	if err := result.Decode(&ad); err != nil {
		log.Printf("[MongoRepository.ResetImpressions] decode error: %v", err)
		return 0, err
	}
	log.Printf("[MongoRepository.ResetImpressions] completed in %v id=%s oldImpressions=%d", time.Since(start), id, ad.Impressions)
	return ad.Impressions, nil
}

// DeleteExpired supprime toutes les annonces expirées et retourne le nombre supprimé
func (r *mongoRepository) DeleteExpired(ctx context.Context) (int64, error) {
	start := time.Now()
	log.Printf("[MongoRepository.DeleteExpired] start")
	result, err := r.collection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": time.Now()}})
	if err != nil {
		log.Printf("[MongoRepository.DeleteExpired] error: %v", err)
		return 0, err
	}
	log.Printf("[MongoRepository.DeleteExpired] completed in %v deletedCount=%d", time.Since(start), result.DeletedCount)
	return result.DeletedCount, nil
}

// List récupère une liste paginée d'annonces selon les critères de filtrage
func (r *mongoRepository) List(ctx context.Context, filter map[string]interface{}, offset, limit int64) ([]*domain.Pub, error) {
	start := time.Now()
	log.Printf("[MongoRepository.List] start filter=%v offset=%d limit=%d", filter, offset, limit)
	opts := options.Find().SetSkip(offset).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("[MongoRepository.List] error find: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var ads []*domain.Pub
	if err := cursor.All(ctx, &ads); err != nil {
		log.Printf("[MongoRepository.List] error decode all: %v", err)
		return nil, err
	}
	log.Printf("[MongoRepository.List] completed in %v returned=%d", time.Since(start), len(ads))
	return ads, nil
}
