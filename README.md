# Fila Aposentadoria

Sistema de gestao da fila de aposentadoria do Estado de Minas Gerais. API REST para triagem, analise automatizada (IA) e distribuicao de processos de aposentadoria recebidos via SEI.

## Arquitetura

O sistema recebe processos do SEI (Sistema Eletronico de Informacoes), extrai texto dos documentos via Azure Document Intelligence, analisa com Azure OpenAI (GPT) para classificar e pontuar os pedidos, e distribui para analistas em uma fila priorizada.

Componentes principais:

- API HTTP (chi/v5) na porta 4000
- PostgreSQL 17 para persistencia
- Redis para cache
- River para fila de tarefas assincronas (analise de documentos)
- Integracoes: SEI (SOAP), DataLake Prodemge (Impala), Azure AI

## Estrutura do Projeto

```bash
cmd/
  api/            Servidor HTTP: handlers, rotas, middleware
  cli/            Ferramenta CLI (criacao de admin)
internal/
  auth/           Autenticacao e gestao de usuarios
  fila/           Dominio principal: fila de aposentadoria e analistas
  processos/      Gestao de processos SEI e pipeline de analise
  aposentadoria/  Consultas ao datalake e schema de analise IA
  database/       Camada de acesso a dados (Store pattern)
  cache/          Interface de cache (Redis + in-memory)
  blob/           Armazenamento de arquivos (filesystem/Azure)
  sei/            Cliente SOAP do SEI
  datalake/       Cliente do DataLake Prodemge
  docintel/       Cliente Azure Document Intelligence
  llm/            Cliente Azure OpenAI
  tasks/          Definicao de jobs assincronos (River)
  mail/           Envio de emails (SMTP)
  config/         Configuracao via variaveis de ambiente
  logging/        Factory de logger (slog)
  pagination/     Helpers de paginacao HTTP
  validator/      Validacao de campos
  postgres/       Pool de conexoes e utilitarios de teste
  infra/          Inicializacao do Redis
  soap/           Structs genericas SOAP
migrations/       Migracoes SQL (goose)
```

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

## Comandos Disponiveis

| Comando | Descricao |
|---------|-----------|
| `task migrate:up` | Aplicar migracoes |
| `task migrate:down` | Reverter ultima migracao |
| `task server:build` | Compilar binario |
| `task server:run` | Compilar e executar |
| `task server:watch` | Modo watch com hot reload |

## Testes

```sh
go test ./... -cover
```

Testes de integracao utilizam `dockertest` para subir containers PostgreSQL temporarios. Docker deve estar rodando.
