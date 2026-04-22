-- +goose Up
-- +goose StatementBegin
CREATE TYPE "status_solicitacao_diligencia" AS ENUM (
    'rascunho',
    'enviada'
);

CREATE TABLE "solicitacoes_diligencia" (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "processo_aposentadoria_id" BIGINT NOT NULL REFERENCES "processos_aposentadoria"("id") ON DELETE CASCADE,
    "analista_id" BIGINT NOT NULL REFERENCES "analistas"("usuario_id") ON DELETE RESTRICT,
    "status" status_solicitacao_diligencia NOT NULL DEFAULT 'rascunho',
    "criado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "enviada_em" TIMESTAMPTZ
);
CREATE INDEX "solicitacoes_diligencia_processo_idx" ON "solicitacoes_diligencia"("processo_aposentadoria_id");
CREATE UNIQUE INDEX "solicitacoes_diligencia_rascunho_unico_idx"
    ON "solicitacoes_diligencia" ("processo_aposentadoria_id", "analista_id")
    WHERE "status" = 'rascunho';

CREATE TABLE "itens_diligencia" (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "solicitacao_diligencia_id" BIGINT NOT NULL REFERENCES "solicitacoes_diligencia"("id") ON DELETE CASCADE,
    "tipo" TEXT NOT NULL,
    "subcategorias" TEXT[] NOT NULL DEFAULT '{}',
    "detalhe" TEXT NOT NULL DEFAULT ''
);
CREATE INDEX "itens_diligencia_solicitacao_idx" ON "itens_diligencia"("solicitacao_diligencia_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "itens_diligencia";
DROP TABLE "solicitacoes_diligencia";
DROP TYPE "status_solicitacao_diligencia";
-- +goose StatementEnd
