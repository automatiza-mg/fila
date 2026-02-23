# Fila Aposentadoria

Sistema de gestao da fila de aposentadoria do Estado de Minas Gerais. API REST para triagem, analise automatizada (IA) e distribuicao de processos de aposentadoria recebidos via SEI.

## Requisitos

1. [Go 1.25](https://go.dev/dl/)
2. [Task](https://taskfile.dev/docs/installation)
3. [Docker](https://www.docker.com)
4. [Goose](https://pressly.github.io/goose/installation)
5. [sqlc](https://docs.sqlc.dev/en/stable/overview/install.html)
6. [PostgreSQL 17](https://www.postgresql.org/download/)
7. [Redis](https://redis.io/docs/getting-started/)

## Configuracao

Copie o arquivo de exemplo e preencha as variaveis:

```sh
cp .env.example .env
```

Consulte `.env.example` para a lista completa de variaveis.

## Executando

```sh
# Aplicar migracoes
task migrate:up

# Rodar em modo desenvolvimento (com hot reload)
task server:watch

# Ou compilar e rodar manualmente
task server:build
./bin/api -dev

# Criar usuario admin
go run ./cmd/cli create-admin --name "Nome" --cpf "00000000000" --email "email@example.com"
```

## Testes

```sh
go test ./... -cover
```

Testes de integracao utilizam `dockertest` para subir containers PostgreSQL temporarios. Docker deve estar rodando.
