-- +goose Up
-- +goose StatementBegin
CREATE TABLE "processos" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "numero" TEXT NOT NULL UNIQUE,
    "status_processamento" TEXT NOT NULL DEFAULT 'PENDENTE',
    "link_acesso" TEXT NOT NULL,
    "sei_unidade_id" TEXT NOT NULL,
    "sei_unidade_sigla" TEXT NOT NULL,
    "metadados_ia" JSONB NOT NULL DEFAULT '{}'::jsonb,
    "aposentadoria" BOOLEAN,
    "analisado_em" TIMESTAMPTZ,
    "criado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "atualizado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "documentos" (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "numero" TEXT NOT NULL,
    "processo_id" UUID NOT NULL REFERENCES "processos"("id") ON DELETE CASCADE,
    "tipo" TEXT NOT NULL,
    "unidade" TEXT NOT NULL,
    "link_acesso" TEXT NOT NULL,
    "content_type" TEXT NOT NULL,
    "chave_storage" TEXT NOT NULL,
    "ocr" TEXT NOT NULL,
    "metadados_api" JSONB NOT NULL DEFAULT '{}'::jsonb,
    "criado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "atualizado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "documentos";
DROP TABLE "processos";
-- +goose StatementEnd
