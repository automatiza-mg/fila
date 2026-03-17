-- +goose Up
CREATE TABLE "solicitacoes_prioridade" (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "processo_aposentadoria_id" BIGINT NOT NULL REFERENCES "processos_aposentadoria"("id") ON DELETE CASCADE,
    "justificativa" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'pendente', -- pendente, aprovado, negado
    "usuario_id" BIGINT NOT NULL REFERENCES "usuarios"("id") ON DELETE CASCADE,
    "criado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "atualizado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE "solicitacoes_prioridade";
