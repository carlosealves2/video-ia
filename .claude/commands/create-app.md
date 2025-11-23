# Create App

Crie um novo app no monorepo seguindo as instruções abaixo.

## Passo 1: Coletar informações

Use a ferramenta AskUserQuestion para perguntar:

1. **Nome do app** - Nome em kebab-case (ex: api, worker, gateway)
2. **Linguagem** - Node.js, Go ou Python
3. **Frameworks** - Quais frameworks/libs instalar (ex: express, prisma, gin, fastapi)

## Passo 2: Criar estrutura do app

Crie os arquivos em `packages/{app-name}/` usando a estrutura específica da linguagem:

### Node.js

```
packages/{app-name}/
├── src/
│   └── index.ts
├── tests/
│   └── index.spec.ts
├── Dockerfile
├── sonar-project.properties
├── README.md
├── .env.example
├── package.json
├── tsconfig.json
├── .eslintrc.json
└── project.json
```

**package.json:**
```json
{
  "name": "@video-ia/{app-name}",
  "version": "0.0.1",
  "main": "dist/index.js",
  "scripts": {
    "build": "tsc",
    "start": "node dist/index.js",
    "dev": "ts-node src/index.ts",
    "test": "jest",
    "lint": "eslint src --ext .ts"
  },
  "dependencies": {
    // adicionar frameworks solicitados
  },
  "devDependencies": {
    "@types/node": "^22.0.0",
    "typescript": "^5.9.0",
    "ts-node": "^10.9.0",
    "jest": "^29.0.0",
    "@types/jest": "^29.0.0",
    "ts-jest": "^29.0.0",
    "eslint": "^8.0.0",
    "@typescript-eslint/eslint-plugin": "^7.0.0",
    "@typescript-eslint/parser": "^7.0.0"
  }
}
```

**tsconfig.json:**
```json
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "tests"]
}
```

**Dockerfile:**
```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY package*.json ./
ENV NODE_ENV=production
EXPOSE 3000
CMD ["node", "dist/index.js"]
```

### Go

```
packages/{app-name}/
├── cmd/
│   └── main.go
├── internal/
│   └── .gitkeep
├── pkg/
│   └── .gitkeep
├── tests/
│   └── main_test.go
├── Dockerfile
├── sonar-project.properties
├── README.md
├── .env.example
├── go.mod
├── .golangci.yml
└── project.json
```

**go.mod:**
```go
module github.com/carlosealves2/video-ia/{app-name}

go 1.23

require (
    // adicionar frameworks solicitados
)
```

**cmd/main.go:**
```go
package main

import "fmt"

func main() {
    fmt.Println("{app-name} started")
}
```

**Dockerfile:**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

**.golangci.yml:**
```yaml
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - ineffassign
    - unused
```

### Python

```
packages/{app-name}/
├── src/
│   └── {app_name_snake}/
│       ├── __init__.py
│       └── main.py
├── tests/
│   └── test_main.py
├── Dockerfile
├── sonar-project.properties
├── README.md
├── .env.example
├── pyproject.toml
├── ruff.toml
└── project.json
```

**pyproject.toml:**
```toml
[project]
name = "{app-name}"
version = "0.0.1"
description = "{app-name} service"
requires-python = ">=3.13"
dependencies = [
    # adicionar frameworks solicitados
]

[project.optional-dependencies]
dev = [
    "pytest>=8.0",
    "ruff>=0.8",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

**src/{app_name_snake}/main.py:**
```python
def main():
    print("{app-name} started")

if __name__ == "__main__":
    main()
```

**Dockerfile:**
```dockerfile
FROM python:3.13-slim AS builder
WORKDIR /app
COPY pyproject.toml ./
RUN pip install --no-cache-dir .

FROM python:3.13-slim
WORKDIR /app
COPY --from=builder /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY src/ ./src/
ENV PYTHONPATH=/app/src
EXPOSE 8000
CMD ["python", "-m", "{app_name_snake}.main"]
```

**ruff.toml:**
```toml
line-length = 100
target-version = "py313"

[lint]
select = ["E", "F", "I", "N", "W"]
```

## Passo 3: Criar sonar-project.properties

Para todas as linguagens, criar `sonar-project.properties`:

```properties
sonar.projectKey=video-ia-{app-name}
sonar.projectName=Video IA - {App Name Title Case}
sonar.sources=src
sonar.tests=tests
sonar.exclusions=**/node_modules/**,**/dist/**,**/__pycache__/**
```

Ajustar paths conforme a linguagem:
- **Node.js**: `sonar.sources=src`, `sonar.typescript.lcov.reportPaths=coverage/lcov.info`
- **Go**: `sonar.sources=cmd,internal,pkg`, `sonar.go.coverage.reportPaths=coverage.out`
- **Python**: `sonar.sources=src`, `sonar.python.coverage.reportPaths=coverage.xml`

## Passo 4: Criar project.json (NX config)

```json
{
  "name": "{app-name}",
  "$schema": "../../node_modules/nx/schemas/project-schema.json",
  "projectType": "application",
  "sourceRoot": "packages/{app-name}/src",
  "targets": {
    "build": {
      // configurar conforme linguagem
    },
    "test": {
      // configurar conforme linguagem
    },
    "lint": {
      // configurar conforme linguagem
    },
    "serve": {
      // configurar conforme linguagem
    }
  }
}
```

## Passo 5: Criar README.md

```markdown
# {App Name}

Descrição do app.

## Desenvolvimento

```bash
# Instalar dependências
npm install  # ou go mod download / pip install -e ".[dev]"

# Rodar em desenvolvimento
npm run dev  # ou go run ./cmd / python -m {app_name}.main

# Rodar testes
npm test  # ou go test ./... / pytest

# Build
npm run build  # ou go build ./cmd / pip install .
```

## Docker

```bash
docker build -t video-ia-{app-name} .
docker run -p 3000:3000 video-ia-{app-name}
```
```

## Passo 6: Criar .env.example

```
# Environment variables for {app-name}
NODE_ENV=development
PORT=3000
```

## Passo 7: Atualizar Release Please

### release-please-config.json

Adicionar ao objeto `packages`:

```json
"packages/{app-name}": {
  "release-type": "node",  // ou "go" ou "python"
  "component": "{app-name}",
  "changelog-path": "CHANGELOG.md"
}
```

### .release-please-manifest.json

Adicionar:

```json
"packages/{app-name}": "0.0.1"
```

## Passo 8: Finalização

1. Informar ao usuário que o app foi criado
2. Listar os próximos passos:
   - Instalar dependências
   - Configurar variáveis de ambiente
   - Criar projeto no SonarQube com a projectKey
   - Começar a desenvolver
