# Service Discover

Serviço de descoberta de serviços.

## Desenvolvimento

```bash
# Instalar dependências
go mod download

# Rodar em desenvolvimento
go run ./cmd

# Rodar testes
go test ./...

# Build
go build -o main ./cmd
```

## Docker

```bash
docker build -t video-ia-service-discover .
docker run -p 8080:8080 video-ia-service-discover
```

## Endpoints

- `GET /` - Mensagem de boas-vindas
- `GET /health` - Health check
