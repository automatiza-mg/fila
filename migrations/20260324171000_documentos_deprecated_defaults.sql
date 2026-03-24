-- +goose Up
ALTER TABLE "documentos" ALTER COLUMN "content_type" SET DEFAULT '';
ALTER TABLE "documentos" ALTER COLUMN "chave_storage" SET DEFAULT '';
ALTER TABLE "documentos" ALTER COLUMN "ocr" SET DEFAULT '';

-- +goose Down
ALTER TABLE "documentos" ALTER COLUMN "content_type" DROP DEFAULT;
ALTER TABLE "documentos" ALTER COLUMN "chave_storage" DROP DEFAULT;
ALTER TABLE "documentos" ALTER COLUMN "ocr" DROP DEFAULT;
