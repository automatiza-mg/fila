-- +goose Up
-- +goose StatementBegin
CREATE TYPE "status_processo" AS ENUM (
    'ANALISE_PENDENTE', 
    'EM_ANALISE', 
    'EM_DILIGENCIA', 
    'RETORNO_DILIGENCIA', 
    'CONCLUIDO', 
    'LEITURA_INVALIDA'
);

CREATE TABLE "processos_aposentadoria" (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "processo_id" UUID NOT NULL UNIQUE REFERENCES "processos"("id") ON DELETE CASCADE,
    "data_requerimento" DATE NOT NULL, -- DataLake (SEI)
    "cpf_requerente" TEXT NOT NULL, -- IA
    "data_nascimento_requerente" DATE NOT NULL, -- IA / DataLake (SISAP)
    "invalidez" BOOLEAN NOT NULL DEFAULT FALSE, -- IA / DataLake (SISAP)
    "judicial" BOOLEAN NOT NULL DEFAULT FALSE, -- IA
    "prioridade" BOOLEAN NOT NULL DEFAULT FALSE, -- Sistema
    "score" INT, -- Sistema
    "status" status_processo NOT NULL DEFAULT 'ANALISE_PENDENTE',
    "analista_id" BIGINT REFERENCES "analistas"("usuario_id") ON DELETE SET NULL,
    "ultimo_analista_id" BIGINT REFERENCES "analistas"("usuario_id") ON DELETE SET NULL, -- Memória do último usuário para qual o processo foi atribuído.
    "criado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "atualizado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX "processos_um_por_analista_idx" ON "processos_aposentadoria"("analista_id") WHERE "status" = 'EM_ANALISE';

CREATE TABLE "historico_status_processo" (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "processo_aposentadoria_id" BIGINT NOT NULL REFERENCES "processos_aposentadoria"("id") ON DELETE CASCADE,
    "status_anterior" status_processo,
    "status_novo" status_processo NOT NULL,
    "usuario_id" BIGINT REFERENCES "usuarios"("id") ON DELETE SET NULL,
    "observacao" TEXT,
    "alterado_em" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "historico_status_processo";
DROP TABLE "processos_aposentadoria";
DROP TYPE "status_processo";
-- +goose StatementEnd
