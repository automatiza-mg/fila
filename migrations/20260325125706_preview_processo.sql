-- +goose Up
ALTER TABLE "processos" ADD COLUMN "preview_hash" TEXT REFERENCES "arquivos"("hash");

-- +goose Down
ALTER TABLE "processos" DROP COLUMN "preview_hash";
