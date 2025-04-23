# Microservice de Diffusion de Publicités

Ce projet implémente un microservice de diffusion de publicités avec suivi d'impressions utilisant gRPC, MongoDB et Redis.

## Prérequis

- Go 1.21 ou supérieur
- MongoDB
- Redis
- protoc (protobuf compiler)

## Installation

1. Cloner le dépôt :
```bash
git clone https://github.com/yourusername/ad-service.git
cd ad-service
```

2. Installer les dépendances :
```bash
go mod download
```

3. Générer le code gRPC :
```bash
chmod +x generate.sh
./generate.sh
```

## Configuration

Les services sont configurés pour se connecter à :
- MongoDB sur localhost:27017
- Redis sur localhost:6379

## Exécution

1. Démarrer MongoDB et Redis

2. Lancer le service de publicité :
```bash
cd ad-service
go run main.go
```

## API gRPC

Le service expose les endpoints suivants :

### CreateAd
Crée une nouvelle publicité
```protobuf
rpc CreateAd(CreateAdRequest) returns (CreateAdResponse)
```

### GetAd
Récupère une publicité par son ID
```protobuf
rpc GetAd(GetAdRequest) returns (GetAdResponse)
```

### ServeAd
Diffuse une publicité avec suivi d'impression
```protobuf
rpc ServeAd(ServeAdRequest) returns (ServeAdResponse)
```

## Fonctionnalités

- Création de publicités avec durée de vie
- Suivi des impressions
- Rate limiting
- Validation des données
- Cache avec Redis

## Points bonus implémentés

- [x] Durée de vie des publicités
- [x] Rate limiting
- [x] Validation des données
- [ ] Multi-langage (service de suivi) 