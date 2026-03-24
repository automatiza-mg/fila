-- +goose Up
ALTER TABLE "processos" ADD COLUMN "resumo" TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE "processos" DROP COLUMN "resumo";
