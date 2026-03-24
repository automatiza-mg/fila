-- +goose Up
CREATE TABLE "arquivos" (
    "hash" TEXT PRIMARY KEY,
    "chave_storage" TEXT NOT NULL,
    "ocr" TEXT NOT NULL,
    "content_type" TEXT NOT NULL,
    "criado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE "documentos" ADD COLUMN "arquivo_hash" TEXT REFERENCES "arquivos"("hash");

-- +goose Down
ALTER TABLE "documentos" DROP COLUMN "arquivo_hash";
DROP TABLE "arquivos";
