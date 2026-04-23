-- +goose Up
ALTER TABLE "processos_aposentadoria" ADD COLUMN "alertas" TEXT[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE "processos_aposentadoria" DROP COLUMN "alertas";
