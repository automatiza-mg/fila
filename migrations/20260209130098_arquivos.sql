-- +goose Up
CREATE TABLE "arquivos" (
    "hash" TEXT PRIMARY KEY,
    "chave_storage" TEXT NOT NULL,
    "content_type" TEXT NOT NULL,
    "conteudo" TEXT NOT NULL,
    "formato_conteudo" TEXT NOT NULL DEFAULT 'plain', -- plain/markdown
    "criado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE "arquivos";
