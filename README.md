# Test Technique Arago - Microservices d'Ad Server et Impression Tracker

Ce projet comprend deux microservices :
- **Ad Server** : Service de gestion et de diffusion de publicités
- **Impression Tracker** : Service de suivi des impressions publicitaires

## Prérequis

- Docker et Docker Compose
- Go 1.23 ou supérieur
- Protoc (protobuf compiler)

## Architecture

```
.
├── adserver/                  # Service de gestion des publicités
│   ├── cmd/                   # Point d'entrée du service
│   ├── internal/              # Code interne du service
│   ├── proto/                 # Définitions gRPC
│   └── docker/                # Configuration Docker
├── impression-tracker/        # Service de suivi des impressions
│   ├── cmd/                   # Point d'entrée du service
│   ├── internal/              # Code interne du service
│   ├── proto/                 # Définitions gRPC
│   └── docker/                # Configuration Docker
└── README.md                  # Documentation
```

## Démarrage des services

1. Créer le réseau Docker pour la communication entre les services :
```bash
docker network create microservices-network
```

2. Démarrer l'Ad Server :
```bash
cd adserver/docker
docker-compose up --build
```

3. Démarrer l'Impression Tracker :
```bash
cd impression-tracker/docker
docker-compose up --build
```

Les services seront accessibles aux adresses suivantes :
- Ad Server : `localhost:50051`
- Impression Tracker : `localhost:50052`
- MongoDB Ad Server : `localhost:27017`
- MongoDB Impression Tracker : `localhost:27018`
- Mongo Express Ad Server : `localhost:8081`
- Mongo Express Impression Tracker : `localhost:8082`
- Dragonfly (cache) : `localhost:6379`

## Définitions gRPC

### Ad Server Service

```protobuf
service AdService {
  // Servir une publicité
  rpc ServeAd(ServeAdRequest) returns (ServeAdResponse) {}
  
  // Supprimer les publicités expirées
  rpc DeleteExpired(DeleteExpiredRequest) returns (DeleteExpiredResponse) {}
}

message ServeAdRequest {
  string ad_id = 1;
}

message ServeAdResponse {
  string ad_id = 1;
  string title = 2;
  string content = 3;
  int64 impressions = 4;
}
```

### Impression Tracker Service

```protobuf
service ImpressionService {
  // Enregistrer une nouvelle impression
  rpc TrackImpression(TrackImpressionRequest) returns (TrackImpressionResponse) {}
  
  // Obtenir le nombre d'impressions pour une publicité
  rpc GetImpressionCount(GetImpressionCountRequest) returns (GetImpressionCountResponse) {}
}

message TrackImpressionRequest {
  string ad_id = 1;
  string impression_id = 2;
}

message TrackImpressionResponse {
  bool success = 1;
}

message GetImpressionCountRequest {
  string ad_id = 1;
}

message GetImpressionCountResponse {
  int64 count = 1;
}
```

## Tester les services

### Prérequis pour tester

1. Installer grpcurl :
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

2. Installer les outils de développement gRPC :
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Tester l'Ad Server

1. Servir une publicité :
```bash
grpcurl -plaintext -d '{"ad_id": "11adbe56-2c75-4d34-a361-8e442daf0e60"}' localhost:50051 ad.AdService/ServeAd
```

2. Supprimer les publicités expirées :
```bash
grpcurl -plaintext localhost:50051 ad.AdService/DeleteExpired
```

### Tester l'Impression Tracker

1. Enregistrer une impression :
```bash
grpcurl -plaintext -d '{"ad_id": "11adbe56-2c75-4d34-a361-8e442daf0e60", "impression_id": "0f1c87f8-d11f-4e0e-bbb9-67f8ce4f7361"}' localhost:50052 impression.ImpressionService/TrackImpression
```

2. Obtenir le nombre d'impressions :
```bash
grpcurl -plaintext -d '{"ad_id": "11adbe56-2c75-4d34-a361-8e442daf0e60"}' localhost:50052 impression.ImpressionService/GetImpressionCount
```

### Liste des services disponibles

Pour voir tous les services et méthodes disponibles :

```bash
# Pour l'Ad Server
grpcurl -plaintext localhost:50051 list

# Pour l'Impression Tracker
grpcurl -plaintext localhost:50052 list
```

### Description des services

Pour voir la description détaillée d'un service :

```bash
# Pour l'Ad Server
grpcurl -plaintext localhost:50051 describe ad.AdService

# Pour l'Impression Tracker
grpcurl -plaintext localhost:50052 describe impression.ImpressionService
```

## Exemple d'utilisation

### Créer une publicité

```bash
curl -X POST http://localhost:50051/v1/ads \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Publicité Test",
    "content": "Contenu de la publicité",
    "expires_at": "2025-12-31T23:59:59Z"
  }'
```

### Servir une publicité

```bash
curl -X GET http://localhost:50051/v1/ads/serve/{ad_id}
```

### Obtenir le nombre d'impressions

```bash
curl -X GET http://localhost:50052/v1/impressions/count/{ad_id}
```

## Configuration

Les services peuvent être configurés via des variables d'environnement :

### Ad Server

- `GRPC_HOST` : Hôte du serveur gRPC (défaut: 0.0.0.0)
- `GRPC_PORT` : Port du serveur gRPC (défaut: 50051)
- `MONGODB_URI` : URI de connexion MongoDB (défaut: mongodb://mongodb:27017)
- `MONGODB_DATABASE` : Nom de la base de données (défaut: adserver)
- `SERVICE_NAME` : Nom du service (défaut: adserver)
- `ENVIRONMENT` : Environnement d'exécution (défaut: development)

### Impression Tracker

- `GRPC_ADDR` : Adresse du serveur gRPC (défaut: :50052)
- `MONGO_URI` : URI de connexion MongoDB (défaut: mongodb://mongodb:27017)
- `MONGO_DB` : Nom de la base de données (défaut: impression_tracker)
- `MONGO_COLLECTION` : Nom de la collection (défaut: impressions)
- `DRAGONFLY_ADDR` : Adresse du serveur Dragonfly (défaut: localhost:6379)
- `SYNC_INTERVAL` : Intervalle de synchronisation (défaut: 1m)

## Architecture technique

- **Ad Server** :
  - Stockage persistant : MongoDB
  - API : gRPC
  - Langage : Go

- **Impression Tracker** :
  - Cache : Dragonfly (compatible Redis)
  - Stockage persistant : MongoDB
  - API : gRPC
  - Langage : Go

## Sécurité

- Les services MongoDB sont accessibles uniquement via le réseau Docker
- Les interfaces d'administration MongoDB (Mongo Express) sont protégées par authentification
- La communication entre les services se fait via gRPC sur le réseau privé Docker

## Monitoring

Les services génèrent des logs détaillés qui peuvent être consultés via :
```bash
docker logs adserver
docker logs impression_tracker_app
```

## Maintenance

Pour arrêter les services :
```bash
cd adserver/docker && docker-compose down
cd impression-tracker/docker && docker-compose down
```

Pour nettoyer les données :
```bash
docker volume rm adserver-mongodb-data impression_tracker-mongodb-data impression_tracker-dragonfly-data
``` 