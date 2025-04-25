# Test Technique Arago - Système de Publicités

Ce projet implémente un système de gestion de publicités avec suivi des impressions, utilisant une architecture microservices basée sur gRPC et suivant les principes de l'architecture hexagonale.

## Architecture

Le système est composé de deux microservices principaux, chacun suivant une architecture hexagonale :

1. **Ad Server** (`adserver/`)
   - **Domaine** : Gestion des publicités
   - **Ports primaires** : API gRPC
   - **Ports secondaires** : MongoDB, Impression Tracker
   - **Adaptateurs** : gRPC, MongoDB, Impression Tracker

2. **Impression Tracker** (`impression-tracker/`)
   - **Domaine** : Suivi des impressions
   - **Ports primaires** : API gRPC
   - **Ports secondaires** : MongoDB
   - **Adaptateurs** : gRPC, MongoDB

## Fonctionnalités

### Ad Server
- Création de publicités avec titre, description et date d'expiration
- Génération automatique d'URLs uniques pour chaque publicité
- Nettoyage automatique des publicités expirées
- Interface gRPC pour la gestion des publicités
- Communication synchrone avec le service d'impressions pour incrémenter le compteur

### Impression Tracker
- Suivi des impressions publicitaires
- Stockage des données d'impression avec horodatage
- API gRPC pour la notification des impressions
- Statistiques d'impressions par publicité

## Prérequis

- Go 1.21 ou supérieur
- Docker et Docker Compose
- grpcurl (pour tester les APIs)
- protoc (protobuf compiler)

## Installation

1. Cloner le repository :
```bash
git clone https://github.com/votre-username/test-technique-arago.git
cd test-technique-arago
```

2. Construire les images Docker :
```bash
docker-compose build
```

3. Démarrer les services :
```bash
docker-compose up -d
```

## Configuration

Les services sont configurables via des variables d'environnement dans le fichier `.env` :

### Ad Server
```env
GRPC_HOST=0.0.0.0
GRPC_PORT=50051
MONGODB_URI=mongodb://mongodb:27017
MONGODB_DATABASE=adserver
IMPRESSION_GRPC_ADDR=impression-tracker:50052
ME_CONFIG_BASICAUTH_USERNAME=admin
ME_CONFIG_BASICAUTH_PASSWORD=admin123
```

### Impression Tracker
```env
GRPC_HOST=0.0.0.0
GRPC_PORT=50052
MONGODB_URI=mongodb://mongodb:27017
MONGODB_DATABASE=impression_tracker
```

## Définitions des Services gRPC

### Ad Service (`adserver/proto/ad_service.proto`)
```protobuf
syntax = "proto3";
package ad.v1;
option go_package = "generated/ad_service";
import "google/protobuf/timestamp.proto";

service AdService {
  rpc CreateAd(CreateAdRequest) returns (AdResponse);
  rpc GetAd(GetAdRequest) returns (AdResponse);
  rpc ServeAd(ServeAdRequest) returns (ServeAdResponse);
  rpc GetImpressionCount(GetImpressionCountRequest) returns (GetImpressionCountResponse);
  rpc IncrementImpressions(IncrementImpressionsRequest) returns (IncrementImpressionsResponse);
  rpc DeleteExpired(DeleteExpiredRequest) returns (DeleteExpiredResponse);
}

message CreateAdRequest {
  string title = 1;
  string description = 2;
  google.protobuf.Timestamp expires_at = 3;
}

message AdResponse {
  string id = 1;
  string title = 2;
  string description = 3;
  string url = 4;
  google.protobuf.Timestamp expires_at = 5;
  int64 impressions = 6;
}

message ServeAdRequest { string id = 1; }
message ServeAdResponse { string url = 1; int64 impressions = 2; }
message GetImpressionCountRequest { string ad_id = 1; }
message GetImpressionCountResponse { int64 impressions = 1; }
message IncrementImpressionsRequest { string ad_id = 1; }
message IncrementImpressionsResponse { int64 impressions = 1; }
message DeleteExpiredRequest {}
message DeleteExpiredResponse { int64 deleted_count = 1; }
```

### Impression Service (`impression-tracker/proto/impression_service.proto`)
```protobuf
syntax = "proto3";
package impression.v1;
option go_package = "generated/impression_service";

service ImpressionService {
  rpc Track(TrackRequest) returns (TrackResponse);
  rpc GetCount(GetCountRequest) returns (GetCountResponse);
}

message TrackRequest { string ad_id = 1; }
message TrackResponse {}
message GetCountRequest { string ad_id = 1; }
message GetCountResponse { int64 count = 1; }
```

## Utilisation avec grpcurl

### 1. Création d'une publicité
```bash
grpcurl -plaintext \
  -d '{
    "title": "Ma publicité",
    "description": "Description de la publicité",
    "expiresAt": "2025-11-01T00:00:00Z"
  }' \
  localhost:50051 \
  ad.v1.AdService/CreateAd
```
**Réponse** :
```json
{
  "id": "94ae2f4f-e619-44df-aba5-58d083a44d2d",
  "title": "Ma publicité",
  "description": "Description de la publicité",
  "url": "https://localhost:8080/ads/94ae2f4f-e619-44df-aba5-58d083a44d2d",
  "expiresAt": "2025-11-01T00:00:00Z",
  "impressions": 0
}
```

### 2. Diffuser la publicité
```bash
grpcurl -plaintext \
  -d '{"id": "497119be-a147-4c5c-a7b4-8ede5a47925c"}' \
  localhost:50051 \
  ad.v1.AdService/ServeAd
```
**Réponse** :
```json
{
  "url": "https://localhost:8080/ads/497119be-a147-4c5c-a7b4-8ede5a47925c?impression_id=...",
  "impressions": 1
}
```

### 3. Obtenir le nombre d'impressions
```bash
grpcurl -plaintext \
  -d '{"adId": "497119be-a147-4c5c-a7b4-8ede5a47925c"}' \
  localhost:50052 \
  impression.v1.ImpressionService/GetCount
```
**Réponse** :
```json
{ "count": 1 }
```

## Structure du Projet

```
.
├── adserver/
│   ├── cmd/
│   ├── internal/
│   └── proto/
├── impression-tracker/
│   ├── cmd/
│   ├── internal/
│   └── proto/
├── docker-compose.yml
└── README.md
```

## Développement

1. Générer les fichiers gRPC :
```bash
protoc --go_out=. --go-grpc_out=. --proto_path=proto proto/ad_service.proto
protoc --go_out=. --go-grpc_out=. --proto_path=proto proto/impression_service.proto
```

2. Lancer les services :
```bash
go run adserver/cmd/main.go
go run impression-tracker/cmd/main.go
```

## Licence

TEST TECHNIQUE

